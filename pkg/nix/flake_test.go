package nix

import (
	"testing"
)

func TestNewFlakeURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "simple path",
			url:  ".",
			want: ".",
		},
		{
			name: "github URL",
			url:  "github:juspay/omnix",
			want: "github:juspay/omnix",
		},
		{
			name: "path with attribute",
			url:  ".#packages.x86_64-linux.default",
			want: ".#packages.x86_64-linux.default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFlakeURL(tt.url)
			if got.String() != tt.want {
				t.Errorf("NewFlakeURL().String() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestFlakeURL_AsLocalPath(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "current directory",
			url:  ".",
			want: ".",
		},
		{
			name: "relative path",
			url:  "./subdir",
			want: "./subdir",
		},
		{
			name: "absolute path",
			url:  "/home/user/project",
			want: "/home/user/project",
		},
		{
			name: "path with attribute",
			url:  ".#packages.x86_64-linux.default",
			want: ".",
		},
		{
			name: "path with query",
			url:  ".?dir=subdir",
			want: ".",
		},
		{
			name: "path: prefix",
			url:  "path:.",
			want: ".",
		},
		{
			name: "github URL",
			url:  "github:juspay/omnix",
			want: "",
		},
		{
			name: "nixpkgs",
			url:  "nixpkgs",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFlakeURL(tt.url)
			if got := f.AsLocalPath(); got != tt.want {
				t.Errorf("FlakeURL.AsLocalPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlakeURL_IsLocal(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "current directory",
			url:  ".",
			want: true,
		},
		{
			name: "relative path",
			url:  "./subdir",
			want: true,
		},
		{
			name: "absolute path",
			url:  "/home/user/project",
			want: true,
		},
		{
			name: "github URL",
			url:  "github:juspay/omnix",
			want: false,
		},
		{
			name: "nixpkgs",
			url:  "nixpkgs",
			want: false,
		},
		{
			name: "path with attribute",
			url:  ".#default",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFlakeURL(tt.url)
			if got := f.IsLocal(); got != tt.want {
				t.Errorf("FlakeURL.IsLocal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlakeURL_WithAttr(t *testing.T) {
	tests := []struct {
		name string
		url  string
		attr string
		want string
	}{
		{
			name: "add attribute to plain URL",
			url:  ".",
			attr: "packages.x86_64-linux.default",
			want: ".#packages.x86_64-linux.default",
		},
		{
			name: "replace existing attribute",
			url:  ".#old",
			attr: "new",
			want: ".#new",
		},
		{
			name: "remove attribute",
			url:  ".#old",
			attr: "",
			want: ".",
		},
		{
			name: "github URL with attribute",
			url:  "github:juspay/omnix",
			attr: "packages.x86_64-linux.default",
			want: "github:juspay/omnix#packages.x86_64-linux.default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFlakeURL(tt.url)
			got := f.WithAttr(tt.attr)
			if got.String() != tt.want {
				t.Errorf("FlakeURL.WithAttr() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestFlakeURL_SplitAttr(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantBase string
		wantAttr string
	}{
		{
			name:     "no attribute",
			url:      ".",
			wantBase: ".",
			wantAttr: "",
		},
		{
			name:     "with attribute",
			url:      ".#packages.x86_64-linux.default",
			wantBase: ".",
			wantAttr: "packages.x86_64-linux.default",
		},
		{
			name:     "github with attribute",
			url:      "github:juspay/omnix#default",
			wantBase: "github:juspay/omnix",
			wantAttr: "default",
		},
		{
			name:     "empty attribute",
			url:      ".#",
			wantBase: ".",
			wantAttr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFlakeURL(tt.url)
			gotBase, gotAttr := f.SplitAttr()
			if gotBase != tt.wantBase {
				t.Errorf("FlakeURL.SplitAttr() base = %v, want %v", gotBase, tt.wantBase)
			}
			if gotAttr != tt.wantAttr {
				t.Errorf("FlakeURL.SplitAttr() attr = %v, want %v", gotAttr, tt.wantAttr)
			}
		})
	}
}

func TestFlakeURL_Clean(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "clean relative path",
			url:  "./././subdir",
			want: "subdir",
		},
		{
			name: "clean with attribute",
			url:  "./././subdir#attr",
			want: "subdir#attr",
		},
		{
			name: "non-local URL unchanged",
			url:  "github:juspay/omnix",
			want: "github:juspay/omnix",
		},
		{
			name: "absolute path",
			url:  "/home/user/././project",
			want: "/home/user/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFlakeURL(tt.url)
			got := f.Clean()
			if got.String() != tt.want {
				t.Errorf("FlakeURL.Clean() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestParseFlakeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid local path",
			input:   ".",
			want:    ".",
			wantErr: false,
		},
		{
			name:    "valid github URL",
			input:   "github:juspay/omnix",
			want:    "github:juspay/omnix",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFlakeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlakeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.String() != tt.want {
				t.Errorf("ParseFlakeURL() = %v, want %v", got.String(), tt.want)
			}
		})
	}
}
