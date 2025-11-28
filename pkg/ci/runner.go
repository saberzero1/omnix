package ci

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/saberzero1/omnix/pkg/nix/store"
	"github.com/saberzero1/omnix/pkg/ui"
	"go.uber.org/zap"
)

const (
	// defaultFlakeAttr is the default flake attribute name for apps and devshells
	defaultFlakeAttr = "default"
)

// getFlakeAttrName returns the flake attribute name, defaulting to "default" if empty
func getFlakeAttrName(name string) string {
	if name == "" {
		return defaultFlakeAttr
	}
	return name
}

// buildFlakeURLWithAttr constructs a flake URL with an optional attribute.
// If the attribute is "default", it returns the base URL without a fragment.
// Otherwise, it appends "#<attr>" to the base URL.
func buildFlakeURLWithAttr(base nix.FlakeURL, attr string) string {
	if attr == defaultFlakeAttr {
		return base.String()
	}
	return base.String() + "#" + attr
}

// appendOverrideInputs appends --override-input flags for each input in the map.
// This is a helper to avoid duplicating the override inputs iteration pattern.
// Inputs are sorted alphabetically for deterministic ordering (matching Rust BTreeMap behavior).
func appendOverrideInputs(args []string, overrideInputs map[string]string) []string {
	// Sort input names for deterministic order
	inputNames := make([]string, 0, len(overrideInputs))
	for inputName := range overrideInputs {
		inputNames = append(inputNames, inputName)
	}
	sort.Strings(inputNames)

	for _, inputName := range inputNames {
		args = append(args, "--override-input", inputName, overrideInputs[inputName])
	}
	return args
}

// appendOverrideInputsForFlake appends --override-input flags with "flake/" prefix.
// This is used when passing override inputs to devour-flake, which expects override
// inputs for the target flake to be prefixed with "flake/".
// Inputs are sorted alphabetically for deterministic ordering (matching Rust BTreeMap behavior).
func appendOverrideInputsForFlake(args []string, overrideInputs map[string]string) []string {
	// Sort input names for deterministic order
	inputNames := make([]string, 0, len(overrideInputs))
	for inputName := range overrideInputs {
		inputNames = append(inputNames, inputName)
	}
	sort.Strings(inputNames)

	for _, inputName := range inputNames {
		args = append(args, "--override-input", "flake/"+inputName, overrideInputs[inputName])
	}
	return args
}

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
func Run(ctx context.Context, flake nix.FlakeURL, subflakesConfig map[string]SubflakeConfig, opts RunOptions) ([]Result, error) {
	// Sort subflake names for deterministic order (matching Rust BTreeMap behavior)
	subflakeNames := make([]string, 0, len(subflakesConfig))
	for name := range subflakesConfig {
		subflakeNames = append(subflakeNames, name)
	}
	sort.Strings(subflakeNames)

	// Collect subflakes to run in sorted order
	var subflakes []struct {
		name   string
		config SubflakeConfig
	}

	for _, name := range subflakeNames {
		subflake := subflakesConfig[name]
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

// RunWithUI executes the CI pipeline with an interactive UI
func RunWithUI(ctx context.Context, flake nix.FlakeURL, configName string, subflakesConfig map[string]SubflakeConfig, opts RunOptions) ([]Result, error) {
	// Sort subflake names for deterministic order (matching Rust BTreeMap behavior)
	sortedSubflakeNames := make([]string, 0, len(subflakesConfig))
	for name := range subflakesConfig {
		sortedSubflakeNames = append(sortedSubflakeNames, name)
	}
	sort.Strings(sortedSubflakeNames)

	// Collect subflakes to run in sorted order
	var subflakes []struct {
		name   string
		config SubflakeConfig
	}

	for _, name := range sortedSubflakeNames {
		subflake := subflakesConfig[name]
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

	// Extract subflake names for UI
	subflakeNames := make([]string, len(subflakes))
	for i, sf := range subflakes {
		subflakeNames[i] = sf.name
	}

	// Create UI model
	model := ui.NewCIRunner(flake.String(), configName, opts.Systems, subflakeNames)
	// Don't use alternate screen mode to allow native terminal interactions
	// (e.g., sudo password prompts, Ctrl+C handling)
	p := tea.NewProgram(model)

	// Use channels to safely collect results and errors
	resultsChan := make(chan resultWithIndex, len(subflakes))

	// Create cancellable context for CI operations
	// This allows us to cancel ongoing builds if user exits the UI
	ciCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Run the CI in a goroutine and send updates to the UI
	go func() {
		defer close(resultsChan)

		if opts.Parallel {
			// Run subflakes in parallel
			runSubflakesParallelWithUI(ciCtx, flake, subflakes, opts, p, resultsChan)
		} else {
			// Run subflakes sequentially
			runSubflakesSequentialWithUI(ciCtx, flake, subflakes, opts, p, resultsChan)
		}
	}()

	// Run the UI
	if _, err := p.Run(); err != nil {
		cancel() // Cancel ongoing CI operations
		return nil, fmt.Errorf("UI error: %w", err)
	}

	// Collect all results after UI is done
	resultsMap := make(map[int]Result)
	var firstError error
	for res := range resultsChan {
		resultsMap[res.index] = res.result
		if res.err != nil && firstError == nil {
			firstError = res.err
		}
	}

	// Only return results for completed subflakes, in original order
	results := make([]Result, 0, len(subflakes))
	for i := 0; i < len(subflakes); i++ {
		if result, ok := resultsMap[i]; ok {
			results = append(results, result)
		}
	}

	return results, firstError
}

// resultWithIndex holds a result with its index for channel communication
type resultWithIndex struct {
	index  int
	result Result
	err    error
}

// runSubflakesSequentialWithUI runs subflakes sequentially with UI updates
func runSubflakesSequentialWithUI(ctx context.Context, flake nix.FlakeURL, subflakes []struct {
	name   string
	config SubflakeConfig
}, opts RunOptions, p *tea.Program, resultsChan chan<- resultWithIndex) {
	var firstError error

	for i, sf := range subflakes {
		// Send subflake start message
		p.Send(ui.SubflakeStartMsg{Index: i, Name: sf.name})

		// Run the subflake and collect result
		result, err := runSubflakeWithUI(ctx, flake, sf.name, sf.config, opts, p, i)
		resultsChan <- resultWithIndex{i, result, err}

		// Send subflake complete message
		errStr := ""
		if err != nil {
			errStr = err.Error()
			if firstError == nil {
				firstError = err
			}
		}
		p.Send(ui.SubflakeCompleteMsg{Index: i, Error: errStr})

		// If there's an error, stop
		if err != nil {
			break
		}
	}

	// Send done message
	if firstError != nil {
		p.Send(ui.ErrorMsg{Err: firstError})
	} else {
		p.Send(ui.DoneMsg{})
	}
}

// runSubflakesParallelWithUI runs subflakes in parallel with UI updates
func runSubflakesParallelWithUI(ctx context.Context, flake nix.FlakeURL, subflakes []struct {
	name   string
	config SubflakeConfig
}, opts RunOptions, p *tea.Program, resultsChan chan<- resultWithIndex) {
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

	jobs := make(chan job, len(subflakes))
	errorsChan := make(chan error, len(subflakes))

	// Start worker goroutines
	for w := 0; w < maxConcurrency; w++ {
		go func() {
			for j := range jobs {
				// Send subflake start message
				p.Send(ui.SubflakeStartMsg{Index: j.index, Name: j.name})

				// Run the subflake and collect result
				result, err := runSubflakeWithUI(ctx, flake, j.name, j.config, opts, p, j.index)
				resultsChan <- resultWithIndex{j.index, result, err}

				// Send subflake complete message
				errStr := ""
				if err != nil {
					errStr = err.Error()
					errorsChan <- err
				} else {
					errorsChan <- nil
				}
				p.Send(ui.SubflakeCompleteMsg{Index: j.index, Error: errStr})
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

	// Wait for all jobs to complete
	var firstError error
	for i := 0; i < len(subflakes); i++ {
		if err := <-errorsChan; err != nil && firstError == nil {
			firstError = err
		}
	}

	// Send done message
	if firstError != nil {
		p.Send(ui.ErrorMsg{Err: firstError})
	} else {
		p.Send(ui.DoneMsg{})
	}
}

// runSubflakeWithUI runs CI for a single subflake and sends UI updates
func runSubflakeWithUI(ctx context.Context, flake nix.FlakeURL, name string, subflake SubflakeConfig, opts RunOptions, p *tea.Program, subflakeIndex int) (Result, error) {
	start := time.Now()

	result := Result{
		Subflake: name,
		Steps:    make(map[string]StepResult),
		Success:  true,
	}

	// Get the subflake URL using the correct method
	subflakeURL := flake.SubFlakeURL(subflake.Dir)

	// Run build step
	if subflake.Steps.Build.Enable {
		p.Send(ui.StepStartMsg{SubflakeIndex: subflakeIndex, StepName: "build"})

		var stepResult StepResult
		if opts.RemoteHost != "" {
			stepResult = runBuildStepRemote(ctx, opts.RemoteHost, subflakeURL, subflake.Steps.Build, opts, subflake)
		} else {
			stepResult = runBuildStep(ctx, subflakeURL, subflake.Steps.Build, opts, subflake)
		}
		result.Steps["build"] = stepResult

		p.Send(ui.StepCompleteMsg{
			SubflakeIndex: subflakeIndex,
			Output:        stepResult.Output,
			Error:         stepResult.Error,
		})

		if !stepResult.Success {
			result.Success = false
		}
	}

	// Run lockfile step
	if subflake.Steps.Lockfile.Enable {
		// Check if we should skip
		if len(subflake.OverrideInputs) > 0 {
			p.Send(ui.StepSkipMsg{
				SubflakeIndex: subflakeIndex,
				StepName:      "lockfile",
				Reason:        "Skipped (has override inputs)",
			})
		} else {
			p.Send(ui.StepStartMsg{SubflakeIndex: subflakeIndex, StepName: "lockfile"})

			var stepResult StepResult
			if opts.RemoteHost != "" {
				stepResult = runLockfileStepRemote(ctx, opts.RemoteHost, subflakeURL, subflake.Steps.Lockfile, subflake)
			} else {
				stepResult = runLockfileStep(ctx, subflakeURL, subflake.Steps.Lockfile, subflake)
			}
			result.Steps["lockfile"] = stepResult

			p.Send(ui.StepCompleteMsg{
				SubflakeIndex: subflakeIndex,
				Output:        stepResult.Output,
				Error:         stepResult.Error,
			})

			if !stepResult.Success {
				result.Success = false
			}
		}
	}

	// Run flake check step
	if subflake.Steps.FlakeCheck.Enable {
		p.Send(ui.StepStartMsg{SubflakeIndex: subflakeIndex, StepName: "flakeCheck"})

		var stepResult StepResult
		if opts.RemoteHost != "" {
			stepResult = runFlakeCheckStepRemote(ctx, opts.RemoteHost, subflakeURL, subflake.Steps.FlakeCheck, subflake)
		} else {
			stepResult = runFlakeCheckStep(ctx, subflakeURL, subflake.Steps.FlakeCheck, subflake)
		}
		result.Steps["flakeCheck"] = stepResult

		p.Send(ui.StepCompleteMsg{
			SubflakeIndex: subflakeIndex,
			Output:        stepResult.Output,
			Error:         stepResult.Error,
		})

		if !stepResult.Success {
			result.Success = false
		}
	}

	// Run custom steps in deterministic (sorted) order
	customStepNames := make([]string, 0, len(subflake.Steps.Custom))
	for stepName := range subflake.Steps.Custom {
		customStepNames = append(customStepNames, stepName)
	}
	sort.Strings(customStepNames)

	for _, stepName := range customStepNames {
		customStep := subflake.Steps.Custom[stepName]
		// Check if this step can run on current systems
		if !customStep.CanRunOn(opts.Systems) {
			p.Send(ui.StepSkipMsg{
				SubflakeIndex: subflakeIndex,
				StepName:      "custom:" + stepName,
				Reason:        "Skipped (system not supported)",
			})
			continue
		}

		p.Send(ui.StepStartMsg{SubflakeIndex: subflakeIndex, StepName: "custom:" + stepName})

		var stepResult StepResult
		if opts.RemoteHost != "" {
			stepResult = runCustomStepRemote(ctx, opts.RemoteHost, subflakeURL, stepName, customStep, subflake)
		} else {
			stepResult = runCustomStep(ctx, subflakeURL, stepName, customStep, subflake)
		}
		result.Steps["custom:"+stepName] = stepResult

		p.Send(ui.StepCompleteMsg{
			SubflakeIndex: subflakeIndex,
			Output:        stepResult.Output,
			Error:         stepResult.Error,
		})

		if !stepResult.Success {
			result.Success = false
		}
	}

	result.Duration = time.Since(start)
	return result, nil
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

	// Get the subflake URL using the correct method
	subflakeURL := flake.SubFlakeURL(subflake.Dir)

	// Run build step
	if subflake.Steps.Build.Enable {
		var stepResult StepResult
		if opts.RemoteHost != "" {
			stepResult = runBuildStepRemote(ctx, opts.RemoteHost, subflakeURL, subflake.Steps.Build, opts, subflake)
		} else {
			stepResult = runBuildStep(ctx, subflakeURL, subflake.Steps.Build, opts, subflake)
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
			stepResult = runLockfileStepRemote(ctx, opts.RemoteHost, subflakeURL, subflake.Steps.Lockfile, subflake)
		} else {
			stepResult = runLockfileStep(ctx, subflakeURL, subflake.Steps.Lockfile, subflake)
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
			stepResult = runFlakeCheckStepRemote(ctx, opts.RemoteHost, subflakeURL, subflake.Steps.FlakeCheck, subflake)
		} else {
			stepResult = runFlakeCheckStep(ctx, subflakeURL, subflake.Steps.FlakeCheck, subflake)
		}
		result.Steps["flakeCheck"] = stepResult
		if !stepResult.Success {
			result.Success = false
		}
	}

	// Run custom steps in deterministic (sorted) order
	customStepNames := make([]string, 0, len(subflake.Steps.Custom))
	for name := range subflake.Steps.Custom {
		customStepNames = append(customStepNames, name)
	}
	sort.Strings(customStepNames)

	for _, name := range customStepNames {
		customStep := subflake.Steps.Custom[name]
		// Check if this step can run on current systems
		if !customStep.CanRunOn(opts.Systems) {
			continue
		}

		var stepResult StepResult
		if opts.RemoteHost != "" {
			stepResult = runCustomStepRemote(ctx, opts.RemoteHost, subflakeURL, name, customStep, subflake)
		} else {
			stepResult = runCustomStep(ctx, subflakeURL, name, customStep, subflake)
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
func runBuildStep(ctx context.Context, flake nix.FlakeURL, step BuildStep, opts RunOptions, subflake SubflakeConfig) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "build",
		Success: true,
	}

	// Use devour-flake to build all outputs with override inputs
	output, err := nix.DevourFlakeWithOverrides(ctx, flake, opts.Systems, step.Impure, subflake.OverrideInputs)
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

	// If requested, fetch all dependencies
	if opts.IncludeAllDependencies && len(output.OutPaths) > 0 {
		allDeps, err := fetchAllDependencies(ctx, output.OutPaths)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to fetch dependencies: %v", err)
		} else {
			result.Output += fmt.Sprintf("\n\nTotal dependencies: %d", len(allDeps))
		}
	}

	result.Duration = time.Since(start)

	return result
}

// fetchAllDependencies fetches all build and runtime dependencies for the given paths
func fetchAllDependencies(ctx context.Context, paths []store.Path) ([]store.Path, error) {
	cmd := store.NewCmd()
	return cmd.FetchAllDeps(ctx, paths)
}

// runLockfileStep executes the lockfile check step
func runLockfileStep(ctx context.Context, flake nix.FlakeURL, step LockfileStep, subflake SubflakeConfig) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "lockfile",
		Success: true,
	}

	// Skip lockfile check if there are override inputs (like Rust version)
	if len(subflake.OverrideInputs) > 0 {
		result.Output = "Skipped (has override inputs)"
		result.Duration = time.Since(start)
		return result
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
func runFlakeCheckStep(ctx context.Context, flake nix.FlakeURL, step FlakeCheckStep, subflake SubflakeConfig) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "flakeCheck",
		Success: true,
	}

	// Build nix flake check command with override inputs
	args := []string{"flake", "check", flake.String()}
	args = appendOverrideInputs(args, subflake.OverrideInputs)

	// Run nix flake check
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

// runCustomStep executes a custom step
func runCustomStep(ctx context.Context, flake nix.FlakeURL, name string, step CustomStep, subflake SubflakeConfig) StepResult {
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
		output, err = runFlakeApp(ctx, flake, step, subflake)
	case CustomStepTypeDevShell:
		// Run a command in a devshell
		output, err = runDevShellCommand(ctx, flake, step, subflake)
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
func runFlakeApp(ctx context.Context, flake nix.FlakeURL, step CustomStep, subflake SubflakeConfig) (string, error) {
	appName := getFlakeAttrName(step.Name)
	appURL := buildFlakeURLWithAttr(flake, appName)

	// Build nix run command
	args := []string{"run", appURL}
	args = appendOverrideInputs(args, subflake.OverrideInputs)

	if len(step.Args) > 0 {
		args = append(args, "--")
		args = append(args, step.Args...)
	}

	cmd := nix.NewCmd()
	return cmd.Run(ctx, args...)
}

// runDevShellCommand runs a command in a devshell
func runDevShellCommand(ctx context.Context, flake nix.FlakeURL, step CustomStep, subflake SubflakeConfig) (string, error) {
	if len(step.Command) == 0 {
		return "", fmt.Errorf("devshell step has no command")
	}

	shellName := getFlakeAttrName(step.Name)
	shellURL := buildFlakeURLWithAttr(flake, shellName)

	// Build nix develop command
	args := []string{"develop", shellURL}
	args = appendOverrideInputs(args, subflake.OverrideInputs)

	args = append(args, "-c")
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
func runBuildStepRemote(ctx context.Context, host string, flake nix.FlakeURL, step BuildStep, opts RunOptions, subflake SubflakeConfig) StepResult {
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

	// Add systems filtering if specified
	if len(opts.Systems) > 0 {
		systemsFlakeURL, err := nix.GetSystemsFlakeURL(opts.Systems)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to get systems flake URL: %v", err)
			result.Duration = time.Since(start)
			return result
		}
		args = append(args, "--override-input", "systems", systemsFlakeURL.String())
	}

	// Add override inputs from subflake config with "flake/" prefix
	// This is necessary because devour-flake expects override inputs for the
	// target flake to be prefixed with "flake/"
	args = appendOverrideInputsForFlake(args, subflake.OverrideInputs)

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
func runLockfileStepRemote(ctx context.Context, host string, flake nix.FlakeURL, step LockfileStep, subflake SubflakeConfig) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "lockfile",
		Success: true,
	}

	// Skip lockfile check if there are override inputs (like Rust version)
	if len(subflake.OverrideInputs) > 0 {
		result.Output = "Skipped (has override inputs)"
		result.Duration = time.Since(start)
		return result
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
func runFlakeCheckStepRemote(ctx context.Context, host string, flake nix.FlakeURL, step FlakeCheckStep, subflake SubflakeConfig) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "flakeCheck",
		Success: true,
	}

	args := []string{"nix", "flake", "check", flake.String()}
	args = appendOverrideInputs(args, subflake.OverrideInputs)

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
func runCustomStepRemote(ctx context.Context, host string, flake nix.FlakeURL, name string, step CustomStep, subflake SubflakeConfig) StepResult {
	start := time.Now()
	result := StepResult{
		Name:    "custom:" + name,
		Success: true,
	}

	var args []string

	switch step.Type {
	case CustomStepTypeApp:
		// Run a flake app
		appName := getFlakeAttrName(step.Name)
		appURL := buildFlakeURLWithAttr(flake, appName)
		args = []string{"nix", "run", appURL}
		args = appendOverrideInputs(args, subflake.OverrideInputs)

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
		shellName := getFlakeAttrName(step.Name)
		shellURL := buildFlakeURLWithAttr(flake, shellName)
		args = []string{"nix", "develop", shellURL}
		args = appendOverrideInputs(args, subflake.OverrideInputs)

		args = append(args, "-c")
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
