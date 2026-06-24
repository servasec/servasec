package parsers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type gitleaksResult struct {
	Description string   `json:"Description"`
	StartLine   int      `json:"StartLine"`
	EndLine     int      `json:"EndLine"`
	File        string   `json:"File"`
	RuleID      string   `json:"RuleID"`
	Match       string   `json:"Match"`
	Secret      string   `json:"Secret"`
	Tags        []string `json:"Tags"`
	Message     string   `json:"Message"`
	Fingerprint string   `json:"Fingerprint"`
}

func ParseGitleaks(data []byte, filename string) ([]FindingInput, error) {
	var results []gitleaksResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("invalid gitleaks JSON: %w", err)
	}

	var findings []FindingInput
	for _, r := range results {
		severity := "high"
		tags := strings.ToLower(strings.Join(r.Tags, " "))
		if strings.Contains(tags, "aws") || strings.Contains(tags, "gcp") ||
			strings.Contains(tags, "google") || strings.Contains(tags, "token") ||
			strings.Contains(tags, "password") || strings.Contains(tags, "secret") ||
			strings.Contains(tags, "private-key") || strings.Contains(tags, "credential") {
			severity = "critical"
		}

		title := r.RuleID
		if r.Description != "" {
			title = r.Description
			if len(title) > 500 {
				title = title[:500]
			}
		}

		filePath := r.File
		if !filepath.IsAbs(filePath) {
			filePath = "/" + filePath
		}

		lineStart := r.StartLine
		lineEnd := r.EndLine

		description := r.Description
		if r.Message != "" {
			description = description + ": " + r.Message
			if len(description) > 2000 {
				description = description[:2000]
			}
		}

		findings = append(findings, FindingInput{
			RuleID:      r.RuleID,
			Title:       title,
			Severity:    severity,
			Description: description,
			FilePath:    filePath,
			LineStart:   &lineStart,
			LineEnd:     &lineEnd,
		})
	}

	return findings, nil
}
