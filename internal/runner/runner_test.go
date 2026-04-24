package runner

import (
	"bytes"
	"strings"
	"testing"

	"hidemyenv/internal/dotenv"
)

func TestMergedEnvVaultValuesOverrideBase(t *testing.T) {
	env := mergedEnv([]string{"DATABASE_URL=old", "APP_NAME=demo"}, map[string]string{
		"DATABASE_URL": "new",
		"JWT_SECRET":   "secret",
	})
	want := map[string]string{
		"DATABASE_URL": "new",
		"APP_NAME":     "demo",
		"JWT_SECRET":   "secret",
	}
	got := map[string]string{}
	for _, item := range env {
		for i := 0; i < len(item); i++ {
			if item[i] == '=' {
				got[item[:i]] = item[i+1:]
				break
			}
		}
	}
	for key, value := range want {
		if got[key] != value {
			t.Fatalf("%s=%q, want %q", key, got[key], value)
		}
	}
}

func TestCopyRedactedMasksSecretOutput(t *testing.T) {
	var out bytes.Buffer
	err := copyRedacted(&out, strings.NewReader("token=sk-secret\n"), map[string]string{
		"OPENAI_API_KEY": "sk-secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "token=" + dotenv.RedactedValue + "\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestRedactorMasksSecretAcrossWrites(t *testing.T) {
	var out bytes.Buffer
	filter := newRedactor(&out, []string{"sk-secret"})
	if _, err := filter.Write([]byte("token=sk-")); err != nil {
		t.Fatal(err)
	}
	if _, err := filter.Write([]byte("secret\n")); err != nil {
		t.Fatal(err)
	}
	if err := filter.Flush(); err != nil {
		t.Fatal(err)
	}
	want := "token=" + dotenv.RedactedValue + "\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}

func TestRedactorPrefersLongerSecret(t *testing.T) {
	var out bytes.Buffer
	filter := newRedactor(&out, []string{"abc", "abcdef"})
	if _, err := filter.Write([]byte("abcdef\n")); err != nil {
		t.Fatal(err)
	}
	if err := filter.Flush(); err != nil {
		t.Fatal(err)
	}
	want := dotenv.RedactedValue + "\n"
	if out.String() != want {
		t.Fatalf("output = %q, want %q", out.String(), want)
	}
}
