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
	for name, customStep := range subflake.Steps.Custom {
		// Check if this step can run on current systems
		if !customStep.CanRunOn(opts.Systems) {
			continue
		}

		var stepResult StepResult
		if opts.RemoteHost != "" {
			stepResult = runCustomStepRemote(ctx, opts.RemoteHost, subflakeURL, name, customStep)
		} else {
			stepResult = runCustomStep(ctx, subflakeURL, name, customStep)
		}
		result.Steps["custom:"+name] = stepResult
		if !stepResult.Success {
			result.Success = false
		}
	}

	result.Duration = time.Since(start)
	return result, nil
}

// runBuildStep executes the build step using devour-flake
func runBuildStep(ctx context.Context, flake nix.FlakeURL, step BuildStep, opts RunOptions) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "build",
		Success: true,
	}

	// Use devour-flake to build all outputs
	output, err := nix.DevourFlake(ctx, flake, opts.Systems, step.Impure)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	// Format output paths as string
	var outPaths []string
	for _, path := range output.OutPaths {
		outPaths = append(outPaths, path.String())
	}
	result.Output = fmt.Sprintf("Built %d outputs:\n%s", len(outPaths), strings.Join(outPaths, "\n"))
	result.Duration = time.Since(start)

	return result
}

// runLockfileStep executes the lockfile check step
func runLockfileStep(ctx context.Context, flake nix.FlakeURL, _ LockfileStep) StepResult {
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
func runFlakeCheckStep(ctx context.Context, flake nix.FlakeURL, _ FlakeCheckStep) StepResult {
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
func runCustomStep(ctx context.Context, flake nix.FlakeURL, name string, step CustomStep) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "custom:" + name,
		Success: true,
	}

	var output string
	var err error

	switch step.Type {
	case CustomStepTypeApp:
		// Run a flake app
		output, err = runFlakeApp(ctx, flake, step)
	case CustomStepTypeDevShell:
		// Run a command in a devshell
		output, err = runDevShellCommand(ctx, flake, step)
	default:
		result.Success = false
		result.Error = fmt.Sprintf("unknown custom step type: %s", step.Type)
		result.Duration = time.Since(start)
		return result
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
	}
	result.Output = output
	result.Duration = time.Since(start)

	return result
}

// runFlakeApp runs a flake app
func runFlakeApp(ctx context.Context, flake nix.FlakeURL, step CustomStep) (string, error) {
	// Determine the app name (default to "default" if not specified)
	appName := "default"
	if step.Name != "" {
		appName = step.Name
	}

	// Build the flake URL with app attribute
	appURL := flake.String() + "#" + appName

	// Build nix run command
	args := []string{"run", appURL}
	if len(step.Args) > 0 {
		args = append(args, "--")
		args = append(args, step.Args...)
	}

	cmd := nix.NewCmd()
	return cmd.Run(ctx, args...)
}

// runDevShellCommand runs a command in a devshell
func runDevShellCommand(ctx context.Context, flake nix.FlakeURL, step CustomStep) (string, error) {
	if len(step.Command) == 0 {
		return "", fmt.Errorf("devshell step has no command")
	}

	// Determine the devshell name (default to "default" if not specified)
	shellName := "default"
	if step.Name != "" {
		shellName = step.Name
	}

	// Build the flake URL with devshell attribute
	shellURL := flake.String()
	if shellName != "default" {
		shellURL = shellURL + "#" + shellName
	}

	// Build nix develop command
	args := []string{"develop", shellURL, "-c"}
	args = append(args, step.Command...)

	cmd := nix.NewCmd()
	return cmd.Run(ctx, args...)
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

// runBuildStepRemote executes the build step on a remote host using devour-flake
func runBuildStepRemote(ctx context.Context, host string, flake nix.FlakeURL, step BuildStep, opts RunOptions) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "build",
		Success: true,
	}

	// Build the devour-flake command
	devourURL := nix.DevourFlakeURL() + "#json"
	args := []string{"nix", "build", devourURL, "-L", "--no-link", "--print-out-paths"}
	if step.Impure {
		args = append(args, "--impure")
	}
	args = append(args, "--override-input", "flake", flake.String())

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
func runCustomStepRemote(ctx context.Context, host string, flake nix.FlakeURL, name string, step CustomStep) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "custom:" + name,
		Success: true,
	}

	var args []string

	switch step.Type {
	case CustomStepTypeApp:
		// Run a flake app
		appName := step.Name
		if appName == "" {
			appName = "default"
		}
		appURL := flake.String() + "#" + appName
		args = []string{"nix", "run", appURL}
		if len(step.Args) > 0 {
			args = append(args, "--")
			args = append(args, step.Args...)
		}
	case CustomStepTypeDevShell:
		// Run a command in a devshell
		if len(step.Command) == 0 {
			result.Success = false
			result.Error = "devshell step has no command"
			result.Duration = time.Since(start)
			return result
		}
		shellName := step.Name
		if shellName == "" {
			shellName = "default"
		}
		shellURL := flake.String()
		if shellName != "default" {
			shellURL = shellURL + "#" + shellName
		}
		args = []string{"nix", "develop", shellURL, "-c"}
		args = append(args, step.Command...)
	default:
		result.Success = false
		result.Error = fmt.Sprintf("unknown custom step type: %s", step.Type)
		result.Duration = time.Since(start)
		return result
	}

	output, err := executeRemoteCommand(ctx, host, args)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("custom step failed: %v", err)
	}
	result.Output = output
	result.Duration = time.Since(start)

	return result
}
