package dotenv

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

const RedactedValue = "***REDACTED***"

func GenerateSafe(examplePath string, secretKeys []string) (string, error) {
	secretSet := map[string]bool{}
	for _, key := range secretKeys {
		secretSet[key] = true
	}
	seen := map[string]bool{}
	var out []string

	file, err := os.Open(examplePath)
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			key, ok := KeyFromLine(line)
			if !ok {
				out = append(out, line)
				continue
			}
			seen[key] = true
			if secretSet[key] {
				out = append(out, fmt.Sprintf("%s=%s", key, RedactedValue))
			} else {
				out = append(out, line)
			}
		}
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("scan example env: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("open example env: %w", err)
	}

	missing := make([]string, 0)
	for _, key := range secretKeys {
		if !seen[key] {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	for _, key := range missing {
		out = append(out, fmt.Sprintf("%s=%s", key, RedactedValue))
	}
	return strings.Join(out, "\n") + "\n", nil
}

func KeyFromLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", false
	}
	trimmed = strings.TrimPrefix(trimmed, "export ")
	idx := strings.Index(trimmed, "=")
	if idx <= 0 {
		return "", false
	}
	key := strings.TrimSpace(trimmed[:idx])
	if key == "" || strings.ContainsAny(key, " \t\"'") {
		return "", false
	}
	return key, true
}
