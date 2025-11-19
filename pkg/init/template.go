package init

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/saberzero1/omnix/pkg/common"
)

// Param represents a template parameter
type Param struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Action      Action `json:"-" yaml:"-"` // Action is polymorphic, needs custom unmarshaling
}

// String returns a string representation of the parameter
func (p Param) String() string {
	return fmt.Sprintf("🪃 %s %s", p.Name, p.Action.String())
}

// SetValue sets the value of the parameter's action
func (p *Param) SetValue(value interface{}) {
	switch a := p.Action.(type) {
	case *ReplaceAction:
		if str, ok := value.(string); ok {
			a.Value = &str
		}
	case *RetainAction:
		if b, ok := value.(bool); ok {
			a.Value = &b
		}
	}
}

// Template represents a Nix template with parameters
type Template struct {
	Path        string  `json:"path" yaml:"path"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	WelcomeText *string `json:"welcomeText,omitempty" yaml:"welcomeText,omitempty"`
	Params      []Param `json:"-" yaml:"-"` // Custom unmarshaling needed
}

// ScaffoldAt scaffolds the template at the given output directory
func (t *Template) ScaffoldAt(ctx context.Context, outDir string) (string, error) {
	// Copy the template directory to the output directory
	if err := common.CopyDirAll(t.Path, outDir); err != nil {
		return "", fmt.Errorf("unable to copy files: %w", err)
	}

	// Apply parameter actions
	if err := t.applyActions(ctx, outDir); err != nil {
		return "", err
	}

	// Canonicalize the path
	absPath, err := filepath.Abs(outDir)
	if err != nil {
		return "", fmt.Errorf("unable to canonicalize path: %w", err)
	}

	return absPath, nil
}

// SetParamValues sets the values of parameters from a map
func (t *Template) SetParamValues(values map[string]interface{}) {
	for i := range t.Params {
		if val, ok := values[t.Params[i].Name]; ok {
			t.Params[i].SetValue(val)
		}
	}
}

// applyActions applies all parameter actions to the output directory
func (t *Template) applyActions(ctx context.Context, outDir string) error {
	// Sort params by action priority (Retain before Replace)
	sortedParams := make([]Param, len(t.Params))
	copy(sortedParams, t.Params)
	sort.Slice(sortedParams, func(i, j int) bool {
		return ActionPriority(sortedParams[i].Action) < ActionPriority(sortedParams[j].Action)
	})

	// Apply each action
	for _, param := range sortedParams {
		if param.Action.HasValue() {
			fmt.Println(param.String())
		}

		if err := param.Action.Apply(ctx, outDir); err != nil {
			return fmt.Errorf("unable to apply param %s: %w", param.Name, err)
		}
	}

	return nil
}

// ScaffoldResult represents the result of a scaffold operation
type ScaffoldResult struct {
	// OutputPath is the absolute path to the scaffolded project
	OutputPath string `json:"output_path"`
	// TemplatePath is the path to the source template
	TemplatePath string `json:"template_path"`
	// AppliedParams lists the parameters that were applied
	AppliedParams []AppliedParam `json:"applied_params"`
}

// AppliedParam represents a parameter that was applied during scaffolding
type AppliedParam struct {
	// Name is the parameter name
	Name string `json:"name"`
	// Type is the action type (replace, retain)
	Type string `json:"type"`
	// Applied indicates whether the action was applied (had a value)
	Applied bool `json:"applied"`
}

// ScaffoldAtWithResult scaffolds the template and returns detailed result information
func (t *Template) ScaffoldAtWithResult(ctx context.Context, outDir string) (*ScaffoldResult, error) {
	// Copy the template directory to the output directory
	if err := common.CopyDirAll(t.Path, outDir); err != nil {
		return nil, fmt.Errorf("unable to copy files: %w", err)
	}

	// Collect applied params
	var appliedParams []AppliedParam
	for _, param := range t.Params {
		ap := AppliedParam{
			Name:    param.Name,
			Applied: param.Action.HasValue(),
		}

		switch param.Action.(type) {
		case *ReplaceAction:
			ap.Type = "replace"
		case *RetainAction:
			ap.Type = "retain"
		default:
			ap.Type = "unknown"
		}

		appliedParams = append(appliedParams, ap)
	}

	// Apply parameter actions
	if err := t.applyActions(ctx, outDir); err != nil {
		return nil, err
	}

	// Canonicalize the path
	absPath, err := filepath.Abs(outDir)
	if err != nil {
		return nil, fmt.Errorf("unable to canonicalize path: %w", err)
	}

	result := &ScaffoldResult{
		OutputPath:    absPath,
		TemplatePath:  t.Path,
		AppliedParams: appliedParams,
	}

	return result, nil
}

// FlakeTemplate represents a template from a flake
type FlakeTemplate struct {
	TemplateName string
	Template     Template
}
