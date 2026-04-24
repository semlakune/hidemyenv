//go:build !darwin

package keychain

import "errors"

var errUnsupported = errors.New("keychain integration is only implemented on macOS")

func Available() bool {
	return false
}

func Store(account, password string) error {
	return errUnsupported
}

func Get(account string) (string, bool, error) {
	return "", false, errUnsupported
}

func Delete(account string) (bool, error) {
	return false, errUnsupported
}
