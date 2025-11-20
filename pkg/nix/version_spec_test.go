package nix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVersionSpec(t *testing.T) {
	tests := []struct {
		name    string
		op      VersionSpecType
		version Version
		wantOp  string
	}{
		{
			name:    "greater than",
			op:      VersionSpecGt,
			version: Version{Major: 2, Minor: 8, Patch: 0},
			wantOp:  ">",
		},
		{
			name:    "greater than or equal",
			op:      VersionSpecGte,
			version: Version{Major: 2, Minor: 8, Patch: 1},
			wantOp:  ">=",
		},
		{
			name:    "less than",
			op:      VersionSpecLt,
			version: Version{Major: 3, Minor: 0, Patch: 0},
			wantOp:  "<",
		},
		{
			name:    "less than or equal",
			op:      VersionSpecLte,
			version: Version{Major: 2, Minor: 14, Patch: 0},
			wantOp:  "<=",
		},
		{
			name:    "not equal",
			op:      VersionSpecNeq,
			version: Version{Major: 2, Minor: 9, Patch: 0},
			wantOp:  "!=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := NewVersionSpec(tt.op, tt.version)
			require.NotNil(t, spec)
			assert.Equal(t, tt.wantOp, spec.operator)
			assert.Equal(t, tt.version, spec.version)
		})
	}
}

func TestParseVersionSpec(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOp  string
		wantVer Version
		wantErr bool
	}{
		{
			name:    "greater than with minor",
			input:   ">2.8",
			wantOp:  ">",
			wantVer: Version{Major: 2, Minor: 8, Patch: 0},
		},
		{
			name:    "greater than or equal with patch",
			input:   ">=2.8.1",
			wantOp:  ">=",
			wantVer: Version{Major: 2, Minor: 8, Patch: 1},
		},
		{
			name:    "less than major only",
			input:   "<3",
			wantOp:  "<",
			wantVer: Version{Major: 3, Minor: 0, Patch: 0},
		},
		{
			name:    "less than or equal",
			input:   "<=2.14.0",
			wantOp:  "<=",
			wantVer: Version{Major: 2, Minor: 14, Patch: 0},
		},
		{
			name:    "not equal",
			input:   "!=2.9",
			wantOp:  "!=",
			wantVer: Version{Major: 2, Minor: 9, Patch: 0},
		},
		{
			name:    "invalid format - no operator",
			input:   "2.8",
			wantErr: true,
		},
		{
			name:    "invalid format - bad operator",
			input:   "~2.8",
			wantErr: true,
		},
		{
			name:    "invalid format - non-numeric",
			input:   ">=2.x.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersionSpec(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantOp, got.operator)
			assert.Equal(t, tt.wantVer, got.version)
		})
	}
}

func TestVersionSpec_Matches(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		version string
		want    bool
	}{
		// Greater than
		{
			name:    "greater than - true",
			spec:    ">2.8",
			version: "2.9.0",
			want:    true,
		},
		{
			name:    "greater than - false (equal)",
			spec:    ">2.8",
			version: "2.8.0",
			want:    false,
		},
		{
			name:    "greater than - false (less)",
			spec:    ">2.8",
			version: "2.7.0",
			want:    false,
		},
		// Greater than or equal
		{
			name:    "greater than or equal - true (greater)",
			spec:    ">=2.8",
			version: "2.9.0",
			want:    true,
		},
		{
			name:    "greater than or equal - true (equal)",
			spec:    ">=2.8",
			version: "2.8.0",
			want:    true,
		},
		{
			name:    "greater than or equal - false",
			spec:    ">=2.8",
			version: "2.7.0",
			want:    false,
		},
		// Less than
		{
			name:    "less than - true",
			spec:    "<3.0",
			version: "2.18.0",
			want:    true,
		},
		{
			name:    "less than - false (equal)",
			spec:    "<3.0",
			version: "3.0.0",
			want:    false,
		},
		{
			name:    "less than - false (greater)",
			spec:    "<3.0",
			version: "3.1.0",
			want:    false,
		},
		// Less than or equal
		{
			name:    "less than or equal - true (less)",
			spec:    "<=2.14",
			version: "2.13.0",
			want:    true,
		},
		{
			name:    "less than or equal - true (equal)",
			spec:    "<=2.14",
			version: "2.14.0",
			want:    true,
		},
		{
			name:    "less than or equal - false",
			spec:    "<=2.14",
			version: "2.15.0",
			want:    false,
		},
		// Not equal
		{
			name:    "not equal - true",
			spec:    "!=2.9",
			version: "2.9.1",
			want:    true,
		},
		{
			name:    "not equal - false",
			spec:    "!=2.9",
			version: "2.9.0",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := ParseVersionSpec(tt.spec)
			require.NoError(t, err)

			version, err := ParseVersion(tt.version)
			require.NoError(t, err)

			got := spec.Matches(version)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseVersionReq(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantSpecs int
		wantErr   bool
	}{
		{
			name:      "single spec",
			input:     ">=2.8",
			wantSpecs: 1,
		},
		{
			name:      "multiple specs",
			input:     ">=2.8, <2.14, !=2.13.4",
			wantSpecs: 3,
		},
		{
			name:      "multiple specs with spaces",
			input:     " >=2.8 ,  <2.14  ",
			wantSpecs: 2,
		},
		{
			name:    "invalid spec in list",
			input:   ">=2.8, invalid, <3.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersionReq(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, got.Specs, tt.wantSpecs)
		})
	}
}

func TestVersionReq_Matches(t *testing.T) {
	tests := []struct {
		name    string
		req     string
		version string
		want    bool
	}{
		{
			name:    "satisfies all specs",
			req:     ">=2.8, <2.14, !=2.9",
			version: "2.10.0",
			want:    true,
		},
		{
			name:    "fails first spec",
			req:     ">=2.8, <2.14, !=2.9",
			version: "2.7.0",
			want:    false,
		},
		{
			name:    "fails second spec",
			req:     ">=2.8, <2.14, !=2.9",
			version: "2.15.0",
			want:    false,
		},
		{
			name:    "fails third spec (not equal)",
			req:     ">=2.8, <2.14, !=2.9",
			version: "2.9.0",
			want:    false,
		},
		{
			name:    "edge case - equals lower bound",
			req:     ">=2.8, <3.0",
			version: "2.8.0",
			want:    true,
		},
		{
			name:    "edge case - equals upper bound",
			req:     ">=2.8, <3.0",
			version: "3.0.0",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseVersionReq(tt.req)
			require.NoError(t, err)

			version, err := ParseVersion(tt.version)
			require.NoError(t, err)

			got := req.Matches(version)
			assert.Equal(t, tt.want, got, "VersionReq(%s).Matches(%s)", tt.req, tt.version)
		})
	}
}

func TestVersionSpec_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "greater than",
			input: ">2.8",
			want:  ">2.8.0",
		},
		{
			name:  "greater than or equal",
			input: ">=2.8.1",
			want:  ">=2.8.1",
		},
		{
			name:  "less than",
			input: "<3.0",
			want:  "<3.0.0",
		},
		{
			name:  "not equal",
			input: "!=2.9.0",
			want:  "!=2.9.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := ParseVersionSpec(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, spec.String())
		})
	}
}

func TestVersionReq_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single spec",
			input: ">=2.8",
			want:  ">=2.8.0",
		},
		{
			name:  "multiple specs",
			input: ">=2.8, <2.14, !=2.9",
			want:  ">=2.8.0, <2.14.0, !=2.9.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseVersionReq(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, req.String())
		})
	}
}
