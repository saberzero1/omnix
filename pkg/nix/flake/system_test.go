package flake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSystem(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantStr  string
		wantOS   string
		wantArch Arch
	}{
		{
			name:     "aarch64-linux",
			input:    "aarch64-linux",
			wantStr:  "aarch64-linux",
			wantOS:   "linux",
			wantArch: ArchAarch64,
		},
		{
			name:     "x86_64-linux",
			input:    "x86_64-linux",
			wantStr:  "x86_64-linux",
			wantOS:   "linux",
			wantArch: ArchX86_64,
		},
		{
			name:     "x86_64-darwin",
			input:    "x86_64-darwin",
			wantStr:  "x86_64-darwin",
			wantOS:   "darwin",
			wantArch: ArchX86_64,
		},
		{
			name:     "aarch64-darwin",
			input:    "aarch64-darwin",
			wantStr:  "aarch64-darwin",
			wantOS:   "darwin",
			wantArch: ArchAarch64,
		},
		{
			name:    "unknown system",
			input:   "riscv64-linux",
			wantStr: "riscv64-linux",
			wantOS:  "riscv64-linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := ParseSystem(tt.input)
			assert.Equal(t, tt.wantStr, sys.String())
			assert.Equal(t, tt.wantOS, sys.os)
			if tt.wantArch != 0 || tt.input != "riscv64-linux" {
				assert.Equal(t, tt.wantArch, sys.arch)
			}
		})
	}
}

func TestSystem_String(t *testing.T) {
	tests := []struct {
		name   string
		system System
		want   string
	}{
		{
			name:   "Linux ARM",
			system: SystemLinuxAarch64,
			want:   "aarch64-linux",
		},
		{
			name:   "Linux x86_64",
			system: SystemLinuxX86_64,
			want:   "x86_64-linux",
		},
		{
			name:   "Darwin x86_64",
			system: SystemDarwinX86_64,
			want:   "x86_64-darwin",
		},
		{
			name:   "Darwin ARM",
			system: SystemDarwinAarch64,
			want:   "aarch64-darwin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.system.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSystem_HumanReadable(t *testing.T) {
	tests := []struct {
		name   string
		system System
		want   string
	}{
		{
			name:   "Linux ARM",
			system: SystemLinuxAarch64,
			want:   "Linux (ARM)",
		},
		{
			name:   "Linux Intel",
			system: SystemLinuxX86_64,
			want:   "Linux (Intel)",
		},
		{
			name:   "macOS Intel",
			system: SystemDarwinX86_64,
			want:   "macOS (Intel)",
		},
		{
			name:   "macOS ARM",
			system: SystemDarwinAarch64,
			want:   "macOS (ARM)",
		},
		{
			name:   "Custom system",
			system: ParseSystem("custom-system"),
			want:   "custom-system",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.system.HumanReadable()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSystem_IsLinux(t *testing.T) {
	assert.True(t, SystemLinuxAarch64.IsLinux())
	assert.True(t, SystemLinuxX86_64.IsLinux())
	assert.False(t, SystemDarwinX86_64.IsLinux())
	assert.False(t, SystemDarwinAarch64.IsLinux())
}

func TestSystem_IsDarwin(t *testing.T) {
	assert.False(t, SystemLinuxAarch64.IsDarwin())
	assert.False(t, SystemLinuxX86_64.IsDarwin())
	assert.True(t, SystemDarwinX86_64.IsDarwin())
	assert.True(t, SystemDarwinAarch64.IsDarwin())
}

func TestSystem_GetArch(t *testing.T) {
	assert.Equal(t, ArchAarch64, SystemLinuxAarch64.GetArch())
	assert.Equal(t, ArchX86_64, SystemLinuxX86_64.GetArch())
	assert.Equal(t, ArchX86_64, SystemDarwinX86_64.GetArch())
	assert.Equal(t, ArchAarch64, SystemDarwinAarch64.GetArch())
}

func TestArch_HumanReadable(t *testing.T) {
	tests := []struct {
		name string
		arch Arch
		want string
	}{
		{
			name: "ARM",
			arch: ArchAarch64,
			want: "ARM",
		},
		{
			name: "Intel",
			arch: ArchX86_64,
			want: "Intel",
		},
		{
			name: "Unknown",
			arch: Arch(999), // Invalid/unknown arch
			want: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.arch.HumanReadable()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestArch_String(t *testing.T) {
	tests := []struct {
		name string
		arch Arch
		want string
	}{
		{
			name: "aarch64",
			arch: ArchAarch64,
			want: "aarch64",
		},
		{
			name: "x86_64",
			arch: ArchX86_64,
			want: "x86_64",
		},
		{
			name: "unknown",
			arch: Arch(999), // Invalid/unknown arch
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.arch.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSystemConstants(t *testing.T) {
	// Verify constants are set correctly
	assert.Equal(t, OSLinux, SystemLinuxAarch64.os)
	assert.Equal(t, ArchAarch64, SystemLinuxAarch64.arch)

	assert.Equal(t, OSLinux, SystemLinuxX86_64.os)
	assert.Equal(t, ArchX86_64, SystemLinuxX86_64.arch)

	assert.Equal(t, OSDarwin, SystemDarwinX86_64.os)
	assert.Equal(t, ArchX86_64, SystemDarwinX86_64.arch)

	assert.Equal(t, OSDarwin, SystemDarwinAarch64.os)
	assert.Equal(t, ArchAarch64, SystemDarwinAarch64.arch)
}
