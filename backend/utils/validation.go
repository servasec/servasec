package utils

import (
	"regexp"
	"unicode"
)

func ValidateUsername(username string) string {
	if len(username) == 0 {
		return "username_required"
	}
	if len(username) > 32 {
		return "username_too_long"
	}

	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(username) {
		return "username_invalid_characters"
	}

	for _, r := range username {
		if unicode.Is(unicode.Cc, r) ||
			unicode.Is(unicode.Cf, r) ||
			unicode.Is(unicode.Co, r) ||
			unicode.Is(unicode.Cs, r) ||
			unicode.Is(unicode.Zl, r) ||
			unicode.Is(unicode.Zp, r) {
			return "username_invalid_characters"
		}

		if r == '\u200B' || r == '\u200C' || r == '\u200D' || r == '\uFEFF' {
			return "username_invalid_characters"
		}
	}

	return ""
}
