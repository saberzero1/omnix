package init

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptForParams interactively prompts the user for parameter values
// Returns a map of parameter name to value
func PromptForParams(params []Param) (map[string]interface{}, error) {
	reader := bufio.NewReader(os.Stdin)
	values := make(map[string]interface{})

	for _, param := range params {
		// Skip if the parameter already has a value
		if param.Action.HasValue() {
			continue
		}

		// Determine the type of input needed based on the action
		switch param.Action.(type) {
		case *ReplaceAction:
			// Prompt for string input
			fmt.Printf("\n%s\n", param.Description)
			fmt.Printf("Enter value for '%s': ", param.Name)

			input, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("failed to read input for '%s': %w", param.Name, err)
			}

			value := strings.TrimSpace(input)
			if value != "" {
				values[param.Name] = value
			}

		case *RetainAction, *ChmodAction, *MoveAction:
			// Prompt for boolean input (enable/disable action)
			fmt.Printf("\n%s\n", param.Description)
			fmt.Printf("Enable '%s'? (y/n) [default: n]: ", param.Name)

			input, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("failed to read input for '%s': %w", param.Name, err)
			}

			value := strings.TrimSpace(strings.ToLower(input))
			switch value {
			case "y", "yes":
				values[param.Name] = true
			case "n", "no", "":
				values[param.Name] = false
			default:
				fmt.Printf("Invalid input '%s', defaulting to 'no'\n", value)
				values[param.Name] = false
			}
		}
	}

	return values, nil
}

// ValidateRequiredParams checks that all parameters have values
// Returns an error if any parameter is missing a value
func ValidateRequiredParams(params []Param) error {
	var missing []string

	for _, param := range params {
		if !param.Action.HasValue() {
			missing = append(missing, param.Name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required parameters: %s", strings.Join(missing, ", "))
	}

	return nil
}
