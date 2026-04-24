package main

import (
	"os"
	"strings"
	"testing"
)

func TestInitProjectCreatesSafeDefaults(t *testing.T) {
	withTempDir(t)
	if err := initProject(); err != nil {
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
}

func TestSetRejectsInvalidKeyBeforePrompt(t *testing.T) {
	withTempDir(t)
	if err := setSecret([]string{"1INVALID"}); err == nil {
		t.Fatal("expected invalid key error")
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
