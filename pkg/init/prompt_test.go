package init

import (
	"testing"
)

func TestValidateRequiredParams(t *testing.T) {
	strPtr := func(s string) *string { return &s }
	boolPtr := func(b bool) *bool { return &b }

	tests := []struct {
		name    string
		params  []Param
		wantErr bool
	}{
		{
			name: "All params have values",
			params: []Param{
				{
					Name:        "name",
					Description: "Project name",
					Action: &ReplaceAction{
						Placeholder: "PROJECT",
						Value:       strPtr("my-project"),
					},
				},
				{
					Name:        "enable-ci",
					Description: "Enable CI",
					Action: &RetainAction{
						Paths: []string{".github/**"},
						Value: boolPtr(true),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing ReplaceAction value",
			params: []Param{
				{
					Name:        "name",
					Description: "Project name",
					Action: &ReplaceAction{
						Placeholder: "PROJECT",
						Value:       nil,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Missing RetainAction value",
			params: []Param{
				{
					Name:        "enable-ci",
					Description: "Enable CI",
					Action: &RetainAction{
						Paths: []string{".github/**"},
						Value: nil,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Mixed - some with values, some without",
			params: []Param{
				{
					Name:        "name",
					Description: "Project name",
					Action: &ReplaceAction{
						Placeholder: "PROJECT",
						Value:       strPtr("my-project"),
					},
				},
				{
					Name:        "author",
					Description: "Author name",
					Action: &ReplaceAction{
						Placeholder: "AUTHOR",
						Value:       nil,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequiredParams(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequiredParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
