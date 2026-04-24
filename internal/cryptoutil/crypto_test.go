package cryptoutil

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	params := DefaultKDFParams()
	params.Memory = 1024
	params.Iterations = 1
	params.Parallelism = 1
	salt := []byte("1234567890abcdef")
	key, err := DeriveKey("password", salt, params)
	if err != nil {
		t.Fatal(err)
	}
	nonce, ciphertext, err := Encrypt(key, []byte("super-secret"), []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(ciphertext, []byte("super-secret")) {
		t.Fatal("ciphertext contains plaintext")
	}
	plaintext, err := Decrypt(key, nonce, ciphertext, []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	if string(plaintext) != "super-secret" {
		t.Fatalf("got %q", plaintext)
	}
}

func TestDecryptRejectsTampering(t *testing.T) {
	params := DefaultKDFParams()
	params.Memory = 1024
	params.Iterations = 1
	params.Parallelism = 1
	key, err := DeriveKey("password", []byte("1234567890abcdef"), params)
	if err != nil {
		t.Fatal(err)
	}
	nonce, ciphertext, err := Encrypt(key, []byte("super-secret"), []byte("aad"))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext[0] ^= 0xff
	if _, err := Decrypt(key, nonce, ciphertext, []byte("aad")); err == nil {
		t.Fatal("expected decrypt failure")
	}
}
