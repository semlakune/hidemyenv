package cryptoutil

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	AlgorithmXChaCha20Poly1305 = "XCHACHA20-POLY1305"
	KDFArgon2id                = "ARGON2ID"
)

type KDFParams struct {
	Name        string `json:"name"`
	Memory      uint32 `json:"memory_kib"`
	Iterations  uint32 `json:"iterations"`
	Parallelism uint8  `json:"parallelism"`
	KeyBytes    uint32 `json:"key_bytes"`
	Salt        string `json:"salt"`
}

func DefaultKDFParams() KDFParams {
	return KDFParams{
		Name:        KDFArgon2id,
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 4,
		KeyBytes:    chacha20poly1305.KeySize,
	}
}

func RandomBytes(size int) ([]byte, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("read random bytes: %w", err)
	}
	return buf, nil
}

func DeriveKey(password string, salt []byte, params KDFParams) ([]byte, error) {
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}
	if params.Name != KDFArgon2id {
		return nil, fmt.Errorf("unsupported kdf: %s", params.Name)
	}
	if params.KeyBytes == 0 {
		return nil, errors.New("kdf key size cannot be zero")
	}
	return argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyBytes), nil
}

func Encrypt(key, plaintext, additionalData []byte) (nonce []byte, ciphertext []byte, err error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, nil, fmt.Errorf("create cipher: %w", err)
	}
	nonce, err = RandomBytes(chacha20poly1305.NonceSizeX)
	if err != nil {
		return nil, nil, err
	}
	return nonce, aead.Seal(nil, nonce, plaintext, additionalData), nil
}

func Decrypt(key, nonce, ciphertext, additionalData []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, additionalData)
	if err != nil {
		return nil, errors.New("decrypt failed")
	}
	return plaintext, nil
}
