package parsers

import (
	"encoding/json"
	"strings"
)

type FindingInput struct {
	RuleID      string `json:"ruleId"`
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	FilePath    string `json:"filePath"`
	LineStart   *int   `json:"lineStart"`
	LineEnd     *int   `json:"lineEnd"`
	CWEID       string `json:"cweId"`
	Remediation string `json:"remediation"`
}

type ParserFunc func(data []byte, filename string) ([]FindingInput, error)

var registry = map[string]ParserFunc{
	"semgrep":      ParseSemgrep,
	"trivy":        ParseTrivy,
	"gitleaks":     ParseGitleaks,
	"grype":        ParseGrype,
	"snyk":         ParseSnyk,
	"checkov":      ParseCheckov,
	"trufflehog":   ParseTrufflehog,
	"nuclei":       ParseNuclei,
	"sarif":        ParseSarif,
	"gosec":        ParseGosec,
	"bandit":       ParseBandit,
	"osv-scanner":  ParseOSVScanner,
	"npm-audit":    ParseNpmAudit,
	"tfsec":        ParseTfsec,
	"kubescape":    ParseKubescape,
	"kube-bench":   ParseKubeBench,
}

func Get(name string) (ParserFunc, bool) {
	f, ok := registry[name]
	return f, ok
}

func Register(name string, fn ParserFunc) {
	registry[name] = fn
}

func DetectScannerType(data []byte) string {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return ""
	}

	str := string(data)
	switch {
	case (strings.Contains(str, `"$schema"`) || strings.Contains(str, `"version"`)) && strings.Contains(str, `"runs"`):
		return "sarif"
	case strings.Contains(str, `"check_id":`):
		return "semgrep"
	case strings.Contains(str, `"Target":`) && strings.Contains(str, `"Type":`):
		return "trivy"
	case strings.Contains(str, `"Description":`) && strings.Contains(str, `"StartLine":`) && strings.Contains(str, `"RuleID":`):
		return "gitleaks"
	case strings.Contains(str, `"matches":`) && strings.Contains(str, `"descriptor":`):
		return "grype"
	case strings.Contains(str, `"vulns":`):
		return "snyk"
	case strings.Contains(str, `"results":`) && strings.Contains(str, `"passed_checks":`):
		return "checkov"
	case strings.Contains(str, `"DetectorName":`) && strings.Contains(str, `"SourceMetadata":`):
		return "trufflehog"
	case strings.Contains(str, `"template-id":`) && strings.Contains(str, `"matched-at":`):
		return "nuclei"
	case strings.Contains(str, `"Issues":`) && strings.Contains(str, `"rule_id":`):
		return "gosec"
	case strings.Contains(str, `"test_id":`) && strings.Contains(str, `"issue_text":`):
		return "bandit"
	case strings.Contains(str, `"source":`) && strings.Contains(str, `"packages":`) && strings.Contains(str, `"ecosystem":`):
		return "osv-scanner"
	case strings.Contains(str, `"packageManager":`) && strings.Contains(str, `"dependencyCount":`):
		return "npm-audit"
	case strings.Contains(str, `"rule_id":`) && strings.Contains(str, `"resolution":`) && strings.Contains(str, `"passed":`):
		return "tfsec"
	case strings.Contains(str, `"controls":`) && strings.Contains(str, `"failedResources":`):
		return "kubescape"
	case strings.Contains(str, `"node_type":`) && strings.Contains(str, `"test_number":`):
		return "kube-bench"
	}

	if arr, ok := raw.([]any); ok && len(arr) > 0 {
		if first, ok := arr[0].(map[string]any); ok {
			if _, ok := first["check_id"]; ok {
				return "semgrep"
			}
		}
	}

	return ""
}
