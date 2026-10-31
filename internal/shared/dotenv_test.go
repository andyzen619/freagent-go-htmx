package config

import (
	"maps"
	"os"
	"strings"
	"testing"
)

func TestParseDotEnv(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr string // substring of the expected error, "" means success
	}{
		{
			name:  "plain",
			input: "FOO=bar",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "export prefix",
			input: "export FOO=bar",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "export prefix with tab",
			input: "export\tFOO=bar",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "key named export",
			input: "export=1",
			want:  map[string]string{"export": "1"},
		},
		{
			name:  "whitespace around key and value",
			input: "  FOO =  bar  ",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "comments and blank lines",
			input: "# a comment\n\nFOO=bar\n   \n# another",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "trailing comment",
			input: "FOO=bar # note",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "trailing comment with tab",
			input: "FOO=bar\t# note",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "hash without preceding space is part of the value",
			input: "FOO=ab#c",
			want:  map[string]string{"FOO": "ab#c"},
		},
		{
			name:  "empty value",
			input: "FOO=",
			want:  map[string]string{"FOO": ""},
		},
		{
			name:  "value that is only a comment",
			input: "FOO= # nothing here",
			want:  map[string]string{"FOO": ""},
		},
		{
			name:  "double quoted empty",
			input: `FOO=""`,
			want:  map[string]string{"FOO": ""},
		},
		{
			name:  "single quotes are literal",
			input: `FOO='a\nb'`,
			want:  map[string]string{"FOO": `a\nb`},
		},
		{
			name:  "single quoted keeps hash",
			input: `FOO='a # b'`,
			want:  map[string]string{"FOO": "a # b"},
		},
		{
			name:  "double quoted escapes",
			input: `FOO="a\nb\tc\r\\\"d"`,
			want:  map[string]string{"FOO": "a\nb\tc\r\\\"d"},
		},
		{
			name:  "double quoted keeps inner spaces",
			input: `FOO="  padded  "`,
			want:  map[string]string{"FOO": "  padded  "},
		},
		{
			name:  "comment after closing quote",
			input: `FOO="bar" # note`,
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "duplicate later wins",
			input: "FOO=one\nFOO=two",
			want:  map[string]string{"FOO": "two"},
		},
		{
			name:  "CRLF line endings",
			input: "FOO=bar\r\nBAZ=qux\r\n",
			want:  map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name:  "UTF-8 BOM",
			input: "\ufeffFOO=bar",
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "unicode value",
			input: "FOO=héllo→",
			want:  map[string]string{"FOO": "héllo→"},
		},
		{
			name:    "missing equals",
			input:   "FOO=bar\nJUSTAKEY",
			wantErr: `line 2: missing '='`,
		},
		{
			name:    "empty key",
			input:   "=bar",
			wantErr: "line 1: empty key",
		},
		{
			name:    "unterminated double quote",
			input:   `FOO="bar`,
			wantErr: "line 1: unterminated double quote",
		},
		{
			name:    "unterminated single quote",
			input:   "FOO='bar",
			wantErr: "line 1: unterminated single quote",
		},
		{
			name:    "junk after closing quote",
			input:   `FOO="bar" baz`,
			wantErr: `line 1: unexpected "baz" after closing quote`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDotEnv(strings.NewReader(tt.input))

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (values %v)", tt.wantErr, got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !maps.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// ParseDotEnv must stay pure: reading a file must never leak into the process env.
func TestParseDotEnvDoesNotSetEnv(t *testing.T) {
	const key = "FREAGENT_PARSE_ONLY_TEST"

	if _, err := ParseDotEnv(strings.NewReader(key + "=leaked")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v, ok := os.LookupEnv(key); ok {
		t.Fatalf("ParseDotEnv modified the process environment: %s=%s", key, v)
	}
}
