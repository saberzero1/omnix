package flake

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCmd is a simple mock implementation of the Cmd interface for testing.
type mockCmd struct {
	output string
	err    error
}

func (m *mockCmd) Run(ctx context.Context, args ...string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.output, nil
}

func TestEvalExpr(t *testing.T) {
	tests := []struct {
		name       string
		expr       string
		mockOutput string
		want       interface{}
		wantErr    bool
	}{
		{
			name:       "simple number",
			expr:       "42",
			mockOutput: "42",
			want:       float64(42),
		},
		{
			name:       "simple string",
			expr:       `"hello"`,
			mockOutput: `"hello"`,
			want:       "hello",
		},
		{
			name:       "simple list",
			mockOutput: "[1,2,3]",
			want:       []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:       "simple attrset",
			mockOutput: `{"foo":"bar"}`,
			want:       map[string]interface{}{"foo": "bar"},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &mockCmd{output: tt.mockOutput}
			var result interface{}
			result, err := EvalExpr[interface{}](ctx, cmd, tt.expr)
			
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
	}
}

func TestEvalExprTyped(t *testing.T) {
	ctx := context.Background()

	t.Run("string type", func(t *testing.T) {
		cmd := &mockCmd{output: `"hello world"`}
		result, err := EvalExpr[string](ctx, cmd, `"hello world"`)
		require.NoError(t, err)
		assert.Equal(t, "hello world", result)
	})

	t.Run("int type (as float64)", func(t *testing.T) {
		cmd := &mockCmd{output: "123"}
		result, err := EvalExpr[float64](ctx, cmd, "123")
		require.NoError(t, err)
		assert.Equal(t, float64(123), result)
	})

	t.Run("list type", func(t *testing.T) {
		cmd := &mockCmd{output: `["a","b","c"]`}
		result, err := EvalExpr[[]string](ctx, cmd, `["a" "b" "c"]`)
		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})

	t.Run("map type", func(t *testing.T) {
		cmd := &mockCmd{output: `{"x":"foo","y":"bar"}`}
		result, err := EvalExpr[map[string]string](ctx, cmd, `{x = "foo"; y = "bar";}`)
		require.NoError(t, err)
		expected := map[string]string{"x": "foo", "y": "bar"}
		assert.Equal(t, expected, result)
	})
}

func TestEvalImpureExpr(t *testing.T) {
	ctx := context.Background()

	t.Run("impure expression", func(t *testing.T) {
		cmd := &mockCmd{output: "1234567890"}
		result, err := EvalImpureExpr[float64](ctx, cmd, "builtins.currentTime")
		require.NoError(t, err)
		assert.Equal(t, float64(1234567890), result)
	})
}

func TestFlakeOptions(t *testing.T) {
	tests := []struct {
		name string
		opts *FlakeOptions
	}{
		{
			name: "nil options",
			opts: nil,
		},
		{
			name: "empty options",
			opts: &FlakeOptions{},
		},
		{
			name: "impure only",
			opts: &FlakeOptions{Impure: true},
		},
		{
			name: "refresh only",
			opts: &FlakeOptions{Refresh: true},
		},
		{
			name: "with override inputs",
			opts: &FlakeOptions{
				OverrideInputs: map[string]string{
					"nixpkgs": "github:NixOS/nixpkgs/nixos-unstable",
				},
			},
		},
		{
			name: "all options",
			opts: &FlakeOptions{
				Impure:  true,
				Refresh: true,
				OverrideInputs: map[string]string{
					"nixpkgs": "github:NixOS/nixpkgs/nixos-unstable",
				},
			},
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &mockCmd{output: "42"}
			result, err := Eval[float64](ctx, cmd, tt.opts, "42")
			require.NoError(t, err)
			assert.Equal(t, float64(42), result)
		})
	}
}

func TestIsMissingAttributeError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "missing attribute error",
			err:  assert.AnError,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isMissingAttributeError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEvalMaybe(t *testing.T) {
	ctx := context.Background()

	t.Run("existing expression", func(t *testing.T) {
		cmd := &mockCmd{output: "42"}
		result, err := EvalMaybe[float64](ctx, cmd, nil, "42")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, float64(42), *result)
	})
}
