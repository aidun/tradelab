package logging

import (
	"strings"
	"testing"
)

func TestRedactStringMasksBearerTokensAndPasswords(t *testing.T) {
	input := "Authorization: Bearer abc.def password=secret postgres://user:supersecret@localhost:5432/db"
	redacted := RedactString(input)

	if redacted == input {
		t.Fatalf("expected string to be redacted")
	}

	if containsAny(redacted, []string{"abc.def", "secret", "supersecret"}) {
		t.Fatalf("expected sensitive values to be removed, got %s", redacted)
	}
}

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
