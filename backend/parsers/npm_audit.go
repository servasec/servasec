package parsers

import (
	"encoding/json"
	"fmt"
	"strings"
)

type npmAuditOutput struct {
	Vulnerabilities map[string]npmAuditVuln `json:"vulnerabilities"`
}

type npmAuditVuln struct {
	Name         string              `json:"name"`
	Severity     string              `json:"severity"`
	Range         string             `json:"range"`
	Via          []npmAuditVia       `json:"via"`
	FixAvailable  interface{}        `json:"fixAvailable"`
}

type npmAuditVia interface{}

type npmAuditViaDirect struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Range string `json:"range"`
}

func ParseNpmAudit(data []byte, filename string) ([]FindingInput, error) {
	var output npmAuditOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("invalid npm audit JSON: %w", err)
	}

	var findings []FindingInput
	for name, vuln := range output.Vulnerabilities {
		severity := strings.ToLower(vuln.Severity)
		if err := validateSeverity(severity); err != nil {
			severity = "medium"
		}

		title := ""
		description := ""
		for _, v := range vuln.Via {
			switch via := v.(type) {
			case string:
				if title == "" {
					title = via
				}
			case map[string]interface{}:
				if t, ok := via["title"].(string); ok && title == "" {
					title = t
				}
			}
		}
		if title == "" {
			title = fmt.Sprintf("Vulnerability in %s", name)
		}
		if len(title) > 500 {
			title = title[:500]
		}

		vulnIDs := []string{}
		for _, v := range vuln.Via {
			if obj, ok := v.(map[string]interface{}); ok {
				if url, ok := obj["url"].(string); ok {
					vulnIDs = append(vulnIDs, url)
				}
			}
		}
		if len(vulnIDs) > 0 {
			description = fmt.Sprintf("Advisories: %s", strings.Join(vulnIDs, ", "))
		} else {
			description = title
		}
		if len(description) > 2000 {
			description = description[:2000]
		}

		ruleID := name
		if len(vulnIDs) > 0 {
			parts := strings.Split(vulnIDs[0], "/")
			if len(parts) > 0 {
				ruleID = parts[len(parts)-1]
			}
		}

		filePath := name
		if vuln.Range != "" {
			filePath = fmt.Sprintf("%s@%s", name, vuln.Range)
		}

		remediation := ""
		if fixAvail, ok := vuln.FixAvailable.(map[string]interface{}); ok {
			if version, ok := fixAvail["version"].(string); ok {
				remediation = fmt.Sprintf("Upgrade %s to %s", name, version)
			}
		} else if fixAvail, ok := vuln.FixAvailable.(bool); ok && fixAvail {
			remediation = fmt.Sprintf("Upgrade %s to latest", name)
		}

		findings = append(findings, FindingInput{
			RuleID:      ruleID,
			Title:       title,
			Severity:    severity,
			Description: description,
			FilePath:    filePath,
			Remediation: remediation,
		})
	}

	return findings, nil
}
