package parsers

import (
	"encoding/json"
	"fmt"
	"strings"
)

type checkovOutput struct {
	Results checkovResults `json:"results"`
}

type checkovResults struct {
	FailedChecks  []checkovCheck `json:"failed_checks"`
	PassedChecks  []checkovCheck `json:"passed_checks"`
	SkippedChecks []checkovCheck `json:"skipped_checks"`
}

type checkovCheck struct {
	Check          checkovCheckDef    `json:"check"`
	CheckID        string             `json:"check_id"`
	FilePath       string             `json:"file_path"`
	FileLineRange []int              `json:"file_line_range"`
	Resource       string             `json:"resource"`
}

type checkovCheckDef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

func ParseCheckov(data []byte, filename string) ([]FindingInput, error) {
	var output checkovOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("invalid checkov JSON: %w", err)
	}

	var findings []FindingInput
	for _, c := range output.Results.FailedChecks {
		checkID := c.CheckID
		if checkID == "" {
			checkID = c.Check.ID
		}

		severity := strings.ToLower(c.Check.Severity)
		if err := validateSeverity(severity); err != nil {
			severity = "medium"
		}

		title := c.Check.Name
		if title == "" {
			title = checkID
		}
		if len(title) > 500 {
			title = title[:500]
		}

		description := c.Check.Description
		if description == "" {
			description = fmt.Sprintf("Check %s failed for %s", checkID, c.Resource)
		}
		if len(description) > 2000 {
			description = description[:2000]
		}

		filePath := c.FilePath
		if !strings.HasPrefix(filePath, "/") {
			filePath = "/" + filePath
		}

		var lineStart, lineEnd *int
		if len(c.FileLineRange) >= 2 {
			lineStart = &c.FileLineRange[0]
			lineEnd = &c.FileLineRange[1]
		} else if len(c.FileLineRange) == 1 {
			lineStart = &c.FileLineRange[0]
		}

		findings = append(findings, FindingInput{
			RuleID:      checkID,
			Title:       title,
			Severity:    severity,
			Description: description,
			FilePath:    filePath,
			LineStart:   lineStart,
			LineEnd:     lineEnd,
		})
	}

	return findings, nil
}
