package dotenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSafeRedactsVaultKeys(t *testing.T) {
	dir := t.TempDir()
	example := filepath.Join(dir, ".env.example")
	content := "# comment\nDATABASE_URL=\nNEXT_PUBLIC_URL=http://localhost:3000\n"
	if err := os.WriteFile(example, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	safe, err := GenerateSafe(example, []string{"DATABASE_URL", "JWT_SECRET"})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"DATABASE_URL=***REDACTED***",
		"NEXT_PUBLIC_URL=http://localhost:3000",
		"JWT_SECRET=***REDACTED***",
	} {
		if !strings.Contains(safe, want) {
			t.Fatalf("safe file missing %q in:\n%s", want, safe)
		}
	}
}

func TestKeyFromLine(t *testing.T) {
	tests := map[string]string{
		"KEY=value":        "KEY",
		" export FOO=bar ": "FOO",
		"# comment":        "",
		"invalid":          "",
	}
	for input, want := range tests {
		got, ok := KeyFromLine(input)
		if want == "" && ok {
			t.Fatalf("expected no key for %q", input)
		}
		if want != "" && got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}

func TestParseFileReadsDotenvValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := strings.Join([]string{
		"# comment",
		"DATABASE_URL=postgres://secret",
		"JWT_SECRET='single quoted secret'",
		`OPENAI_API_KEY="double quoted secret"`,
		"EMPTY=",
		"export EXPORTED=value",
	}, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	values, err := ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"DATABASE_URL":   "postgres://secret",
		"JWT_SECRET":     "single quoted secret",
		"OPENAI_API_KEY": "double quoted secret",
		"EMPTY":          "",
		"EXPORTED":       "value",
	}
	for key, value := range want {
		if values[key] != value {
			t.Fatalf("%s=%q, want %q", key, values[key], value)
		}
	}
}
