package parsers

import (
	"testing"
)

func TestParseGrype(t *testing.T) {
	data := []byte(`{
  "matches": [
    {
      "vulnerability": {
        "id": "CVE-2021-36159",
        "severity": "Critical",
        "description": "libapk-tools before 2.10.7-r0 has a double free.",
        "fix": {
          "versions": ["2.10.7-r0"],
          "state": "fixed"
        }
      },
      "artifact": {
        "name": "apk-tools",
        "version": "2.10.6-r0"
      }
    },
    {
      "vulnerability": {
        "id": "CVE-2024-9999",
        "severity": "High",
        "description": "",
        "fix": {
          "versions": [],
          "state": "not-fixed"
        }
      },
      "artifact": {
        "name": "openssl",
        "version": "1.1.1k"
      }
    }
  ]
}`)

	findings, err := ParseGrype(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "CVE-2021-36159" {
		t.Errorf("expected CVE-2021-36159, got %s", f1.RuleID)
	}
	if f1.Severity != "critical" {
		t.Errorf("expected critical, got %s", f1.Severity)
	}
	if f1.FilePath != "apk-tools@2.10.6-r0" {
		t.Errorf("expected apk-tools@2.10.6-r0, got %s", f1.FilePath)
	}
	if f1.Remediation != "Upgrade apk-tools to version 2.10.7-r0" {
		t.Errorf("unexpected remediation: %s", f1.Remediation)
	}

	f2 := findings[1]
	if f2.Severity != "high" {
		t.Errorf("expected high, got %s", f2.Severity)
	}
	if f2.Remediation != "" {
		t.Errorf("expected empty remediation for not-fixed, got %s", f2.Remediation)
	}
}

func TestParseGrype_Empty(t *testing.T) {
	findings, err := ParseGrype([]byte(`{"matches": []}`), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestParseGrype_InvalidJSON(t *testing.T) {
	_, err := ParseGrype([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
