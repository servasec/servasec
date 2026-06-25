package parsers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// nucleiResult models a single finding from nuclei's JSON/JSONL output. Nuclei is
// a DAST/template scanner, so findings are located by URL (matched-at) rather than
// by file path and line number.
type nucleiResult struct {
	TemplateID string     `json:"template-id"`
	Type       string     `json:"type"`
	Host       string     `json:"host"`
	MatchedAt  string     `json:"matched-at"`
	Info       nucleiInfo `json:"info"`
}

type nucleiInfo struct {
	Name           string                `json:"name"`
	Severity       string                `json:"severity"`
	Description    string                `json:"description"`
	Remediation    string                `json:"remediation"`
	Classification *nucleiClassification `json:"classification"`
}

// nucleiClassification holds CVE/CWE identifiers. Depending on the nuclei version
// these are emitted as a single string or as an array of strings, so they are kept
// as RawMessage and normalized via extractCWE (see semgrep.go).
type nucleiClassification struct {
	CVEID json.RawMessage `json:"cve-id"`
	CWEID json.RawMessage `json:"cwe-id"`
}

// ParseNuclei accepts either a JSON array (nuclei -json-export) or JSONL
// (nuclei -jsonl, one finding per line) and returns normalized findings.
func ParseNuclei(data []byte, filename string) ([]FindingInput, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty nuclei input")
	}

	if trimmed[0] == '[' {
		return parseNucleiArray(trimmed)
	}

	return parseNucleiLines(trimmed)
}

func parseNucleiArray(data []byte) ([]FindingInput, error) {
	var results []nucleiResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("invalid nuclei JSON array: %w", err)
	}

	var findings []FindingInput
	for _, r := range results {
		if f := nucleiToFinding(r); f != nil {
			findings = append(findings, *f)
		}
	}
	return findings, nil
}

func parseNucleiLines(data []byte) ([]FindingInput, error) {
	var findings []FindingInput
	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Nuclei JSONL lines can embed full HTTP request/response pairs and easily
	// exceed bufio.Scanner's default 64KB token limit, so grow the buffer.
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	lineNum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lineNum++

		var r nucleiResult
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			return nil, fmt.Errorf("invalid nuclei JSON on line %d: %w", lineNum, err)
		}

		if f := nucleiToFinding(r); f != nil {
			findings = append(findings, *f)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading nuclei input: %w", err)
	}
	return findings, nil
}

func nucleiToFinding(r nucleiResult) *FindingInput {
	// A valid nuclei finding always carries at least a template-id or a name.
	if r.TemplateID == "" && r.Info.Name == "" {
		return nil
	}

	// Nuclei severities (info/low/medium/high/critical) already match servasec's
	// scheme. Anything else (e.g. "unknown") is treated as informational.
	severity := strings.ToLower(r.Info.Severity)
	if validateSeverity(severity) != nil {
		severity = "info"
	}

	ruleID := r.TemplateID
	if ruleID == "" {
		ruleID = r.Info.Name
	}

	title := r.TemplateID
	if r.Info.Name != "" {
		title = r.Info.Name
	}
	if len(title) > 500 {
		title = title[:500]
	}

	// DAST findings are located by URL; matched-at is the precise endpoint, host
	// is the fallback base target.
	location := r.MatchedAt
	if location == "" {
		location = r.Host
	}

	description := r.Info.Description
	if location != "" {
		if description != "" {
			description = fmt.Sprintf("%s\nMatched at: %s", description, location)
		} else {
			description = fmt.Sprintf("Matched at: %s", location)
		}
	}
	if len(description) > 2000 {
		description = description[:2000]
	}

	cweID := ""
	if r.Info.Classification != nil {
		cweID = extractCWE(r.Info.Classification.CWEID)
	}

	return &FindingInput{
		RuleID:      ruleID,
		Title:       title,
		Severity:    severity,
		Description: description,
		FilePath:    location,
		CWEID:       cweID,
		Remediation: r.Info.Remediation,
	}
}
