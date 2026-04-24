package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hidemyenv/internal/cryptoutil"
)

func TestVaultSetValuesAndSaveLoad(t *testing.T) {
	f := testVault()
	if err := f.Set("DATABASE_URL", "postgres://secret", "password"); err != nil {
		t.Fatal(err)
	}
	values, err := f.Values("password")
	if err != nil {
		t.Fatal(err)
	}
	if values["DATABASE_URL"] != "postgres://secret" {
		t.Fatalf("unexpected value: %q", values["DATABASE_URL"])
	}

	path := filepath.Join(t.TempDir(), ".env.hidemyenv")
	if err := f.Save(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "postgres://secret") {
		t.Fatal("vault file contains plaintext secret")
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	loadedValues, err := loaded.Values("password")
	if err != nil {
		t.Fatal(err)
	}
	if loadedValues["DATABASE_URL"] != "postgres://secret" {
		t.Fatalf("unexpected loaded value: %q", loadedValues["DATABASE_URL"])
	}
}

func TestVaultWrongPasswordFails(t *testing.T) {
	f := testVault()
	if err := f.Set("JWT_SECRET", "secret", "password"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Values("wrong-password"); err == nil {
		t.Fatal("expected wrong password to fail")
	}
}

func testVault() *File {
	f := New()
	f.KDF = cryptoutil.DefaultKDFParams()
	f.KDF.Memory = 1024
	f.KDF.Iterations = 1
	f.KDF.Parallelism = 1
	return f
}
