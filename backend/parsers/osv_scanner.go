package parsers

import (
	"encoding/json"
	"fmt"
	"strings"
)

type osvOutput struct {
	Results []osvResult `json:"results"`
}

type osvResult struct {
	Source   osvSource   `json:"source"`
	Packages []osvPackage `json:"packages"`
}

type osvSource struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type osvPackage struct {
	Package        osvPackageInfo    `json:"package"`
	Vulnerabilities []osvVulnerability `json:"vulnerabilities"`
}

type osvPackageInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Ecosystem string `json:"ecosystem"`
}

type osvVulnerability struct {
	ID      string   `json:"id"`
	Aliases []string `json:"aliases"`
	Summary string   `json:"summary"`
	Details string   `json:"details"`
	Severity []osvSeverity `json:"severity"`
}

type osvSeverity struct {
	Type  string `json:"type"`
	Score string `json:"score"`
}

func ParseOSVScanner(data []byte, filename string) ([]FindingInput, error) {
	var output osvOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("invalid osv-scanner JSON: %w", err)
	}

	var findings []FindingInput
	for _, result := range output.Results {
		for _, pkg := range result.Packages {
			for _, vuln := range pkg.Vulnerabilities {
				vulnID := vuln.ID
				if len(vuln.Aliases) > 0 {
					for _, alias := range vuln.Aliases {
						if strings.HasPrefix(alias, "CVE-") {
							vulnID = alias
							break
						}
					}
				}

				severity := "medium"
				if len(vuln.Severity) > 0 {
					severity = mapOSVSeverity(vuln.Severity[0].Score)
				}

				title := vuln.Summary
				if title == "" {
					title = vulnID
				}
				if len(title) > 500 {
					title = title[:500]
				}

				description := vuln.Details
				if description == "" {
					description = fmt.Sprintf("%s@%s is affected by %s", pkg.Package.Name, pkg.Package.Version, vulnID)
				}
				if len(description) > 2000 {
					description = description[:2000]
				}

				filePath := fmt.Sprintf("%s@%s", pkg.Package.Name, pkg.Package.Version)
				if result.Source.Path != "" {
					filePath = result.Source.Path + " (" + filePath + ")"
				}

				findings = append(findings, FindingInput{
					RuleID:      vulnID,
					Title:       title,
					Severity:    severity,
					Description: description,
					FilePath:    filePath,
				})
			}
		}
	}

	return findings, nil
}

func mapOSVSeverity(score string) string {
	lower := strings.ToLower(score)
	switch {
	case strings.Contains(lower, "critical") || strings.Contains(lower, "10.0") || strings.Contains(lower, "9."):
		return "critical"
	case strings.Contains(lower, "high") || strings.Contains(lower, "7.") || strings.Contains(lower, "8."):
		return "high"
	case strings.Contains(lower, "medium") || strings.Contains(lower, "4.") || strings.Contains(lower, "5.") || strings.Contains(lower, "6."):
		return "medium"
	case strings.Contains(lower, "low") || strings.Contains(lower, "0.") || strings.Contains(lower, "1.") || strings.Contains(lower, "2.") || strings.Contains(lower, "3."):
		return "low"
	default:
		return "medium"
	}
}
