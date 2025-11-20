package flake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAttr(t *testing.T) {
	attr := NewAttr("foo")
	assert.False(t, attr.IsNone())
	assert.Equal(t, "foo", attr.GetName())
	assert.Equal(t, "foo", attr.String())
}

func TestNoneAttr(t *testing.T) {
	attr := NoneAttr()
	assert.True(t, attr.IsNone())
	assert.Equal(t, "default", attr.GetName())
	assert.Equal(t, "", attr.String())
}

func TestAttr_GetName(t *testing.T) {
	tests := []struct {
		name string
		attr Attr
		want string
	}{
		{
			name: "simple attribute",
			attr: NewAttr("foo"),
			want: "foo",
		},
		{
			name: "nested attribute",
			attr: NewAttr("packages.x86_64-linux.hello"),
			want: "packages.x86_64-linux.hello",
		},
		{
			name: "no attribute returns default",
			attr: NoneAttr(),
			want: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.attr.GetName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAttr_IsNone(t *testing.T) {
	tests := []struct {
		name string
		attr Attr
		want bool
	}{
		{
			name: "with attribute",
			attr: NewAttr("foo"),
			want: false,
		},
		{
			name: "none attribute",
			attr: NoneAttr(),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.attr.IsNone()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAttr_AsList(t *testing.T) {
	tests := []struct {
		name string
		attr Attr
		want []string
	}{
		{
			name: "simple attribute",
			attr: NewAttr("foo"),
			want: []string{"foo"},
		},
		{
			name: "nested attribute with dots",
			attr: NewAttr("packages.x86_64-linux.hello"),
			want: []string{"packages", "x86_64-linux", "hello"},
		},
		{
			name: "deeply nested",
			attr: NewAttr("a.b.c.d.e"),
			want: []string{"a", "b", "c", "d", "e"},
		},
		{
			name: "no attribute returns empty list",
			attr: NoneAttr(),
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.attr.AsList()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAttr_String(t *testing.T) {
	tests := []struct {
		name string
		attr Attr
		want string
	}{
		{
			name: "simple attribute",
			attr: NewAttr("foo"),
			want: "foo",
		},
		{
			name: "nested attribute",
			attr: NewAttr("packages.x86_64-linux.hello"),
			want: "packages.x86_64-linux.hello",
		},
		{
			name: "no attribute returns empty string",
			attr: NoneAttr(),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.attr.String()
			assert.Equal(t, tt.want, got)
		})
	}
}
