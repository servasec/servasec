package parsers

import (
	"encoding/json"
	"fmt"
	"strings"
)

type grypeOutput struct {
	Matches []grypeMatch `json:"matches"`
}

type grypeMatch struct {
	Vulnerability grypeVulnerability `json:"vulnerability"`
	Artifact      grypeArtifact      `json:"artifact"`
}

type grypeVulnerability struct {
	ID          string       `json:"id"`
	Severity    string       `json:"severity"`
	Description string       `json:"description"`
	Fix         *grypeFix    `json:"fix"`
}

type grypeFix struct {
	Versions []string `json:"versions"`
	State    string   `json:"state"`
}

type grypeArtifact struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func ParseGrype(data []byte, filename string) ([]FindingInput, error) {
	var output grypeOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("invalid grype JSON: %w", err)
	}

	var findings []FindingInput
	for _, m := range output.Matches {
		severity := strings.ToLower(m.Vulnerability.Severity)
		if err := validateSeverity(severity); err != nil {
			severity = "medium"
		}

		title := m.Vulnerability.ID
		if title == "" {
			title = m.Artifact.Name
		}
		if len(title) > 500 {
			title = title[:500]
		}

		description := m.Vulnerability.Description
		if description == "" {
			description = fmt.Sprintf("%s@%s is affected by %s", m.Artifact.Name, m.Artifact.Version, m.Vulnerability.ID)
		}
		if len(description) > 2000 {
			description = description[:2000]
		}

		remediation := ""
		if m.Vulnerability.Fix != nil && m.Vulnerability.Fix.State == "fixed" && len(m.Vulnerability.Fix.Versions) > 0 {
			remediation = fmt.Sprintf("Upgrade %s to version %s", m.Artifact.Name, m.Vulnerability.Fix.Versions[0])
		}

		filePath := fmt.Sprintf("%s@%s", m.Artifact.Name, m.Artifact.Version)

		findings = append(findings, FindingInput{
			RuleID:      m.Vulnerability.ID,
			Title:       title,
			Severity:    severity,
			Description: description,
			FilePath:    filePath,
			Remediation: remediation,
		})
	}

	return findings, nil
}
