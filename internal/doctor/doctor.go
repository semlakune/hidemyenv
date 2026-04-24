package doctor

import (
	"fmt"
	"os"
	"strings"

	"hidemyenv/internal/vault"
)

type Finding struct {
	Severity string
	Message  string
}

func Check(workdir string) []Finding {
	var findings []Finding
	for _, name := range []string{".env", ".env.local", ".env.decrypted", ".env.plain"} {
		if exists(workdir + string(os.PathSeparator) + name) {
			findings = append(findings, Finding{"high", fmt.Sprintf("plaintext env file exists: %s", name)})
		}
	}
	if !exists(workdir + string(os.PathSeparator) + vault.DefaultPath) {
		findings = append(findings, Finding{"medium", "encrypted vault is missing"})
	}
	gitignore := workdir + string(os.PathSeparator) + ".gitignore"
	data, err := os.ReadFile(gitignore)
	if err != nil {
		findings = append(findings, Finding{"medium", ".gitignore is missing or unreadable"})
		return findings
	}
	content := string(data)
	for _, pattern := range []string{".env", ".env.*", "!.env.example", "!.env.safe"} {
		if !hasLine(content, pattern) {
			findings = append(findings, Finding{"medium", fmt.Sprintf(".gitignore should include %s", pattern)})
		}
	}
	return findings
}

func hasLine(content, want string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
