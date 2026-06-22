package parsers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type semgrepResult struct {
	CheckID string `json:"check_id"`
	Path    string `json:"path"`
	Start   struct {
		Line int `json:"line"`
	} `json:"start"`
	End struct {
		Line int `json:"line"`
	} `json:"end"`
	Extra struct {
		Message  string `json:"message"`
		Severity string `json:"severity"`
		Metadata struct {
			CWE        json.RawMessage `json:"cwe"`
			Technology []string        `json:"technology"`
		} `json:"metadata"`
		Lines string `json:"lines"`
	} `json:"extra"`
}

type semgrepResponse struct {
	Results []semgrepResult `json:"results"`
}

func extractCWE(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil && len(arr) > 0 {
		return arr[0]
	}
	return ""
}

func ParseSemgrep(data []byte, filename string) ([]FindingInput, error) {
	var resp semgrepResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("invalid semgrep JSON: %w", err)
	}

	var findings []FindingInput
	for _, r := range resp.Results {
		severity := strings.ToLower(r.Extra.Severity)
		switch severity {
		case "error":
			severity = "high"
		case "warning":
			severity = "medium"
		case "info":
			severity = "low"
		}

		lineStart := r.Start.Line
		lineEnd := r.End.Line

		title := r.CheckID
		if r.Extra.Message != "" {
			title = r.Extra.Message
			if len(title) > 500 {
				title = title[:500]
			}
		}

		filePath := r.Path
		if !filepath.IsAbs(filePath) {
			filePath = "/" + filePath
		}

		findings = append(findings, FindingInput{
			RuleID:      r.CheckID,
			Title:       title,
			Severity:    severity,
			Description: r.Extra.Message,
			FilePath:    filePath,
			LineStart:   &lineStart,
			LineEnd:     &lineEnd,
			CWEID:       extractCWE(r.Extra.Metadata.CWE),
		})
	}

	return findings, nil
}
