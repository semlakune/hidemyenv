package keychain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

const Service = "hidemyenv"

func ProjectAccount(workdir string) (string, error) {
	if workdir == "" {
		var err error
		workdir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working directory: %w", err)
		}
	}
	abs, err := filepath.Abs(workdir)
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	sum := sha256.Sum256([]byte(abs))
	return hex.EncodeToString(sum[:]), nil
}
