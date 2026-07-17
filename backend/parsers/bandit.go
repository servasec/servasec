package parsers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type banditOutput struct {
	Results []banditResult `json:"results"`
}

type banditResult struct {
	TestID     string     `json:"test_id"`
	TestName   string     `json:"test_name"`
	IssueText  string     `json:"issue_text"`
	IssueSeverity string  `json:"issue_severity"`
	IssueConfidence string `json:"issue_confidence"`
	FileName   string     `json:"filename"`
	LineNumber int        `json:"line_number"`
	IssueCwe   *banditCwe `json:"issue_cwe"`
}

type banditCwe struct {
	ID   json.RawMessage `json:"id"`
	Link string          `json:"link"`
}

func ParseBandit(data []byte, filename string) ([]FindingInput, error) {
	var output banditOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("invalid bandit JSON: %w", err)
	}

	var findings []FindingInput
	for _, r := range output.Results {
		severity := strings.ToLower(r.IssueSeverity)
		if err := validateSeverity(severity); err != nil {
			severity = "medium"
		}

		title := r.IssueText
		if title == "" {
			title = fmt.Sprintf("%s: %s", r.TestID, r.TestName)
		}
		if len(title) > 500 {
			title = title[:500]
		}

		filePath := r.FileName
		if filePath != "" && !filepath.IsAbs(filePath) {
			filePath = "/" + filePath
		}

		var lineStart *int
		if r.LineNumber > 0 {
			lineStart = &r.LineNumber
		}

		cweID := ""
		if r.IssueCwe != nil && len(r.IssueCwe.ID) > 0 {
			// CWE ID can be a number (79), string ("79"), or prefixed ("CWE-79")
			var n json.Number
			if err := json.Unmarshal(r.IssueCwe.ID, &n); err == nil {
				cweID = "CWE-" + n.String()
			} else {
				var s string
				if err := json.Unmarshal(r.IssueCwe.ID, &s); err == nil {
					if strings.HasPrefix(s, "CWE-") {
						cweID = s
					} else {
						cweID = "CWE-" + s
					}
				}
			}
		}

		description := r.IssueText
		if r.TestName != "" {
			description = fmt.Sprintf("[%s] %s", r.TestName, r.IssueText)
		}
		if len(description) > 2000 {
			description = description[:2000]
		}

		findings = append(findings, FindingInput{
			RuleID:      r.TestID,
			Title:       title,
			Severity:    severity,
			Description: description,
			FilePath:    filePath,
			LineStart:   lineStart,
			CWEID:       cweID,
		})
	}

	return findings, nil
}
