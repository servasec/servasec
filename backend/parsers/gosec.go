package parsers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

type gosecOutput struct {
	Issues []gosecIssue `json:"Issues"`
}

type gosecIssue struct {
	RuleID     string      `json:"rule_id"`
	Details    string      `json:"details"`
	Severity   string      `json:"severity"`
	Confidence string      `json:"confidence"`
	File       string      `json:"file"`
	Line       string      `json:"line"`
	Code       string      `json:"code"`
	Cwe        *gosecCwe   `json:"cwe"`
}

type gosecCwe struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
}

func ParseGosec(data []byte, filename string) ([]FindingInput, error) {
	var output gosecOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("invalid gosec JSON: %w", err)
	}

	var findings []FindingInput
	for _, issue := range output.Issues {
		severity := strings.ToLower(issue.Severity)
		if err := validateSeverity(severity); err != nil {
			severity = "medium"
		}

		title := issue.Details
		if title == "" {
			title = issue.RuleID
		}
		if len(title) > 500 {
			title = title[:500]
		}

		filePath := issue.File
		if filePath != "" && !filepath.IsAbs(filePath) {
			filePath = "/" + filePath
		}

		var lineStart *int
		if issue.Line != "" {
			if n, err := strconv.Atoi(issue.Line); err == nil && n > 0 {
				lineStart = &n
			}
		}

		cweID := ""
		if issue.Cwe != nil && issue.Cwe.ID != "" {
			cweID = issue.Cwe.ID
		}

		findings = append(findings, FindingInput{
			RuleID:      issue.RuleID,
			Title:       title,
			Severity:    severity,
			Description: issue.Details,
			FilePath:    filePath,
			LineStart:   lineStart,
			CWEID:       cweID,
		})
	}

	return findings, nil
}
