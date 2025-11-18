package nix

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Version
		wantErr bool
	}{
		{
			name:  "standard nix format",
			input: "nix (Nix) 2.13.0",
			want: Version{
				Major: 2,
				Minor: 13,
				Patch: 0,
			},
			wantErr: false,
		},
		{
			name:  "simple version format",
			input: "2.13.0",
			want: Version{
				Major: 2,
				Minor: 13,
				Patch: 0,
			},
			wantErr: false,
		},
		{
			name:  "determinate nix format",
			input: "nix (Determinate Nix 3.6.6) 2.29.0",
			want: Version{
				Major: 2,
				Minor: 29,
				Patch: 0,
			},
			wantErr: false,
		},
		{
			name:  "nix 2.18.0",
			input: "nix (Nix) 2.18.0",
			want: Version{
				Major: 2,
				Minor: 18,
				Patch: 0,
			},
			wantErr: false,
		},
		{
			name:  "nix 2.3.10",
			input: "nix (Nix) 2.3.10",
			want: Version{
				Major: 2,
				Minor: 3,
				Patch: 10,
			},
			wantErr: false,
		},
		{
			name:    "invalid format - no version",
			input:   "nix (Nix)",
			want:    Version{},
			wantErr: true,
		},
		{
			name:    "invalid format - empty string",
			input:   "",
			want:    Version{},
			wantErr: true,
		},
		{
			name:    "invalid format - random text",
			input:   "hello world",
			want:    Version{},
			wantErr: true,
		},
		{
			name:    "invalid format - partial version",
			input:   "2.13",
			want:    Version{},
			wantErr: true,
		},
		{
			name:    "invalid format - letters in version",
			input:   "nix (Nix) 2.x.0",
			want:    Version{},
			wantErr: true,
		},
		{
			name:    "invalid format - non-numeric",
			input:   "nix (Nix) a.b.c",
			want:    Version{},
			wantErr: true,
		},
		{
			name:    "invalid format - only text",
			input:   "version info",
			want:    Version{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version Version
		want    string
	}{
		{
			name: "standard version",
			version: Version{
				Major: 2,
				Minor: 13,
				Patch: 0,
			},
			want: "2.13.0",
		},
		{
			name: "version with patch",
			version: Version{
				Major: 2,
				Minor: 3,
				Patch: 10,
			},
			want: "2.3.10",
		},
		{
			name: "version 0.0.0",
			version: Version{
				Major: 0,
				Minor: 0,
				Patch: 0,
			},
			want: "0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.String(); got != tt.want {
				t.Errorf("Version.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionCompare(t *testing.T) {
	tests := []struct {
		name string
		v1   Version
		v2   Version
		want int
	}{
		{
			name: "equal versions",
			v1:   Version{2, 13, 0},
			v2:   Version{2, 13, 0},
			want: 0,
		},
		{
			name: "v1 less than v2 - major",
			v1:   Version{1, 13, 0},
			v2:   Version{2, 13, 0},
			want: -1,
		},
		{
			name: "v1 greater than v2 - major",
			v1:   Version{3, 13, 0},
			v2:   Version{2, 13, 0},
			want: 1,
		},
		{
			name: "v1 less than v2 - minor",
			v1:   Version{2, 12, 0},
			v2:   Version{2, 13, 0},
			want: -1,
		},
		{
			name: "v1 greater than v2 - minor",
			v1:   Version{2, 14, 0},
			v2:   Version{2, 13, 0},
			want: 1,
		},
		{
			name: "v1 less than v2 - patch",
			v1:   Version{2, 13, 0},
			v2:   Version{2, 13, 1},
			want: -1,
		},
		{
			name: "v1 greater than v2 - patch",
			v1:   Version{2, 13, 2},
			v2:   Version{2, 13, 1},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.Compare(tt.v2); got != tt.want {
				t.Errorf("Version.Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionLessThan(t *testing.T) {
	tests := []struct {
		name string
		v1   Version
		v2   Version
		want bool
	}{
		{
			name: "less than - major",
			v1:   Version{1, 0, 0},
			v2:   Version{2, 0, 0},
			want: true,
		},
		{
			name: "less than - minor",
			v1:   Version{2, 12, 0},
			v2:   Version{2, 13, 0},
			want: true,
		},
		{
			name: "less than - patch",
			v1:   Version{2, 13, 0},
			v2:   Version{2, 13, 1},
			want: true,
		},
		{
			name: "equal versions",
			v1:   Version{2, 13, 0},
			v2:   Version{2, 13, 0},
			want: false,
		},
		{
			name: "greater than",
			v1:   Version{3, 0, 0},
			v2:   Version{2, 0, 0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.LessThan(tt.v2); got != tt.want {
				t.Errorf("Version.LessThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionGreaterThan(t *testing.T) {
	tests := []struct {
		name string
		v1   Version
		v2   Version
		want bool
	}{
		{
			name: "greater than - major",
			v1:   Version{3, 0, 0},
			v2:   Version{2, 0, 0},
			want: true,
		},
		{
			name: "greater than - minor",
			v1:   Version{2, 14, 0},
			v2:   Version{2, 13, 0},
			want: true,
		},
		{
			name: "greater than - patch",
			v1:   Version{2, 13, 1},
			v2:   Version{2, 13, 0},
			want: true,
		},
		{
			name: "equal versions",
			v1:   Version{2, 13, 0},
			v2:   Version{2, 13, 0},
			want: false,
		},
		{
			name: "less than",
			v1:   Version{1, 0, 0},
			v2:   Version{2, 0, 0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.GreaterThan(tt.v2); got != tt.want {
				t.Errorf("Version.GreaterThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionEqual(t *testing.T) {
	tests := []struct {
		name string
		v1   Version
		v2   Version
		want bool
	}{
		{
			name: "equal versions",
			v1:   Version{2, 13, 0},
			v2:   Version{2, 13, 0},
			want: true,
		},
		{
			name: "different major",
			v1:   Version{1, 13, 0},
			v2:   Version{2, 13, 0},
			want: false,
		},
		{
			name: "different minor",
			v1:   Version{2, 12, 0},
			v2:   Version{2, 13, 0},
			want: false,
		},
		{
			name: "different patch",
			v1:   Version{2, 13, 0},
			v2:   Version{2, 13, 1},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v1.Equal(tt.v2); got != tt.want {
				t.Errorf("Version.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
