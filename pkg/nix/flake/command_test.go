package flake

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCmdForCommands is a mock implementation of the Cmd interface for testing commands
type mockCmdForCommands struct {
	runFunc func(ctx context.Context, args ...string) (string, error)
}

func (m *mockCmdForCommands) Run(ctx context.Context, args ...string) (string, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, args...)
	}
	return "", nil
}

func TestOutPath_FirstOutput(t *testing.T) {
	tests := []struct {
		name    string
		outPath OutPath
		want    *string
	}{
		{
			name: "single output",
			outPath: OutPath{
				Outputs: map[string]string{
					"out": "/nix/store/abc-hello",
				},
			},
			want: strPtr("/nix/store/abc-hello"),
		},
		{
			name: "multiple outputs",
			outPath: OutPath{
				Outputs: map[string]string{
					"out": "/nix/store/abc-hello",
					"dev": "/nix/store/def-hello-dev",
				},
			},
			want: strPtr("/nix/store/abc-hello"), // or dev, order is not guaranteed
		},
		{
			name: "no outputs",
			outPath: OutPath{
				Outputs: map[string]string{},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.outPath.FirstOutput()
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				// For multiple outputs, just check we got something
				if len(tt.outPath.Outputs) > 0 {
					assert.NotEmpty(t, *got)
				}
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func TestRun(t *testing.T) {
	tests := []struct {
		name        string
		opts        *CommandOptions
		url         string
		appArgs     []string
		expectedCmd []string
		mockError   error
		wantErr     bool
	}{
		{
			name:        "simple run",
			opts:        nil,
			url:         "nixpkgs#hello",
			appArgs:     []string{},
			expectedCmd: []string{"run", "nixpkgs#hello", "--"},
			wantErr:     false,
		},
		{
			name: "run with override inputs",
			opts: &CommandOptions{
				OverrideInputs: map[string]string{
					"nixpkgs": "github:NixOS/nixpkgs/nixos-unstable",
				},
			},
			url:         ".#default",
			appArgs:     []string{"--version"},
			expectedCmd: []string{"run", "--override-input", "nixpkgs", "github:NixOS/nixpkgs/nixos-unstable", ".#default", "--", "--version"},
			wantErr:     false,
		},
		{
			name:      "run with error",
			opts:      nil,
			url:       "nixpkgs#hello",
			appArgs:   []string{},
			mockError: errors.New("command failed"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCmdForCommands{
				runFunc: func(ctx context.Context, args ...string) (string, error) {
					if tt.mockError == nil {
						assert.Equal(t, tt.expectedCmd, args)
					}
					return "", tt.mockError
				},
			}

			err := Run(context.Background(), mock, tt.opts, tt.url, tt.appArgs)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDevelop(t *testing.T) {
	tests := []struct {
		name        string
		opts        *CommandOptions
		url         string
		command     []string
		expectedCmd []string
		wantErr     bool
	}{
		{
			name:        "simple develop",
			opts:        nil,
			url:         ".#default",
			command:     []string{"bash"},
			expectedCmd: []string{"develop", ".#default", "-c", "bash"},
			wantErr:     false,
		},
		{
			name:        "develop with multiple command args",
			opts:        nil,
			url:         ".#default",
			command:     []string{"bash", "-c", "echo hello"},
			expectedCmd: []string{"develop", ".#default", "-c", "bash", "-c", "echo hello"},
			wantErr:     false,
		},
		{
			name:    "develop with empty command",
			opts:    nil,
			url:     ".#default",
			command: []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCmdForCommands{
				runFunc: func(ctx context.Context, args ...string) (string, error) {
					if !tt.wantErr {
						assert.Equal(t, tt.expectedCmd, args)
					}
					return "", nil
				},
			}

			err := Develop(context.Background(), mock, tt.opts, tt.url, tt.command)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name        string
		opts        *CommandOptions
		url         string
		mockOutput  string
		expectedCmd []string
		wantErr     bool
		wantPaths   []OutPath
	}{
		{
			name: "successful build",
			opts: nil,
			url:  ".#default",
			mockOutput: `[
				{
					"drvPath": "/nix/store/abc.drv",
					"outputs": {
						"out": "/nix/store/xyz-hello"
					}
				}
			]`,
			expectedCmd: []string{"build", "--no-link", "--json", ".#default"},
			wantErr:     false,
			wantPaths: []OutPath{
				{
					DrvPath: "/nix/store/abc.drv",
					Outputs: map[string]string{
						"out": "/nix/store/xyz-hello",
					},
				},
			},
		},
		{
			name:        "invalid json output",
			opts:        nil,
			url:         ".#default",
			mockOutput:  "invalid json",
			expectedCmd: []string{"build", "--no-link", "--json", ".#default"},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCmdForCommands{
				runFunc: func(ctx context.Context, args ...string) (string, error) {
					assert.Equal(t, tt.expectedCmd, args)
					return tt.mockOutput, nil
				},
			}

			paths, err := Build(context.Background(), mock, tt.opts, tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantPaths, paths)
			}
		})
	}
}

func TestFlakeLock(t *testing.T) {
	tests := []struct {
		name        string
		opts        *CommandOptions
		url         string
		extraArgs   []string
		expectedCmd []string
		wantErr     bool
	}{
		{
			name:        "simple lock",
			opts:        nil,
			url:         ".",
			extraArgs:   []string{},
			expectedCmd: []string{"flake", "lock", "."},
			wantErr:     false,
		},
		{
			name:        "lock with update input",
			opts:        nil,
			url:         ".",
			extraArgs:   []string{"--update-input", "nixpkgs"},
			expectedCmd: []string{"flake", "lock", ".", "--update-input", "nixpkgs"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCmdForCommands{
				runFunc: func(ctx context.Context, args ...string) (string, error) {
					assert.Equal(t, tt.expectedCmd, args)
					return "", nil
				},
			}

			err := FlakeLock(context.Background(), mock, tt.opts, tt.url, tt.extraArgs)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name        string
		opts        *CommandOptions
		url         string
		expectedCmd []string
		wantErr     bool
	}{
		{
			name:        "simple check",
			opts:        nil,
			url:         ".",
			expectedCmd: []string{"flake", "check", "."},
			wantErr:     false,
		},
		{
			name: "check with no-write-lock-file",
			opts: &CommandOptions{
				NoWriteLockFile: true,
			},
			url:         ".",
			expectedCmd: []string{"flake", "check", ".", "--no-write-lock-file"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCmdForCommands{
				runFunc: func(ctx context.Context, args ...string) (string, error) {
					assert.Equal(t, tt.expectedCmd, args)
					return "", nil
				},
			}

			err := Check(context.Background(), mock, tt.opts, tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShow(t *testing.T) {
	tests := []struct {
		name        string
		opts        *CommandOptions
		url         string
		mockOutput  string
		expectedCmd []string
		wantOutput  string
		wantErr     bool
	}{
		{
			name:        "simple show",
			opts:        nil,
			url:         ".",
			mockOutput:  "flake output\n",
			expectedCmd: []string{"flake", "show", "."},
			wantOutput:  "flake output",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCmdForCommands{
				runFunc: func(ctx context.Context, args ...string) (string, error) {
					assert.Equal(t, tt.expectedCmd, args)
					return tt.mockOutput, nil
				},
			}

			output, err := Show(context.Background(), mock, tt.opts, tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantOutput, output)
			}
		})
	}
}

func TestApplyOptions(t *testing.T) {
	tests := []struct {
		name string
		args []string
		opts *CommandOptions
		want []string
	}{
		{
			name: "nil options",
			args: []string{"build"},
			opts: nil,
			want: []string{"build"},
		},
		{
			name: "empty options",
			args: []string{"build"},
			opts: &CommandOptions{},
			want: []string{"build"},
		},
		{
			name: "with override inputs",
			args: []string{"build"},
			opts: &CommandOptions{
				OverrideInputs: map[string]string{
					"nixpkgs": "github:NixOS/nixpkgs",
				},
			},
			want: []string{"build", "--override-input", "nixpkgs", "github:NixOS/nixpkgs"},
		},
		{
			name: "with no-write-lock-file",
			args: []string{"build"},
			opts: &CommandOptions{
				NoWriteLockFile: true,
			},
			want: []string{"build", "--no-write-lock-file"},
		},
		{
			name: "with multiple options",
			args: []string{"build"},
			opts: &CommandOptions{
				OverrideInputs: map[string]string{
					"nixpkgs":     "github:NixOS/nixpkgs",
					"flake-utils": "github:numtide/flake-utils",
				},
				NoWriteLockFile: true,
			},
			// Note: map iteration order is not guaranteed, so we'll check length and contents
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyOptions(tt.args, tt.opts)
			if tt.name == "with multiple options" {
				// For multiple override inputs, just check we have the right components
				assert.Contains(t, strings.Join(got, " "), "--override-input")
				assert.Contains(t, strings.Join(got, " "), "nixpkgs")
				assert.Contains(t, strings.Join(got, " "), "flake-utils")
				assert.Contains(t, strings.Join(got, " "), "--no-write-lock-file")
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
