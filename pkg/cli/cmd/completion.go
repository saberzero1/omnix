package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// NewCompletionCmd creates the completion command
func NewCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for omnix.

To load completions:

Bash:
  $ source <(om completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ om completion bash > /etc/bash_completion.d/om
  # macOS:
  $ om completion bash > $(brew --prefix)/etc/bash_completion.d/om

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ om completion zsh > "${fpath[1]}/_om"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ om completion fish | source

  # To load completions for each session, execute once:
  $ om completion fish > ~/.config/fish/completions/om.fish

PowerShell:
  PS> om completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> om completion powershell > om.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		RunE:                  runCompletion,
	}

	return cmd
}

func runCompletion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	}
	return nil
}
