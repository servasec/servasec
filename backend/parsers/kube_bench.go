package parsers

import (
	"encoding/json"
	"fmt"
	"strings"
)

type kubeBenchOutput struct {
	Controls []kubeBenchControl `json:"controls"`
}

type kubeBenchControl struct {
	ID       string           `json:"id"`
	Text     string           `json:"text"`
	NodeType string           `json:"node_type"`
	Version  string           `json:"version"`
	Tests    []kubeBenchTest  `json:"tests"`
}

type kubeBenchTest struct {
	Section  string             `json:"section"`
	Desc     string             `json:"desc"`
	Results  []kubeBenchResult  `json:"results"`
}

type kubeBenchResult struct {
	TestNumber  string `json:"test_number"`
	TestDesc    string `json:"test_desc"`
	Status      string `json:"status"`
	Remediation string `json:"remediation"`
}

func ParseKubeBench(data []byte, filename string) ([]FindingInput, error) {
	var output kubeBenchOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("invalid kube-bench JSON: %w", err)
	}

	var findings []FindingInput
	for _, ctrl := range output.Controls {
		for _, test := range ctrl.Tests {
			for _, r := range test.Results {
				if r.Status != "FAIL" {
					continue
				}

				severity := "medium"
				if ctrl.NodeType == "master" || strings.HasPrefix(ctrl.ID, "1") {
					severity = "high"
				}

				title := r.TestDesc
				if title == "" {
					title = r.TestNumber
				}
				if len(title) > 500 {
					title = title[:500]
				}

				description := r.TestDesc
				if len(description) > 2000 {
					description = description[:2000]
				}

				filePath := fmt.Sprintf("kube-bench:%s:%s", ctrl.NodeType, ctrl.ID)

				findings = append(findings, FindingInput{
					RuleID:      r.TestNumber,
					Title:       title,
					Severity:    severity,
					Description: description,
					FilePath:    filePath,
					Remediation: r.Remediation,
				})
			}
		}
	}

	return findings, nil
}
