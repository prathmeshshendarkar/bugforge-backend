package helpers

import (
	"regexp"
	"strings"
)

var usernameRegex = regexp.MustCompile(`^[a-z0-9._]{3,30}$`)

func ValidateUsername(u string) bool {
    return usernameRegex.MatchString(u)
}

// Converts name → username (e.g. “John Doe” → “john.doe”)
func GenerateUsername(name string) string {
    name = strings.ToLower(strings.TrimSpace(name))
    name = strings.ReplaceAll(name, " ", ".")
    name = strings.ReplaceAll(name, "_", ".")
    name = strings.ReplaceAll(name, "..", ".")
    return name
}
