package parsers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type sarifLog struct {
	Version string    `json:"version"`
	Schema  string    `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
	Properties sarifProperties `json:"properties,omitempty"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name  string      `json:"name"`
	Rules []sarifRule `json:"rules,omitempty"`
}

type sarifDefaultConfiguration struct {
	Level string `json:"level"`
}

type sarifRule struct {
	ID                   string                    `json:"id"`
	Properties           sarifProperties           `json:"properties,omitempty"`
	DefaultConfiguration *sarifDefaultConfiguration `json:"defaultConfiguration,omitempty"`
}

type sarifResult struct {
	RuleID      string           `json:"ruleId"`
	RuleIndex   int              `json:"ruleIndex"`
	Level       string           `json:"level"`
	Message     sarifMessage     `json:"message"`
	Locations   []sarifLocation  `json:"locations"`
	Properties  sarifProperties  `json:"properties,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           *sarifRegion           `json:"region,omitempty"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int  `json:"startLine"`
	EndLine   *int `json:"endLine,omitempty"`
}

type sarifProperties struct {
	Severity string   `json:"severity,omitempty"`
	Priority string   `json:"priority,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

func ExtractSarifToolName(data []byte) string {
	var header struct {
		Runs []struct {
			Tool struct {
				Driver struct {
					Name string `json:"name"`
				} `json:"driver"`
			} `json:"tool"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return ""
	}
	if len(header.Runs) == 0 {
		return ""
	}
	name := strings.ToLower(header.Runs[0].Tool.Driver.Name)
	switch name {
	case "opengrep", "opengrep oss":
		return "semgrep"
	}
	return name
}

func sarifSeverity(level string, props sarifProperties) string {
	for _, s := range []string{props.Severity, props.Priority, level} {
		switch strings.ToLower(s) {
		case "critical", "error", "high":
			return "high"
		case "warning", "medium":
			return "medium"
		case "note", "low":
			return "low"
		case "none", "info":
			return "info"
		}
	}
	return "info"
}

func sarifCWE(tags []string) string {
	for _, tag := range tags {
		lower := strings.ToLower(tag)
		var suffix string
		switch {
		case strings.HasPrefix(lower, "cwe-"):
			suffix = tag[4:]
		case strings.HasPrefix(lower, "cwe/"):
			suffix = tag[4:]
		case strings.HasPrefix(lower, "cwe:"):
			suffix = tag[4:]
		default:
			if idx := strings.Index(lower, "/cwe-"); idx >= 0 {
				suffix = tag[idx+5:]
			} else if idx := strings.Index(lower, "/cwe:"); idx >= 0 {
				suffix = tag[idx+5:]
			} else {
				continue
			}
		}
		if suffix != "" {
			return "CWE-" + suffix
		}
	}
	return ""
}

func sarifFilePath(uri string) string {
	p := strings.TrimPrefix(uri, "file://")
	if !filepath.IsAbs(p) {
		p = "/" + p
	}
	return p
}

func ParseSarif(data []byte, filename string) ([]FindingInput, error) {
	var log sarifLog
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, fmt.Errorf("invalid SARIF JSON: %w", err)
	}

	if len(log.Runs) == 0 {
		return nil, fmt.Errorf("SARIF log contains no runs")
	}

	var findings []FindingInput
	for _, run := range log.Runs {
		// ruleId - defaultConfiguration.level for severity fallback
		ruleDefaultLevel := make(map[string]string)
		for _, rule := range run.Tool.Driver.Rules {
			if rule.DefaultConfiguration != nil && rule.DefaultConfiguration.Level != "" {
				ruleDefaultLevel[rule.ID] = rule.DefaultConfiguration.Level
			}
		}

		ruleTagMap := make(map[int][]string)
		for idx, rule := range run.Tool.Driver.Rules {
			ruleTagMap[idx] = rule.Properties.Tags
		}

		for _, r := range run.Results {
			level := r.Level
			if level == "" {
				if dl, ok := ruleDefaultLevel[r.RuleID]; ok {
					level = dl
				}
			}
			severity := sarifSeverity(level, r.Properties)

			title := r.Message.Text
			if title == "" {
				title = r.RuleID
			}
			if len(title) > 500 {
				title = title[:500]
			}

			description := r.Message.Text
			if len(description) > 2000 {
				description = description[:2000]
			}

			cweID := sarifCWE(r.Properties.Tags)
			if cweID == "" && r.RuleIndex >= 0 {
				if tags, ok := ruleTagMap[r.RuleIndex]; ok {
					cweID = sarifCWE(tags)
				}
			}

			filePath := ""
			var lineStart, lineEnd int
			hasLocation := false
			if len(r.Locations) > 0 {
				loc := r.Locations[0]
				filePath = sarifFilePath(loc.PhysicalLocation.ArtifactLocation.URI)
				if loc.PhysicalLocation.Region != nil {
					lineStart = loc.PhysicalLocation.Region.StartLine
					if loc.PhysicalLocation.Region.EndLine != nil {
						lineEnd = *loc.PhysicalLocation.Region.EndLine
					}
				}
				hasLocation = true
			}

			f := FindingInput{
				RuleID:      r.RuleID,
				Title:       title,
				Severity:    severity,
				Description: description,
				FilePath:    filePath,
				CWEID:       cweID,
			}
			if hasLocation && lineStart > 0 {
				f.LineStart = &lineStart
			}
			if hasLocation && lineEnd > 0 {
				f.LineEnd = &lineEnd
			}
			findings = append(findings, f)
		}
	}

	return findings, nil
}
