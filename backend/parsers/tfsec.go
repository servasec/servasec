package parsers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type tfsecOutput struct {
	Results tfsecResults `json:"results"`
}

type tfsecResults struct {
	Passed []tfsecResult `json:"passed"`
	Failed []tfsecResult `json:"failed"`
}

type tfsecResult struct {
	RuleID      string        `json:"rule_id"`
	Severity    string        `json:"severity"`
	Description string        `json:"description"`
	Resolution  string        `json:"resolution"`
	Location    tfsecLocation `json:"location"`
}

type tfsecLocation struct {
	Filename  string `json:"filename"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

func ParseTfsec(data []byte, filename string) ([]FindingInput, error) {
	var output tfsecOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("invalid tfsec JSON: %w", err)
	}

	var findings []FindingInput
	for _, r := range output.Results.Failed {
		severity := strings.ToLower(r.Severity)
		if err := validateSeverity(severity); err != nil {
			severity = "medium"
		}

		title := r.Description
		if title == "" {
			title = r.RuleID
		}
		if len(title) > 500 {
			title = title[:500]
		}

		description := r.Description
		if r.Resolution != "" {
			if description != "" {
				description = description + "\nResolution: " + r.Resolution
			} else {
				description = r.Resolution
			}
		}
		if len(description) > 2000 {
			description = description[:2000]
		}

		filePath := r.Location.Filename
		if filePath != "" && !filepath.IsAbs(filePath) {
			filePath = "/" + filePath
		}

		var lineStart, lineEnd *int
		if r.Location.StartLine > 0 {
			lineStart = &r.Location.StartLine
		}
		if r.Location.EndLine > 0 {
			lineEnd = &r.Location.EndLine
		}

		findings = append(findings, FindingInput{
			RuleID:      r.RuleID,
			Title:       title,
			Severity:    severity,
			Description: description,
			FilePath:    filePath,
			LineStart:   lineStart,
			LineEnd:     lineEnd,
			Remediation: r.Resolution,
		})
	}

	return findings, nil
}
