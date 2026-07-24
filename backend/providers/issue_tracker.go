package providers

import (
	"fmt"
	"os"
)

type IssueTrackerProvider interface {
	CreateIssue(cfg IssueTrackerConfig, finding IssueData) (*ExternalIssue, error)
	TestConnection(cfg IssueTrackerConfig) error
	Name() string
}

type IssueTrackerConfig struct {
	Provider             string
	RepositoryURL        string
	Token                string
	ServaSecURL          string
	AuthType             string
	GitHubAppID          *int64
	GitHubPrivateKey     string
	GitHubInstallationID *int64
}

type IssueData struct {
	FindingID   uint
	Title       string
	Description string
	Severity    string
	FilePath    string
	LineStart   *int
	LineEnd     *int
	CWEID       string
	RuleID      string
	ScannerType string
}

type ExternalIssue struct {
	IssueID  string
	IssueURL string
}

var SeverityLabels = map[string]string{
	"critical": "severity::critical",
	"high":     "severity::high",
	"medium":   "severity::medium",
	"low":      "severity::low",
	"info":     "severity::info",
}

var SeverityPriority = map[string]int{
	"critical": 5,
	"high":     4,
	"medium":   3,
	"low":      2,
	"info":     1,
}

func SeverityMeetsThreshold(severity, threshold string) bool {
	return SeverityPriority[severity] <= SeverityPriority[threshold]
}

func GetServaSecURL() string {
	return os.Getenv("SSC_PUBLIC_URL")
}

func TruncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}

func FormatFindingBody(cfg IssueTrackerConfig, finding IssueData) string {
	body := ""

	if finding.FilePath != "" {
		body += fmt.Sprintf("**File:** `%s`\n", finding.FilePath)
		if finding.LineStart != nil {
			body += fmt.Sprintf("**Line:** %d", *finding.LineStart)
			if finding.LineEnd != nil && *finding.LineEnd != *finding.LineStart {
				body += fmt.Sprintf("-%d", *finding.LineEnd)
			}
			body += "\n"
		}
		body += "\n"
	}

	if finding.CWEID != "" {
		body += fmt.Sprintf("**CWE:** [%s](https://cwe.mitre.org/data/definitions/%s.html)\n\n", finding.CWEID, finding.CWEID)
	}

	if finding.RuleID != "" {
		body += fmt.Sprintf("**Rule:** %s\n\n", finding.RuleID)
	}

	if finding.Description != "" {
		body += finding.Description + "\n\n"
	}

	if cfg.ServaSecURL != "" && finding.FindingID > 0 {
		body += fmt.Sprintf("---\n\n[View in servasec](%s/findings/%d)", cfg.ServaSecURL, finding.FindingID)
	}

	return body
}
