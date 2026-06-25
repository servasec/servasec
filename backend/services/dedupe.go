package services

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// DedupeHash returns a stable identity for a finding within an application.
// Two findings sharing the same scanner, rule, file, line and severity are
// treated as the same issue and deduplicated on ingest. lineStart is optional
// (e.g. URL-based DAST findings have no line) and a nil line is distinct from
// line 0.
func DedupeHash(scannerType, ruleID, filePath string, lineStart *int, severity string) string {
	line := ""
	if lineStart != nil {
		line = strconv.Itoa(*lineStart)
	}
	key := strings.Join([]string{scannerType, ruleID, filePath, line, severity}, "|")
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}
