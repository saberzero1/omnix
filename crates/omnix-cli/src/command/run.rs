use anyhow::{Context, Result};
use clap::Parser;
use colored::Colorize;
use nix_rs::{
    command::NixCmd,
    flake::{system::System, url::FlakeUrl},
    info::NixInfo,
};
use omnix_ci::{
    command::run::{check_nix_version, ci_run, RunCommand as CiRunCommand},
    flake_ref::FlakeRef,
    github::actions::in_github_log_group,
};
use omnix_common::config::{OmConfig, OmConfigTree};
use serde::Deserialize;
use std::{collections::BTreeMap, env, io::Write, path::PathBuf};

/// Run tasks from om/ directory
#[derive(Parser, Debug, Clone)]
pub struct RunCommand {
    /// Task name to run (defaults to "default")
    ///
    /// This will run the file om/{name}.yaml
    #[arg(default_value = "default")]
    pub name: String,

    /// The systems list to build for. If empty, build for current system.
    ///
    /// Must be a flake reference which, when imported, must return a Nix list
    /// of systems. You may use one of the lists from
    /// <https://github.com/nix-systems>.
    ///
    /// You can also pass the individual system name, if they are supported by omnix.
    #[arg(long)]
    pub systems: Option<nix_rs::system_list::SystemsListFlakeRef>,

    /// Symlink to build results (as JSON)
    #[arg(
        long,
        short = 'o',
        default_value = "result",
        conflicts_with = "no_link",
        name = "PATH"
    )]
    out_link: Option<PathBuf>,

    /// Do not create a symlink to build results JSON
    #[arg(long)]
    no_link: bool,

    /// Flake URL or github URL
    #[arg(default_value = ".")]
    pub flake_ref: FlakeRef,

    /// Print Github Actions log groups (enabled by default when run in Github Actions)
    #[clap(long, default_value_t = env::var("GITHUB_ACTION").is_ok())]
    pub github_output: bool,

    /// Nix command global options
    #[command(flatten)]
    pub nixcmd: NixCmd,
}

/// Simplified config format for om run
#[derive(Debug, Deserialize, Clone)]
struct RunConfig {
    /// Subdirectory in which the flake lives
    #[serde(default = "default_dir")]
    dir: String,

    /// List of CI steps to run
    #[serde(default)]
    steps: serde_json::Value,

    /// Cache configuration
    #[serde(default)]
    caches: Option<CachesConfig>,

    /// Inputs to override (via --override-input)
    #[serde(rename = "overrideInputs", default)]
    override_inputs: BTreeMap<String, FlakeUrl>,

    /// An optional whitelist of systems to build on (others are ignored)
    #[serde(default)]
    systems: Option<Vec<System>>,
}

fn default_dir() -> String {
    ".".to_string()
}

/// Cache configuration
#[derive(Debug, Deserialize, Clone)]
struct CachesConfig {
    /// Required caches
    #[serde(default)]
    required: Vec<String>,
}

impl RunCommand {
    /// Run the command
    pub async fn run(&self) -> Result<()> {
        tracing::info!("{}", "\n👟 Reading run config from om/ directory".bold());

        let url = self.flake_ref.to_flake_url().await?;
        let config_path = self.get_config_path(&url).await?;

        // Load the simplified config
        let yaml_str = std::fs::read_to_string(&config_path)
            .with_context(|| format!("Failed to read config from {:?}", config_path))?;
        let run_config: RunConfig = serde_yaml::from_str(&yaml_str)
            .with_context(|| format!("Failed to parse config from {:?}", config_path))?;

        // Convert to the format expected by ci_run
        let om_config = self.convert_to_om_config(&url, run_config)?;

        // Create a CiRunCommand with appropriate defaults
        let ci_cmd = self.to_ci_run_command();

        // Run the CI pipeline
        let nix_info = in_github_log_group("info", self.github_output, || async {
            tracing::info!("{}", "\n👟 Gathering NixInfo".bold());
            NixInfo::get()
                .await
                .as_ref()
                .with_context(|| "Unable to gather nix info")
        })
        .await?;

        // Health check
        in_github_log_group("health", self.github_output, || async {
            tracing::info!("{}", "\n🫀 Performing health check".bold());
            check_nix_version(&om_config, nix_info).await
        })
        .await?;

        // Run CI steps
        tracing::info!(
            "{}",
            format!("\n🤖 Running task '{}' for {}", self.name, self.flake_ref).bold()
        );
        let res = ci_run(&self.nixcmd, &ci_cmd, &om_config, &nix_info.nix_config).await?;

        let msg =
            in_github_log_group::<Result<String>, _, _>("outlink", self.github_output, || async {
                let m_out_link = ci_cmd.get_out_link();
                let s = serde_json::to_string(&res)?;
                let mut path = tempfile::Builder::new()
                    .prefix("om-run-results-")
                    .suffix(".json")
                    .tempfile()?;
                path.write_all(s.as_bytes())?;

                let results_path = nix_rs::flake::functions::addstringcontext::addstringcontext(
                    &self.nixcmd,
                    path.path(),
                    m_out_link,
                )
                .await?;
                println!("{}", results_path.display());

                let msg = format!(
                    "Result available at {:?}{}",
                    results_path.as_path(),
                    m_out_link
                        .map(|p| format!(" and symlinked at {:?}", p))
                        .unwrap_or_default()
                );
                Ok(msg)
            })
            .await?;

        tracing::info!("{}", msg);

        Ok(())
    }

    /// Get the path to the config file
    async fn get_config_path(&self, url: &FlakeUrl) -> Result<PathBuf> {
        let base_path = if let Some(local_path) = url.without_attr().as_local_path() {
            local_path.to_path_buf()
        } else {
            url.without_attr()
                .as_local_path_or_fetch(&self.nixcmd)
                .await?
        };

        let config_path = base_path.join("om").join(format!("{}.yaml", self.name));

        if !config_path.exists() {
            anyhow::bail!(
                "Config file not found: {:?}\nExpected om/{}.yaml to exist",
                config_path,
                self.name
            );
        }

        Ok(config_path)
    }

    /// Convert RunConfig to OmConfig format expected by ci_run
    fn convert_to_om_config(&self, url: &FlakeUrl, run_config: RunConfig) -> Result<OmConfig> {
        // Build the config tree using raw JSON values to avoid serialization issues
        let ci_config_value = serde_json::json!({
            "default": {
                "ROOT": {
                    "dir": run_config.dir,
                    "skip": false,
                    "overrideInputs": run_config.override_inputs,
                    "systems": run_config.systems,
                    "steps": {
                        "lockfile": {
                            "enable": false
                        },
                        "build": {
                            "enable": false
                        },
                        "flake-check": {
                            "enable": false
                        },
                        "custom": run_config.steps
                    }
                }
            }
        });

        let mut config_tree_map = serde_json::Map::new();
        config_tree_map.insert("ci".to_string(), ci_config_value);

        // Add caches to health config if present
        if let Some(caches) = run_config.caches {
            let health_config = serde_json::json!({
                "default": {
                    "caches": {
                        "required": caches.required
                    }
                }
            });
            config_tree_map.insert("health".to_string(), health_config);
        }

        let config_tree: OmConfigTree =
            serde_json::from_value(serde_json::Value::Object(config_tree_map))?;

        Ok(OmConfig {
            flake_url: url.without_attr(),
            reference: vec![],
            config: config_tree,
        })
    }

    /// Convert to CiRunCommand with appropriate defaults
    fn to_ci_run_command(&self) -> CiRunCommand {
        let out_link = if self.no_link {
            None
        } else {
            self.out_link.clone()
        };

        CiRunCommand::default().local_with(self.flake_ref.clone(), out_link)
    }
}
