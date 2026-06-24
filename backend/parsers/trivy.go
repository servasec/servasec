package parsers

import (
	"encoding/json"
	"fmt"
	"strings"
)

type trivyResult struct {
	Target          string          `json:"Target"`
	Type            string          `json:"Type"`
	Vulnerabilities []trivyVuln     `json:"Vulnerabilities"`
	Secrets         []trivySecret   `json:"Secrets"`
	Misconfigurations []trivyMisconfig `json:"Misconfigurations"`
}

type trivyVuln struct {
	VulnerabilityID   string `json:"VulnerabilityID"`
	PkgName           string `json:"PkgName"`
	Severity          string `json:"Severity"`
	Title             string `json:"Title"`
	Description       string `json:"Description"`
	InstalledVersion  string `json:"InstalledVersion"`
	FixedVersion      string `json:"FixedVersion"`
	PrimaryURL        string `json:"PrimaryURL"`
}

type trivySecret struct {
	RuleID    string   `json:"RuleID"`
	Title     string   `json:"Title"`
	Severity  string   `json:"Severity"`
	Match     string   `json:"Match"`
	Target    string   `json:"Target"`
	Code      trivyCode `json:"Code"`
}

type trivyCode struct {
	Lines []trivyLine `json:"Lines"`
}

type trivyLine struct {
	Number int `json:"Number"`
}

type trivyMisconfig struct {
	ID       string   `json:"ID"`
	Title    string   `json:"Title"`
	Severity string   `json:"Severity"`
	Message  string   `json:"Message"`
	Type     string   `json:"Type"`
	CauseMetadata trivyCauseMetadata `json:"CauseMetadata"`
}

type trivyCauseMetadata struct {
	StartLine int `json:"StartLine"`
	EndLine   int `json:"EndLine"`
}

type trivyOutput struct {
	Results []trivyResult `json:"Results"`
}

func ParseTrivy(data []byte, filename string) ([]FindingInput, error) {
	var output trivyOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("invalid trivy JSON: %w", err)
	}

	var findings []FindingInput
	for _, r := range output.Results {
		for _, v := range r.Vulnerabilities {
			severity := strings.ToLower(v.Severity)
			if err := validateSeverity(severity); err != nil {
				severity = "medium"
			}

			title := v.VulnerabilityID
			if v.Title != "" {
				title = v.Title
				if len(title) > 500 {
					title = title[:500]
				}
			}

			description := v.Description
			if v.PkgName != "" {
				if v.InstalledVersion != "" && v.FixedVersion != "" {
					description = fmt.Sprintf("%s@%s - fixed in %s: %s", v.PkgName, v.InstalledVersion, v.FixedVersion, v.Description)
				} else if v.PkgName != "" {
					description = fmt.Sprintf("%s: %s", v.PkgName, v.Description)
				}
			}
			if len(description) > 2000 {
				description = description[:2000]
			}

			remediation := ""
			if v.FixedVersion != "" {
				remediation = fmt.Sprintf("Upgrade %s to version %s", v.PkgName, v.FixedVersion)
			}

			findings = append(findings, FindingInput{
				RuleID:      v.VulnerabilityID,
				Title:       title,
				Severity:    severity,
				Description: description,
				FilePath:    r.Target,
				Remediation: remediation,
			})
		}

		for _, s := range r.Secrets {
			severity := strings.ToLower(s.Severity)
			if err := validateSeverity(severity); err != nil {
				severity = "high"
			}

			var lineStart, lineEnd *int
			if len(s.Code.Lines) > 0 {
				first := s.Code.Lines[0].Number
				last := s.Code.Lines[len(s.Code.Lines)-1].Number
				lineStart = &first
				lineEnd = &last
			}

			findings = append(findings, FindingInput{
				RuleID:    s.RuleID,
				Title:     s.Title,
				Severity:  severity,
				FilePath:  s.Target,
				LineStart: lineStart,
				LineEnd:   lineEnd,
			})
		}

		for _, m := range r.Misconfigurations {
			severity := strings.ToLower(m.Severity)
			if err := validateSeverity(severity); err != nil {
				severity = "medium"
			}

			title := m.ID
			if m.Title != "" {
				title = m.Title
				if len(title) > 500 {
					title = title[:500]
				}
			}

			var lineStart, lineEnd *int
			if m.CauseMetadata.StartLine > 0 {
				lineStart = &m.CauseMetadata.StartLine
				lineEnd = &m.CauseMetadata.EndLine
			}

			findings = append(findings, FindingInput{
				RuleID:    m.ID,
				Title:     title,
				Severity:  severity,
				Description: m.Message,
				FilePath:  r.Target,
				LineStart: lineStart,
				LineEnd:   lineEnd,
			})
		}
	}

	return findings, nil
}

func validateSeverity(s string) error {
	switch s {
	case "critical", "high", "medium", "low", "info":
		return nil
	default:
		return fmt.Errorf("invalid severity: %s", s)
	}
}
