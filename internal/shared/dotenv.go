package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

// ParseDotEnv reads KEY=VALUE lines from r and returns them as a map.
//
// It is pure: it never reads or writes the process environment. Applying the
// result is LoadDotEnv's job.
//
// Supported syntax:
//
//	FOO=bar                 plain value
//	export FOO=bar          optional export prefix
//	  FOO = bar             whitespace around key and value is trimmed
//	# comment               full-line comment
//	FOO=bar # note          trailing comment (needs whitespace before '#')
//	FOO=""                  empty value
//	FOO='a\nb'              single quotes: literal, no escapes
//	FOO="a\nb"              double quotes: \n \t \r \\ \" are unescaped
//
// Later duplicates win. Not supported: multi-line values and $VAR expansion.
func ParseDotEnv(r io.Reader) (map[string]string, error) {
	values := make(map[string]string)

	sc := bufio.NewScanner(r)
	for lineNo := 1; sc.Scan(); lineNo++ {
		line := strings.TrimSuffix(sc.Text(), "\r")

		fmt.Println(line)

		if lineNo == 1 {
			line = strings.TrimPrefix(line, "\ufeff") // UTF-8 BOM
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = trimExportPrefix(line)

		key, raw, found := strings.Cut(line, "=")
		if !found {
			return nil, fmt.Errorf("line %d: missing '=' in %q", lineNo, line)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNo)
		}

		value, err := parseValue(raw)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}

		values[key] = value
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

// LoadDotEnv reads path and applies its values to the process environment.
//
// Variables already present in the environment are never overwritten, so a
// real deployment env always beats the file. A missing file is not an error:
// production simply has no .env.
func LoadDotEnv(path string) error {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	values, err := ParseDotEnv(f)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}

	for key, value := range values {
		if _, set := os.LookupEnv(key); set {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	return nil
}

// trimExportPrefix removes a leading "export " (or "export\t") if present.
func trimExportPrefix(line string) string {
	rest, found := strings.CutPrefix(line, "export")
	if !found {
		return line
	}
	if !strings.HasPrefix(rest, " ") && !strings.HasPrefix(rest, "\t") {
		// A key that merely starts with "export", e.g. "export=1".
		return line
	}
	return strings.TrimSpace(rest)
}

// parseValue turns the right-hand side of KEY=VALUE into its final value.
func parseValue(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "#") {
		return "", nil
	}

	switch raw[0] {
	case '"':
		return parseDoubleQuoted(raw[1:])
	case '\'':
		return parseSingleQuoted(raw[1:])
	default:
		return trimInlineComment(raw), nil
	}
}

func parseSingleQuoted(s string) (string, error) {
	end := strings.IndexByte(s, '\'')
	if end < 0 {
		return "", errors.New("unterminated single quote")
	}
	if err := checkTrailing(s[end+1:]); err != nil {
		return "", err
	}
	return s[:end], nil
}

func parseDoubleQuoted(s string) (string, error) {
	var b strings.Builder

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '"':
			if err := checkTrailing(s[i+1:]); err != nil {
				return "", err
			}
			return b.String(), nil
		case '\\':
			i++
			if i >= len(s) {
				return "", errors.New("unterminated double quote")
			}
			switch s[i] {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			default:
				b.WriteByte(s[i]) // unknown escape: keep the character
			}
		default:
			b.WriteByte(s[i])
		}
	}

	return "", errors.New("unterminated double quote")
}

// checkTrailing accepts whatever may follow a closing quote: nothing or a comment.
func checkTrailing(s string) error {
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "#") {
		return nil
	}
	return fmt.Errorf("unexpected %q after closing quote", s)
}

// trimInlineComment removes a trailing comment from an unquoted value.
// A '#' only starts a comment when preceded by whitespace, so "a#b" stays intact.
func trimInlineComment(s string) string {
	space := strings.IndexAny(s, " \t")
	if space < 0 {
		return s
	}
	if !strings.Contains(s[space:], "#") {
		return s
	}
	return strings.TrimSpace(s[:space])
}
