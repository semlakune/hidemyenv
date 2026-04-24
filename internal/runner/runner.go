package runner

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"hidemyenv/internal/dotenv"
)

func Run(command []string, values map[string]string) error {
	if len(command) == 0 {
		return fmt.Errorf("missing command after --")
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Env = mergedEnv(os.Environ(), values)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("capture stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("capture stderr: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	copyOutput := func(dst io.Writer, src io.Reader, label string) {
		defer wg.Done()
		if err := copyRedacted(dst, src, values); err != nil {
			errs <- fmt.Errorf("copy %s: %w", label, err)
		}
	}
	wg.Add(2)
	go copyOutput(os.Stdout, stdout, "stdout")
	go copyOutput(os.Stderr, stderr, "stderr")

	wg.Wait()
	waitErr := cmd.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			return err
		}
	}
	return waitErr
}

func mergedEnv(base []string, values map[string]string) []string {
	seen := map[string]bool{}
	env := make([]string, 0, len(base)+len(values))
	for _, item := range base {
		key := item
		for i := 0; i < len(item); i++ {
			if item[i] == '=' {
				key = item[:i]
				break
			}
		}
		if values[key] != "" || hasKey(values, key) {
			continue
		}
		seen[key] = true
		env = append(env, item)
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		if !seen[key] {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		env = append(env, key+"="+values[key])
	}
	return env
}

func hasKey(values map[string]string, key string) bool {
	_, ok := values[key]
	return ok
}

func copyRedacted(dst io.Writer, src io.Reader, values map[string]string) error {
	secrets := redactableValues(values)
	if len(secrets) == 0 {
		_, err := io.Copy(dst, src)
		return err
	}

	filter := newRedactor(dst, secrets)
	if _, err := io.Copy(filter, src); err != nil {
		return err
	}
	return filter.Flush()
}

func redactableValues(values map[string]string) []string {
	seen := map[string]bool{}
	secrets := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		secrets = append(secrets, value)
	}
	sort.Slice(secrets, func(i, j int) bool {
		return len(secrets[i]) > len(secrets[j])
	})
	return secrets
}

type redactor struct {
	dst     io.Writer
	secrets []string
	pending string
}

func newRedactor(dst io.Writer, secrets []string) *redactor {
	ordered := append([]string(nil), secrets...)
	sort.Slice(ordered, func(i, j int) bool {
		return len(ordered[i]) > len(ordered[j])
	})
	return &redactor{dst: dst, secrets: ordered}
}

func (r *redactor) Write(p []byte) (int, error) {
	for _, b := range p {
		r.pending += string(b)
		if err := r.drain(false); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (r *redactor) Flush() error {
	return r.drain(true)
}

func (r *redactor) drain(final bool) error {
	for r.pending != "" {
		if !final && r.pendingCanMatchSecret() {
			return nil
		}
		if matched, err := r.writeSecretMatch(); matched || err != nil {
			if err != nil {
				return err
			}
			continue
		}
		if _, err := io.WriteString(r.dst, r.pending[:1]); err != nil {
			return err
		}
		r.pending = r.pending[1:]
	}
	return nil
}

func (r *redactor) writeSecretMatch() (bool, error) {
	for _, secret := range r.secrets {
		if strings.HasPrefix(r.pending, secret) {
			_, err := io.WriteString(r.dst, dotenv.RedactedValue)
			r.pending = r.pending[len(secret):]
			return true, err
		}
	}
	return false, nil
}

func (r *redactor) pendingCanMatchSecret() bool {
	for _, secret := range r.secrets {
		if strings.HasPrefix(secret, r.pending) {
			return true
		}
	}
	return false
}
