//go:build darwin

package keychain

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func Available() bool {
	_, err := exec.LookPath("security")
	return err == nil
}

func Store(account, password string) error {
	if account == "" {
		return errors.New("keychain account cannot be empty")
	}
	if password == "" {
		return errors.New("vault password cannot be empty")
	}
	cmd := exec.Command("security", "add-generic-password", "-U", "-s", Service, "-a", account, "-w", password)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("store keychain item: %s", sanitizeSecurityOutput(out))
	}
	return nil
}

func Get(account string) (string, bool, error) {
	if account == "" {
		return "", false, errors.New("keychain account cannot be empty")
	}
	cmd := exec.Command("security", "find-generic-password", "-w", "-s", Service, "-a", account)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := sanitizeSecurityOutput(out)
		if isNotFound(msg) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("read keychain item: %s", msg)
	}
	return strings.TrimRight(string(out), "\r\n"), true, nil
}

func Delete(account string) (bool, error) {
	if account == "" {
		return false, errors.New("keychain account cannot be empty")
	}
	cmd := exec.Command("security", "delete-generic-password", "-s", Service, "-a", account)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := sanitizeSecurityOutput(out)
		if isNotFound(msg) {
			return false, nil
		}
		return false, fmt.Errorf("delete keychain item: %s", msg)
	}
	return true, nil
}

func sanitizeSecurityOutput(out []byte) string {
	msg := strings.TrimSpace(string(bytes.TrimSpace(out)))
	if msg == "" {
		return "security command failed"
	}
	return msg
}

func isNotFound(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "could not be found") || strings.Contains(lower, "the specified item could not be found")
}
