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

func ParseFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open env file: %w", err)
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		key, value, ok, err := parseLine(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("parse %s:%d: %w", path, lineNumber, err)
		}
		if ok {
			values[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan env file: %w", err)
	}
	return values, nil
}

func parseLine(line string) (key string, value string, ok bool, err error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", "", false, nil
	}
	trimmed = strings.TrimPrefix(trimmed, "export ")
	idx := strings.Index(trimmed, "=")
	if idx <= 0 {
		return "", "", false, nil
	}
	key = strings.TrimSpace(trimmed[:idx])
	if key == "" || strings.ContainsAny(key, " \t\"'") {
		return "", "", false, fmt.Errorf("invalid key %q", key)
	}
	value = strings.TrimSpace(trimmed[idx+1:])
	return key, unquoteValue(value), true, nil
}

func unquoteValue(value string) string {
	if len(value) < 2 {
		return value
	}
	quote := value[0]
	if (quote != '\'' && quote != '"') || value[len(value)-1] != quote {
		return value
	}
	inner := value[1 : len(value)-1]
	if quote == '\'' {
		return inner
	}
	inner = strings.ReplaceAll(inner, `\n`, "\n")
	inner = strings.ReplaceAll(inner, `\r`, "\r")
	inner = strings.ReplaceAll(inner, `\t`, "\t")
	inner = strings.ReplaceAll(inner, `\"`, `"`)
	inner = strings.ReplaceAll(inner, `\\`, `\`)
	return inner
}
