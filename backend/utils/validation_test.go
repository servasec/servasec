package utils

import (
	"strings"
	"testing"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{"empty", "", "username_required"},
		{"too long", strings.Repeat("a", 33), "username_too_long"},
		{"valid alphanumeric", "user123", ""},
		{"valid with underscores", "valid_user_123", ""},
		{"valid with hyphens", "valid-user-123", ""},
		{"valid single char", "u", ""},
		{"valid max length", strings.Repeat("a", 32), ""},
		{"special chars", "user@name", "username_invalid_characters"},
		{"space in middle", "user name", "username_invalid_characters"},
		{"unicode chars", "usér", "username_invalid_characters"},
		{"control char", "user\x00name", "username_invalid_characters"},
		{"zero-width space", "user\u200Bname", "username_invalid_characters"},
		{"zero-width non-joiner", "user\u200Cname", "username_invalid_characters"},
		{"zero-width joiner", "user\u200Dname", "username_invalid_characters"},
		{"BOM", "user\uFEFFname", "username_invalid_characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateUsername(tt.username)
			if got != tt.want {
				t.Errorf("ValidateUsername(%q) = %q, want %q", tt.username, got, tt.want)
			}
		})
	}
}
