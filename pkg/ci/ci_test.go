package ci

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/juspay/omnix/pkg/nix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	assert.NotEmpty(t, config.Default)
	assert.Contains(t, config.Default, ".")

	rootConfig := config.Default["."]
	assert.False(t, rootConfig.Skip)
	assert.Equal(t, ".", rootConfig.Dir)
	assert.True(t, rootConfig.Steps.Build.Enable)
	assert.True(t, rootConfig.Steps.Lockfile.Enable)
	assert.True(t, rootConfig.Steps.FlakeCheck.Enable)
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "om.yaml")

	configContent := `
ci:
  default:
    ".":
      dir: "."
      skip: false
      steps:
        build:
          enable: true
          impure: false
        lockfile:
          enable: true
        flakeCheck:
          enable: false
    "tests":
      dir: "tests"
      systems:
        - "x86_64-linux"
      steps:
        build:
          enable: true
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	config, err := LoadConfig(configPath)
	require.NoError(t, err)

	assert.Contains(t, config.Default, ".")
	assert.Contains(t, config.Default, "tests")

	rootConfig := config.Default["."]
	assert.True(t, rootConfig.Steps.Build.Enable)
	assert.True(t, rootConfig.Steps.Lockfile.Enable)
	assert.False(t, rootConfig.Steps.FlakeCheck.Enable)

	testsConfig := config.Default["tests"]
	assert.Equal(t, "tests", testsConfig.Dir)
	assert.Equal(t, []string{"x86_64-linux"}, testsConfig.Systems)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "om.yaml")

	err := os.WriteFile(configPath, []byte("invalid: [yaml"), 0644)
	require.NoError(t, err)

	_, err = LoadConfig(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config YAML")
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/om.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestSubflakeConfig_CanRunOn(t *testing.T) {
	tests := []struct {
		name     string
		config   SubflakeConfig
		systems  []string
		expected bool
	}{
		{
			name: "no whitelist - can run on any system",
			config: SubflakeConfig{
				Systems: []string{},
			},
			systems:  []string{"x86_64-linux"},
			expected: true,
		},
		{
			name: "whitelist matches",
			config: SubflakeConfig{
				Systems: []string{"x86_64-linux", "aarch64-darwin"},
			},
			systems:  []string{"x86_64-linux"},
			expected: true,
		},
		{
			name: "whitelist doesn't match",
			config: SubflakeConfig{
				Systems: []string{"aarch64-darwin"},
			},
			systems:  []string{"x86_64-linux"},
			expected: false,
		},
		{
			name: "multiple systems - one matches",
			config: SubflakeConfig{
				Systems: []string{"aarch64-darwin"},
			},
			systems:  []string{"x86_64-linux", "aarch64-darwin"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.CanRunOn(tt.systems)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStepsConfig_GetEnabledSteps(t *testing.T) {
	tests := []struct {
		name     string
		config   StepsConfig
		expected []string
	}{
		{
			name: "all steps enabled",
			config: StepsConfig{
				Build:      BuildStep{Enable: true},
				Lockfile:   LockfileStep{Enable: true},
				FlakeCheck: FlakeCheckStep{Enable: true},
				Custom:     []CustomStep{},
			},
			expected: []string{"build", "lockfile", "flakeCheck"},
		},
		{
			name: "only build enabled",
			config: StepsConfig{
				Build:      BuildStep{Enable: true},
				Lockfile:   LockfileStep{Enable: false},
				FlakeCheck: FlakeCheckStep{Enable: false},
				Custom:     []CustomStep{},
			},
			expected: []string{"build"},
		},
		{
			name: "with custom steps",
			config: StepsConfig{
				Build:    BuildStep{Enable: true},
				Lockfile: LockfileStep{Enable: false},
				FlakeCheck: FlakeCheckStep{Enable: false},
				Custom: []CustomStep{
					{Name: "test", Enable: true},
					{Name: "lint", Enable: false},
				},
			},
			expected: []string{"build", "custom:test"},
		},
		{
			name: "no steps enabled",
			config: StepsConfig{
				Build:      BuildStep{Enable: false},
				Lockfile:   LockfileStep{Enable: false},
				FlakeCheck: FlakeCheckStep{Enable: false},
				Custom:     []CustomStep{},
			},
			expected: nil, // Changed from []string{} to nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetEnabledSteps()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateMatrix(t *testing.T) {
	tests := []struct {
		name           string
		systems        []string
		config         Config
		expectedCount  int
		expectedFirst  *GitHubMatrixRow
	}{
		{
			name:    "single system, single subflake",
			systems: []string{"x86_64-linux"},
			config: Config{
				Default: map[string]SubflakeConfig{
					".": {Dir: ".", Skip: false},
				},
			},
			expectedCount: 1,
			expectedFirst: &GitHubMatrixRow{
				System:   "x86_64-linux",
				Subflake: ".",
			},
		},
		{
			name:    "multiple systems, single subflake",
			systems: []string{"x86_64-linux", "aarch64-darwin"},
			config: Config{
				Default: map[string]SubflakeConfig{
					".": {Dir: ".", Skip: false},
				},
			},
			expectedCount: 2,
		},
		{
			name:    "single system, multiple subflakes",
			systems: []string{"x86_64-linux"},
			config: Config{
				Default: map[string]SubflakeConfig{
					".":     {Dir: ".", Skip: false},
					"tests": {Dir: "tests", Skip: false},
				},
			},
			expectedCount: 2,
		},
		{
			name:    "skip subflake",
			systems: []string{"x86_64-linux"},
			config: Config{
				Default: map[string]SubflakeConfig{
					".":     {Dir: ".", Skip: false},
					"tests": {Dir: "tests", Skip: true},
				},
			},
			expectedCount: 1,
		},
		{
			name:    "system whitelist",
			systems: []string{"x86_64-linux", "aarch64-darwin"},
			config: Config{
				Default: map[string]SubflakeConfig{
					".": {
						Dir:     ".",
						Skip:    false,
						Systems: []string{"x86_64-linux"},
					},
				},
			},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matrix := GenerateMatrix(tt.systems, tt.config)
			assert.Equal(t, tt.expectedCount, matrix.Count())

			if tt.expectedFirst != nil && len(matrix.Include) > 0 {
				assert.Equal(t, tt.expectedFirst.System, matrix.Include[0].System)
				assert.Equal(t, tt.expectedFirst.Subflake, matrix.Include[0].Subflake)
			}
		})
	}
}

func TestGitHubMatrix_ToJSON(t *testing.T) {
	matrix := GitHubMatrix{
		Include: []GitHubMatrixRow{
			{System: "x86_64-linux", Subflake: "."},
			{System: "aarch64-darwin", Subflake: "tests"},
		},
	}

	json, err := matrix.ToJSON()
	require.NoError(t, err)
	assert.Contains(t, json, "x86_64-linux")
	assert.Contains(t, json, "aarch64-darwin")
	assert.Contains(t, json, "include")
}

func TestRunSubflake_DisabledSteps(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	flake, err := nix.ParseFlakeURL(".")
	require.NoError(t, err)

	subflake := SubflakeConfig{
		Dir:  ".",
		Skip: false,
		Steps: StepsConfig{
			Build:      BuildStep{Enable: false},
			Lockfile:   LockfileStep{Enable: false},
			FlakeCheck: FlakeCheckStep{Enable: false},
			Custom:     []CustomStep{},
		},
	}

	opts := RunOptions{
		Systems:      []string{"x86_64-linux"},
		GitHubOutput: false,
	}

	result, err := runSubflake(ctx, flake, ".", subflake, opts)
	require.NoError(t, err)
	assert.Equal(t, ".", result.Subflake)
	assert.Empty(t, result.Steps)
	assert.True(t, result.Success)
}

func TestStepResult(t *testing.T) {
	result := StepResult{
		Name:    "test",
		Success: true,
		Output:  "test output",
	}

	assert.Equal(t, "test", result.Name)
	assert.True(t, result.Success)
	assert.Equal(t, "test output", result.Output)
	assert.Empty(t, result.Error)
}
