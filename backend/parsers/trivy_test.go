package parsers

import (
	"testing"
)

func TestParseTrivy(t *testing.T) {
	data := []byte(`{
  "Results": [
    {
      "Target": "package-lock.json",
      "Type": "npm",
      "Vulnerabilities": [
        {
          "VulnerabilityID": "CVE-2024-1234",
          "PkgName": "lodash",
          "Severity": "CRITICAL",
          "Title": "Prototype Pollution in lodash",
          "Description": "A vulnerability in lodash allows prototype pollution.",
          "InstalledVersion": "4.17.20",
          "FixedVersion": "4.17.21",
          "PrimaryURL": "https://avd.aquasec.com/npm/cve-2024-1234"
        },
        {
          "VulnerabilityID": "CVE-2024-5678",
          "PkgName": "axios",
          "Severity": "HIGH",
          "Title": "SSRF in axios",
          "Description": "Server-Side Request Forgery in axios.",
          "InstalledVersion": "0.21.0",
          "FixedVersion": "0.21.1"
        }
      ]
    },
    {
      "Target": "Dockerfile",
      "Type": "dockerfile",
      "Misconfigurations": [
        {
          "ID": "DS001",
          "Title": "Missing USER directive",
          "Severity": "MEDIUM",
          "Message": "The Dockerfile should specify a USER directive",
          "Type": "dockerfile-security",
          "CauseMetadata": {
            "StartLine": 1,
            "EndLine": 20
          }
        }
      ]
    }
  ]
}`)

	findings, err := ParseTrivy(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "CVE-2024-1234" {
		t.Errorf("expected CVE-2024-1234, got %s", f1.RuleID)
	}
	if f1.Severity != "critical" {
		t.Errorf("expected critical, got %s", f1.Severity)
	}
	if f1.Title != "Prototype Pollution in lodash" {
		t.Errorf("unexpected title: %s", f1.Title)
	}
	if f1.FilePath != "package-lock.json" {
		t.Errorf("expected package-lock.json, got %s", f1.FilePath)
	}
	if f1.Remediation != "Upgrade lodash to version 4.17.21" {
		t.Errorf("unexpected remediation: %s", f1.Remediation)
	}
	if f1.Description == "" {
		t.Error("expected non-empty description")
	}

	f2 := findings[1]
	if f2.RuleID != "CVE-2024-5678" {
		t.Errorf("expected CVE-2024-5678, got %s", f2.RuleID)
	}
	if f2.Severity != "high" {
		t.Errorf("expected high, got %s", f2.Severity)
	}

	f3 := findings[2]
	if f3.RuleID != "DS001" {
		t.Errorf("expected DS001, got %s", f3.RuleID)
	}
	if f3.Severity != "medium" {
		t.Errorf("expected medium, got %s", f3.Severity)
	}
	if f3.LineStart == nil || *f3.LineStart != 1 {
		t.Errorf("expected line 1, got %v", f3.LineStart)
	}
}

func TestParseTrivy_InvalidJSON(t *testing.T) {
	_, err := ParseTrivy([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseTrivy_Empty(t *testing.T) {
	findings, err := ParseTrivy([]byte(`{"Results": []}`), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestValidateSeverity(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"critical", true},
		{"high", true},
		{"medium", true},
		{"low", true},
		{"info", true},
		{"INVALID", false},
		{"", false},
	}
	for _, tt := range tests {
		err := validateSeverity(tt.input)
		if tt.valid && err != nil {
			t.Errorf("expected %s to be valid, got error: %v", tt.input, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("expected %s to be invalid", tt.input)
		}
	}
}
