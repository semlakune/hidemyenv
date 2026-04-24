package runner

import "testing"

func TestMergedEnvVaultValuesOverrideBase(t *testing.T) {
	env := mergedEnv([]string{"DATABASE_URL=old", "APP_NAME=demo"}, map[string]string{
		"DATABASE_URL": "new",
		"JWT_SECRET":   "secret",
	})
	want := map[string]string{
		"DATABASE_URL": "new",
		"APP_NAME":     "demo",
		"JWT_SECRET":   "secret",
	}
	got := map[string]string{}
	for _, item := range env {
		for i := 0; i < len(item); i++ {
			if item[i] == '=' {
				got[item[:i]] = item[i+1:]
				break
			}
		}
	}
	for key, value := range want {
		if got[key] != value {
			t.Fatalf("%s=%q, want %q", key, got[key], value)
		}
	}
}
