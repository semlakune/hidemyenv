package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckReportsPlaintextEnv(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".env"), "SECRET=value\n")
	mustWrite(t, filepath.Join(dir, ".env.hidemyenv"), "{}\n")
	mustWrite(t, filepath.Join(dir, ".gitignore"), ".env\n.env.*\n!.env.example\n!.env.safe\n")

	findings := Check(dir)
	if !containsFinding(findings, "plaintext env file exists: .env") {
		t.Fatalf("expected plaintext finding, got %#v", findings)
	}
}

func TestCheckRequiresExactGitignoreLines(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, ".env.hidemyenv"), "{}\n")
	mustWrite(t, filepath.Join(dir, ".gitignore"), "!.env.example\n!.env.safe\n")

	findings := Check(dir)
	if !containsFinding(findings, ".gitignore should include .env") {
		t.Fatalf("expected missing .env finding, got %#v", findings)
	}
	if !containsFinding(findings, ".gitignore should include .env.*") {
		t.Fatalf("expected missing .env.* finding, got %#v", findings)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func containsFinding(findings []Finding, want string) bool {
	for _, finding := range findings {
		if strings.Contains(finding.Message, want) {
			return true
		}
	}
	return false
}
