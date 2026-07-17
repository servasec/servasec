package parsers

import (
	"encoding/json"
	"fmt"
	"strings"
)

type kubescapeOutput struct {
	Results []kubescapeResult `json:"results"`
}

type kubescapeResult struct {
	Resources  []kubescapeResource  `json:"resources"`
	Controls   []kubescapeControl   `json:"controls"`
}

type kubescapeResource struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Namespace string `json:"namespace"`
}

type kubescapeControl struct {
	ControlID   string `json:"controlID"`
	Name        string `json:"name"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	FailedResources int `json:"failedResources"`
}

func ParseKubescape(data []byte, filename string) ([]FindingInput, error) {
	var output kubescapeOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("invalid kubescape JSON: %w", err)
	}

	var findings []FindingInput
	for _, result := range output.Results {
		for _, ctrl := range result.Controls {
			if ctrl.FailedResources == 0 {
				continue
			}

			severity := strings.ToLower(ctrl.Severity)
			if err := validateSeverity(severity); err != nil {
				severity = "medium"
			}

			title := ctrl.Name
			if title == "" {
				title = ctrl.ControlID
			}
			if len(title) > 500 {
				title = title[:500]
			}

			description := ctrl.Description
			if description == "" {
				description = fmt.Sprintf("Control %s: %d resources failed", ctrl.ControlID, ctrl.FailedResources)
			}
			if len(description) > 2000 {
				description = description[:2000]
			}

			for _, res := range result.Resources {
				filePath := fmt.Sprintf("%s/%s", res.Kind, res.Name)
				if res.Namespace != "" {
					filePath = fmt.Sprintf("%s/%s/%s", res.Namespace, res.Kind, res.Name)
				}

				findings = append(findings, FindingInput{
					RuleID:      ctrl.ControlID,
					Title:       title,
					Severity:    severity,
					Description: description,
					FilePath:    filePath,
				})
			}

			if len(result.Resources) == 0 {
				findings = append(findings, FindingInput{
					RuleID:      ctrl.ControlID,
					Title:       title,
					Severity:    severity,
					Description: description,
				})
			}
		}
	}

	return findings, nil
}
