package utils

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
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

func IsPrivateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 10 ||
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) ||
			(ip4[0] == 192 && ip4[1] == 168) ||
			ip4[0] == 127 ||
			(ip4[0] == 169 && ip4[1] == 254)
	}
	return ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

func ValidateRepositoryURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}

	if u.Host == "" {
		return fmt.Errorf("URL must have a valid host")
	}

	hostname := u.Hostname()

	blocked := []string{
		"localhost",
		"127.0.0.1",
		"::1",
		"metadata.google.internal",
		"169.254.169.254",
		"instance-data.internal",
		"169.254.169.254.xip.io",
	}
	for _, b := range blocked {
		if strings.EqualFold(hostname, b) {
			return fmt.Errorf("URL cannot point to internal/loopback addresses")
		}
	}

	ips, err := net.LookupIP(hostname)
	if err == nil {
		for _, ip := range ips {
			if IsPrivateIP(ip) {
				return fmt.Errorf("URL cannot resolve to private/internal networks")
			}
		}
	}

	return nil
}
