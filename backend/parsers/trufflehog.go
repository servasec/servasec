package parsers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type trufflehogResult struct {
	SourceMetadata *trufflehogSourceMetadata `json:"SourceMetadata"`
	DetectorName   string                    `json:"DetectorName"`
	DetectorType   int                       `json:"DetectorType"`
	Verified       bool                      `json:"Verified"`
	Raw            string                    `json:"Raw"`
	Redacted       string                    `json:"Redacted"`
	FileMetadata   *trufflehogFileMetadata   `json:"FileMetadata"`
	ExtraRawData   json.RawMessage           `json:"ExtraRawData,omitempty"`
	Commit         string                    `json:"Commit"`
}

type trufflehogSourceMetadata struct {
	Data struct {
		Filesystem *trufflehogFileData `json:"Filesystem,omitempty"`
		Git        *trufflehogGitData  `json:"Git,omitempty"`
		S3         *trufflehogS3Data   `json:"S3,omitempty"`
	} `json:"Data"`
}

type trufflehogFileData struct {
	File string `json:"file"`
}

type trufflehogGitData struct {
	Commit string `json:"commit"`
	File   string `json:"file"`
}

type trufflehogS3Data struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
}

type trufflehogFileMetadata struct {
	Filename string   `json:"Filename"`
	Filetype string   `json:"Filetype,omitempty"`
	Mode     string   `json:"Mode,omitempty"`
	Linkname []string `json:"Linkname,omitempty"`
	Owner    string   `json:"Owner,omitempty"`
}

func ParseTrufflehog(data []byte, filename string) ([]FindingInput, error) {
	// Detect if the input is a JSON array or JSON lines
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("empty trufflehog input")
	}

	if trimmed[0] == '[' {
		return parseTrufflehogArray(trimmed)
	}

	return parseTrufflehogLines(trimmed)
}

func parseTrufflehogArray(data []byte) ([]FindingInput, error) {
	var results []trufflehogResult
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("invalid trufflehog JSON array: %w", err)
	}

	var findings []FindingInput
	for _, r := range results {
		finding := trufflehogToFinding(r)
		if finding != nil {
			findings = append(findings, *finding)
		}
	}
	return findings, nil
}

func parseTrufflehogLines(data []byte) ([]FindingInput, error) {
	var findings []FindingInput
	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lineNum++

		var r trufflehogResult
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			return nil, fmt.Errorf("invalid trufflehog JSON on line %d: %w", lineNum, err)
		}

		finding := trufflehogToFinding(r)
		if finding != nil {
			findings = append(findings, *finding)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading trufflehog input: %w", err)
	}
	return findings, nil
}

func trufflehogToFinding(r trufflehogResult) *FindingInput {
	if r.DetectorName == "" {
		return nil
	}

	severity := "medium"
	if r.Verified {
		severity = "high"
	}

	filePath := resolveTrufflehogPath(r)

	description := fmt.Sprintf("Detected %s", r.DetectorName)
	if r.Redacted != "" {
		description = fmt.Sprintf("Detected %s: %s", r.DetectorName, r.Redacted)
		if len(description) > 2000 {
			description = description[:2000]
		}
	}

	title := r.DetectorName
	if len(title) > 500 {
		title = title[:500]
	}

	return &FindingInput{
		RuleID:      r.DetectorName,
		Title:       title,
		Severity:    severity,
		Description: description,
		FilePath:    filePath,
	}
}

func resolveTrufflehogPath(r trufflehogResult) string {
	if r.FileMetadata != nil && r.FileMetadata.Filename != "" {
		return r.FileMetadata.Filename
	}
	if r.SourceMetadata != nil {
		if fs := r.SourceMetadata.Data.Filesystem; fs != nil && fs.File != "" {
			return fs.File
		}
		if git := r.SourceMetadata.Data.Git; git != nil && git.File != "" {
			return git.File
		}
		if s3 := r.SourceMetadata.Data.S3; s3 != nil && s3.Key != "" {
			return fmt.Sprintf("s3://%s/%s", s3.Bucket, s3.Key)
		}
	}
	return ""
}
