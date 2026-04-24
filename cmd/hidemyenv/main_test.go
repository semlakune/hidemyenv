package main

import (
	"os"
	"strings"
	"testing"

	"hidemyenv/internal/vault"
)

func TestInitProjectCreatesSafeDefaults(t *testing.T) {
	withTempDir(t)
	if err := initProject(nil); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{".hidemyenv.toml", ".env.example", ".env.safe", ".env.hidemyenv", ".gitignore"} {
		if _, err := os.Stat(name); err != nil {
			t.Fatalf("expected %s: %v", name, err)
		}
	}
	gitignore, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{".env", ".env.*", "!.env.example", "!.env.safe"} {
		if !strings.Contains(string(gitignore), want) {
			t.Fatalf(".gitignore missing %s", want)
		}
	}
	if _, err := os.Stat("justfile"); !os.IsNotExist(err) {
		t.Fatal("init without --scripts should not create justfile")
	}
}

func TestInitProjectWithScriptsCreatesJustfile(t *testing.T) {
	withTempDir(t)
	if err := os.WriteFile("main.py", []byte("print('hello')\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("uv.lock", []byte("version = 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := initProject([]string{"--scripts"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("justfile")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, want := range []string{
		"env-run *args:",
		"hidemyenv run -- {{args}}",
		"dev:",
		"hidemyenv run -- uv run main.py",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("justfile missing %q in:\n%s", want, content)
		}
	}
}

func TestInitProjectWithScriptsAppendsExistingJustfile(t *testing.T) {
	withTempDir(t)
	if err := os.WriteFile("Justfile", []byte("test:\n\tgo test ./...\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := initProject([]string{"--scripts"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("Justfile")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "test:\n\tgo test ./...") {
		t.Fatalf("existing justfile content was not preserved:\n%s", content)
	}
	if !strings.Contains(content, "env-run *args:") {
		t.Fatalf("env-run recipe was not appended:\n%s", content)
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() == "justfile" {
			t.Fatal("should append to existing Justfile instead of creating lowercase justfile")
		}
	}
}

func TestInitProjectWithScriptsDoesNotDuplicateExistingDevRecipe(t *testing.T) {
	withTempDir(t)
	if err := os.WriteFile("justfile", []byte("dev:\n\tgo run .\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("main.py", []byte("print('hello')\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("uv.lock", []byte("version = 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := initProject([]string{"--scripts"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("justfile")
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Count(content, "dev:") != 1 {
		t.Fatalf("expected existing dev recipe to be preserved without duplication:\n%s", content)
	}
	if !strings.Contains(content, "env-run *args:") {
		t.Fatalf("env-run recipe was not appended:\n%s", content)
	}
}

func TestInitProjectRejectsUnknownFlag(t *testing.T) {
	withTempDir(t)
	if err := initProject([]string{"--unknown"}); err == nil {
		t.Fatal("expected usage error")
	}
}

func TestSetRejectsInvalidKeyBeforePrompt(t *testing.T) {
	withTempDir(t)
	if err := setSecret([]string{"1INVALID"}); err == nil {
		t.Fatal("expected invalid key error")
	}
}

func TestImportEnvFileEncryptsValues(t *testing.T) {
	withTempDir(t)
	if err := os.WriteFile(".env", []byte("DATABASE_URL=postgres://secret\nJWT_SECRET=secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	count, err := importEnvFile(".env", "password")
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
	data, err := os.ReadFile(vault.DefaultPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "postgres://secret") || strings.Contains(string(data), "JWT_SECRET=secret") {
		t.Fatal("vault contains plaintext secret")
	}
	f, err := vault.Load(vault.DefaultPath)
	if err != nil {
		t.Fatal(err)
	}
	values, err := f.Values("password")
	if err != nil {
		t.Fatal(err)
	}
	if values["DATABASE_URL"] != "postgres://secret" || values["JWT_SECRET"] != "secret" {
		t.Fatalf("unexpected imported values: %#v", values)
	}
}

func TestContainsLine(t *testing.T) {
	content := "!.env.example\n!.env.safe\n"
	if containsLine(content, ".env") {
		t.Fatal("substring matched as line")
	}
	if !containsLine(content, "!.env.safe") {
		t.Fatal("expected exact line match")
	}
}

func TestKeychainCommandRejectsInvalidArgs(t *testing.T) {
	if err := keychainCommand(nil); err == nil {
		t.Fatal("expected usage error")
	}
	if err := keychainCommand([]string{"unknown"}); err == nil {
		t.Fatal("expected usage error")
	}
}

func withTempDir(t *testing.T) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(old); err != nil {
			t.Fatal(err)
		}
	})
}
