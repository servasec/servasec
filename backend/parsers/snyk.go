package parsers

import (
	"encoding/json"
	"fmt"
	"strings"
)

type snykOutput struct {
	Vulnerabilities []snykVulnerability `json:"vulnerabilities"`
	PackageManager  string              `json:"packageManager"`
}

type snykVulnerability struct {
	ID               string           `json:"id"`
	Title            string           `json:"title"`
	Severity         string           `json:"severity"`
	Description      string           `json:"description"`
	PackageName      string           `json:"packageName"`
	Version          string           `json:"version"`
	Identifiers      snykIdentifiers  `json:"identifiers"`
	UpgradePath      []string         `json:"upgradePath"`
	IsUpgradable     bool             `json:"isUpgradable"`
	SemVer           snykSemVer       `json:"semver"`
}

type snykIdentifiers struct {
	CWE []string `json:"CWE"`
	CVE []string `json:"CVE"`
}

type snykSemVer struct {
	Vulnerable []string `json:"vulnerable"`
}

func ParseSnyk(data []byte, filename string) ([]FindingInput, error) {
	var output snykOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("invalid snyk JSON: %w", err)
	}

	var findings []FindingInput
	for _, v := range output.Vulnerabilities {
		severity := strings.ToLower(v.Severity)
		if err := validateSeverity(severity); err != nil {
			severity = "medium"
		}

		title := v.Title
		if title == "" {
			title = v.ID
		}
		if len(title) > 500 {
			title = title[:500]
		}

		description := v.Description
		if description == "" {
			description = fmt.Sprintf("%s@%s is affected by %s", v.PackageName, v.Version, v.ID)
		}
		if len(description) > 2000 {
			description = description[:2000]
		}

		filePath := fmt.Sprintf("%s@%s", v.PackageName, v.Version)

		cweID := ""
		if len(v.Identifiers.CWE) > 0 {
			cweID = v.Identifiers.CWE[0]
		}

		remediation := ""
		if v.IsUpgradable && len(v.UpgradePath) > 1 {
			remediation = fmt.Sprintf("Upgrade %s to %s", v.PackageName, v.UpgradePath[len(v.UpgradePath)-1])
		}

		findings = append(findings, FindingInput{
			RuleID:      v.ID,
			Title:       title,
			Severity:    severity,
			Description: description,
			FilePath:    filePath,
			CWEID:       cweID,
			Remediation: remediation,
		})
	}

	return findings, nil
}
