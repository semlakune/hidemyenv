package runner

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
)

func Run(command []string, values map[string]string) error {
	if len(command) == 0 {
		return fmt.Errorf("missing command after --")
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = mergedEnv(os.Environ(), values)
	return cmd.Run()
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
