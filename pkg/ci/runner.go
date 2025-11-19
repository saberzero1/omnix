package ci

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/saberzero1/omnix/pkg/nix"
	"go.uber.org/zap"
)

// RunOptions contains options for running CI
type RunOptions struct {
	// Systems to build for
	Systems []string

	// GitHubOutput controls whether to print GitHub Actions log groups
	GitHubOutput bool

	// IncludeAllDependencies includes all dependencies in results
	IncludeAllDependencies bool

	// RemoteHost specifies a remote host for SSH-based builds (e.g., "user@host")
	RemoteHost string

	// Parallel controls whether to run steps in parallel
	Parallel bool

	// MaxConcurrency limits the number of parallel steps (0 = unlimited)
	MaxConcurrency int
}

// Result represents the result of a CI run
type Result struct {
	// Subflake is the name of the subflake
	Subflake string `json:"subflake"`

	// Steps contains results for each step
	Steps map[string]StepResult `json:"steps"`

	// Duration is how long the CI run took
	Duration time.Duration `json:"duration"`

	// Success indicates if all steps passed
	Success bool `json:"success"`
}

// StepResult represents the result of a single CI step
type StepResult struct {
	// Name of the step
	Name string `json:"name"`

	// Success indicates if the step passed
	Success bool `json:"success"`

	// Error contains error message if step failed
	Error string `json:"error,omitempty"`

	// Output contains step output
	Output string `json:"output,omitempty"`

	// Duration is how long the step took
	Duration time.Duration `json:"duration"`
}

// Run executes the CI pipeline for a flake
func Run(ctx context.Context, flake nix.FlakeURL, config Config, opts RunOptions) ([]Result, error) {
	// Collect subflakes to run
	var subflakes []struct {
		name   string
		config SubflakeConfig
	}

	for name, subflake := range config.Default {
		// Skip if marked to skip
		if subflake.Skip {
			continue
		}

		// Skip if can't run on requested systems
		if !subflake.CanRunOn(opts.Systems) {
			continue
		}

		subflakes = append(subflakes, struct {
			name   string
			config SubflakeConfig
		}{name, subflake})
	}

	// Run sequentially or in parallel based on opts
	if opts.Parallel {
		return runSubflakesParallel(ctx, flake, subflakes, opts)
	}

	return runSubflakesSequential(ctx, flake, subflakes, opts)
}

// runSubflakesSequential runs subflakes one after another
func runSubflakesSequential(ctx context.Context, flake nix.FlakeURL, subflakes []struct {
	name   string
	config SubflakeConfig
}, opts RunOptions) ([]Result, error) {
	var results []Result

	for _, sf := range subflakes {
		result, err := runSubflake(ctx, flake, sf.name, sf.config, opts)
		if err != nil {
			return results, fmt.Errorf("failed to run subflake %s: %w", sf.name, err)
		}

		results = append(results, result)
	}

	return results, nil
}

// runSubflakesParallel runs subflakes in parallel
func runSubflakesParallel(ctx context.Context, flake nix.FlakeURL, subflakes []struct {
	name   string
	config SubflakeConfig
}, opts RunOptions) ([]Result, error) {
	// Determine concurrency limit
	maxConcurrency := opts.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = len(subflakes)
	}

	// Create channels for work distribution
	type job struct {
		index  int
		name   string
		config SubflakeConfig
	}

	type jobResult struct {
		index  int
		result Result
		err    error
	}

	jobs := make(chan job, len(subflakes))
	jobResults := make(chan jobResult, len(subflakes))

	// Start worker goroutines
	for w := 0; w < maxConcurrency; w++ {
		go func() {
			for j := range jobs {
				result, err := runSubflake(ctx, flake, j.name, j.config, opts)
				jobResults <- jobResult{
					index:  j.index,
					result: result,
					err:    err,
				}
			}
		}()
	}

	// Queue all jobs
	for i, sf := range subflakes {
		jobs <- job{
			index:  i,
			name:   sf.name,
			config: sf.config,
		}
	}
	close(jobs)

	// Collect results
	resultsMap := make(map[int]Result)
	var firstError error

	for i := 0; i < len(subflakes); i++ {
		jr := <-jobResults
		if jr.err != nil && firstError == nil {
			firstError = jr.err
		}
		resultsMap[jr.index] = jr.result
	}

	// Return error if any occurred
	if firstError != nil {
		return nil, firstError
	}

	// Sort results by original order
	results := make([]Result, len(subflakes))
	for i := 0; i < len(subflakes); i++ {
		results[i] = resultsMap[i]
	}

	return results, nil
}

// runSubflake runs CI for a single subflake
func runSubflake(ctx context.Context, flake nix.FlakeURL, name string, subflake SubflakeConfig, opts RunOptions) (Result, error) {
	start := time.Now()

	result := Result{
		Subflake: name,
		Steps:    make(map[string]StepResult),
		Success:  true,
	}

	// Get the subflake URL
	subflakeURL := flake
	if subflake.Dir != "." {
		urlStr := flake.String() + "#" + subflake.Dir
		var err error
		subflakeURL, err = nix.ParseFlakeURL(urlStr)
		if err != nil {
			return result, fmt.Errorf("failed to parse subflake URL: %w", err)
		}
	}

	// Run build step
	if subflake.Steps.Build.Enable {
		var stepResult StepResult
		if opts.RemoteHost != "" {
			stepResult = runBuildStepRemote(ctx, opts.RemoteHost, subflakeURL, subflake.Steps.Build, opts)
		} else {
			stepResult = runBuildStep(ctx, subflakeURL, subflake.Steps.Build, opts)
		}
		result.Steps["build"] = stepResult
		if !stepResult.Success {
			result.Success = false
		}
	}

	// Run lockfile step
	if subflake.Steps.Lockfile.Enable {
		var stepResult StepResult
		if opts.RemoteHost != "" {
			stepResult = runLockfileStepRemote(ctx, opts.RemoteHost, subflakeURL, subflake.Steps.Lockfile)
		} else {
			stepResult = runLockfileStep(ctx, subflakeURL, subflake.Steps.Lockfile)
		}
		result.Steps["lockfile"] = stepResult
		if !stepResult.Success {
			result.Success = false
		}
	}

	// Run flake check step
	if subflake.Steps.FlakeCheck.Enable {
		var stepResult StepResult
		if opts.RemoteHost != "" {
			stepResult = runFlakeCheckStepRemote(ctx, opts.RemoteHost, subflakeURL, subflake.Steps.FlakeCheck)
		} else {
			stepResult = runFlakeCheckStep(ctx, subflakeURL, subflake.Steps.FlakeCheck)
		}
		result.Steps["flakeCheck"] = stepResult
		if !stepResult.Success {
			result.Success = false
		}
	}

	// Run custom steps
	for _, customStep := range subflake.Steps.Custom {
		if customStep.Enable {
			var stepResult StepResult
			if opts.RemoteHost != "" {
				stepResult = runCustomStepRemote(ctx, opts.RemoteHost, subflakeURL, customStep)
			} else {
				stepResult = runCustomStep(ctx, subflakeURL, customStep)
			}
			result.Steps["custom:"+customStep.Name] = stepResult
			if !stepResult.Success {
				result.Success = false
			}
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// runBuildStep executes the build step
func runBuildStep(ctx context.Context, flake nix.FlakeURL, step BuildStep, opts RunOptions) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "build",
		Success: true,
	}

	// Build the flake
	args := []string{"build", flake.String(), "--no-link", "--print-out-paths"}
	if step.Impure {
		args = append(args, "--impure")
	}

	cmd := nix.NewCmd()
	output, err := cmd.Run(ctx, args...)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
	}
	result.Output = output
	result.Duration = time.Since(start)

	return result
}

// runLockfileStep executes the lockfile check step
func runLockfileStep(ctx context.Context, flake nix.FlakeURL, step LockfileStep) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "lockfile",
		Success: true,
	}

	// Check if flake.lock is up to date
	cmd := nix.NewCmd()
	output, err := cmd.Run(ctx, "flake", "lock", "--no-update-lock-file", flake.String())
	if err != nil {
		result.Success = false
		result.Error = "flake.lock is out of date"
	}
	result.Output = output
	result.Duration = time.Since(start)

	return result
}

// runFlakeCheckStep executes the flake check step
func runFlakeCheckStep(ctx context.Context, flake nix.FlakeURL, step FlakeCheckStep) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "flakeCheck",
		Success: true,
	}

	// Run nix flake check
	cmd := nix.NewCmd()
	output, err := cmd.Run(ctx, "flake", "check", flake.String())
	if err != nil {
		result.Success = false
		result.Error = err.Error()
	}
	result.Output = output
	result.Duration = time.Since(start)

	return result
}

// runCustomStep executes a custom step
func runCustomStep(ctx context.Context, flake nix.FlakeURL, step CustomStep) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "custom:" + step.Name,
		Success: true,
	}

	if len(step.Command) == 0 {
		result.Success = false
		result.Error = "custom step has no command"
		result.Duration = time.Since(start)
		return result
	}

	// Use nix.Cmd for nix commands, exec.Command for others
	var output string
	var err error

	if step.Command[0] == "nix" {
		cmd := nix.NewCmd()
		output, err = cmd.Run(ctx, step.Command[1:]...)
	} else {
		// For non-nix commands, use exec.Command directly
		execCmd := exec.CommandContext(ctx, step.Command[0], step.Command[1:]...)
		outputBytes, execErr := execCmd.CombinedOutput()
		output = string(outputBytes)
		err = execErr
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	}
	result.Output = output
	result.Duration = time.Since(start)

	return result
}

// LogResult logs the CI result using the logger
func LogResult(result Result, logger *zap.Logger) {
	logger.Info("CI Result",
		zap.String("subflake", result.Subflake),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration))

	for name, stepResult := range result.Steps {
		logger.Info("  Step",
			zap.String("name", name),
			zap.Bool("success", stepResult.Success),
			zap.Duration("duration", stepResult.Duration))

		if !stepResult.Success {
			logger.Error("  Step failed",
				zap.String("name", name),
				zap.String("error", stepResult.Error))
		}
	}
}

// executeRemoteCommand executes a command on a remote host via SSH
func executeRemoteCommand(ctx context.Context, host string, command []string) (string, error) {
	if host == "" {
		return "", fmt.Errorf("remote host not specified")
	}

	// Build SSH command
	// SSH command format: ssh user@host "command args..."
	sshArgs := []string{host}

	// Convert command array to shell command string with proper escaping
	// Use POSIX shell escaping: wrap each argument in single quotes and escape any embedded single quotes
	cmdParts := make([]string, len(command))
	for i, part := range command {
		// Replace single quotes with '\'' (end quote, escaped quote, start quote)
		escaped := strings.ReplaceAll(part, "'", "'\\''")
		cmdParts[i] = "'" + escaped + "'"
	}
	cmdStr := strings.Join(cmdParts, " ")
	sshArgs = append(sshArgs, cmdStr)

	// Execute SSH command
	cmd := exec.CommandContext(ctx, "ssh", sshArgs...)
	output, err := cmd.CombinedOutput()

	return string(output), err
}

// runBuildStepRemote executes the build step on a remote host
func runBuildStepRemote(ctx context.Context, host string, flake nix.FlakeURL, step BuildStep, opts RunOptions) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "build",
		Success: true,
	}

	// Build the command
	args := []string{"nix", "build", flake.String(), "--no-link", "--print-out-paths"}
	if step.Impure {
		args = append(args, "--impure")
	}

	output, err := executeRemoteCommand(ctx, host, args)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("remote build failed: %v", err)
	}
	result.Output = output
	result.Duration = time.Since(start)

	return result
}

// runLockfileStepRemote executes the lockfile check step on a remote host
func runLockfileStepRemote(ctx context.Context, host string, flake nix.FlakeURL, step LockfileStep) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "lockfile",
		Success: true,
	}

	args := []string{"nix", "flake", "lock", "--no-update-lock-file", flake.String()}
	output, err := executeRemoteCommand(ctx, host, args)
	if err != nil {
		result.Success = false
		result.Error = "flake.lock is out of date"
	}
	result.Output = output
	result.Duration = time.Since(start)

	return result
}

// runFlakeCheckStepRemote executes the flake check step on a remote host
func runFlakeCheckStepRemote(ctx context.Context, host string, flake nix.FlakeURL, step FlakeCheckStep) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "flakeCheck",
		Success: true,
	}

	args := []string{"nix", "flake", "check", flake.String()}
	output, err := executeRemoteCommand(ctx, host, args)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("flake check failed: %v", err)
	}
	result.Output = output
	result.Duration = time.Since(start)

	return result
}

// runCustomStepRemote executes a custom step on a remote host
func runCustomStepRemote(ctx context.Context, host string, flake nix.FlakeURL, step CustomStep) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "custom:" + step.Name,
		Success: true,
	}

	output, err := executeRemoteCommand(ctx, host, step.Command)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("custom step failed: %v", err)
	}
	result.Output = output
	result.Duration = time.Since(start)

	return result
}
