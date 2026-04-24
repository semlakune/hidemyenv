package keychain

import "testing"

func TestProjectAccountIsStableAndOpaque(t *testing.T) {
	a, err := ProjectAccount("/tmp/project")
	if err != nil {
		t.Fatal(err)
	}
	b, err := ProjectAccount("/tmp/project")
	if err != nil {
		t.Fatal(err)
	}
	c, err := ProjectAccount("/tmp/other")
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("expected stable account")
	}
	if a == c {
		t.Fatal("expected different projects to have different accounts")
	}
	if len(a) != 64 {
		t.Fatalf("account length = %d, want 64", len(a))
	}
}
