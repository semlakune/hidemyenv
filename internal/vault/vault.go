package vault

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"

	"hidemyenv/internal/cryptoutil"
)

const (
	DefaultPath = ".env.hidemyenv"
	Version     = 1
)

type File struct {
	Version   int                       `json:"version"`
	Algorithm string                    `json:"algorithm"`
	KDF       cryptoutil.KDFParams      `json:"kdf"`
	Entries   map[string]EncryptedEntry `json:"entries"`
}

type EncryptedEntry struct {
	Nonce      string `json:"nonce"`
	Ciphertext string `json:"ciphertext"`
}

func New() *File {
	return &File{
		Version:   Version,
		Algorithm: cryptoutil.AlgorithmXChaCha20Poly1305,
		KDF:       cryptoutil.DefaultKDFParams(),
		Entries:   map[string]EncryptedEntry{},
	}
}

func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return New(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read vault: %w", err)
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse vault: %w", err)
	}
	if f.Version != Version {
		return nil, fmt.Errorf("unsupported vault version: %d", f.Version)
	}
	if f.Algorithm != cryptoutil.AlgorithmXChaCha20Poly1305 {
		return nil, fmt.Errorf("unsupported vault algorithm: %s", f.Algorithm)
	}
	if f.Entries == nil {
		f.Entries = map[string]EncryptedEntry{}
	}
	return &f, nil
}

func (f *File) Save(path string) error {
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return fmt.Errorf("encode vault: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write vault: %w", err)
	}
	return nil
}

func (f *File) Keys() []string {
	keys := make([]string, 0, len(f.Entries))
	for key := range f.Entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (f *File) Set(name, value, password string) error {
	if !ValidKey(name) {
		return fmt.Errorf("invalid env key: %s", name)
	}
	key, err := f.deriveKey(password)
	if err != nil {
		return err
	}
	if len(f.Entries) > 0 {
		if _, err := f.Values(password); err != nil {
			return err
		}
	}
	nonce, ciphertext, err := cryptoutil.Encrypt(key, []byte(value), additionalData(name))
	if err != nil {
		return err
	}
	f.Entries[name] = EncryptedEntry{
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}
	return nil
}

func (f *File) Values(password string) (map[string]string, error) {
	key, err := f.deriveKey(password)
	if err != nil {
		return nil, err
	}
	values := make(map[string]string, len(f.Entries))
	for name, entry := range f.Entries {
		nonce, err := base64.StdEncoding.DecodeString(entry.Nonce)
		if err != nil {
			return nil, fmt.Errorf("decode nonce for %s: %w", name, err)
		}
		ciphertext, err := base64.StdEncoding.DecodeString(entry.Ciphertext)
		if err != nil {
			return nil, fmt.Errorf("decode ciphertext for %s: %w", name, err)
		}
		plaintext, err := cryptoutil.Decrypt(key, nonce, ciphertext, additionalData(name))
		if err != nil {
			return nil, fmt.Errorf("unlock vault: %w", err)
		}
		values[name] = string(plaintext)
	}
	return values, nil
}

func (f *File) deriveKey(password string) ([]byte, error) {
	if f.KDF.Salt == "" {
		salt, err := cryptoutil.RandomBytes(16)
		if err != nil {
			return nil, err
		}
		f.KDF = cryptoutil.DefaultKDFParams()
		f.KDF.Salt = base64.StdEncoding.EncodeToString(salt)
	}
	salt, err := base64.StdEncoding.DecodeString(f.KDF.Salt)
	if err != nil {
		return nil, fmt.Errorf("decode kdf salt: %w", err)
	}
	return cryptoutil.DeriveKey(password, salt, f.KDF)
}

func additionalData(name string) []byte {
	return []byte(fmt.Sprintf("hidemyenv:v%d:%s", Version, name))
}

func ValidKey(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if r == '_' || ('A' <= r && r <= 'Z') || ('a' <= r && r <= 'z') || (i > 0 && '0' <= r && r <= '9') {
			continue
		}
		return false
	}
	return true
}
