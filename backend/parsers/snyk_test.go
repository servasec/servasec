package parsers

import (
	"testing"
)

func TestParseSnyk(t *testing.T) {
	data := []byte(`{
  "vulnerabilities": [
    {
      "id": "npm:lodash:20180130",
      "title": "Prototype Pollution",
      "severity": "high",
      "packageName": "lodash",
      "version": "4.17.4",
      "description": "Prototype Pollution in lodash.",
      "identifiers": {
        "CWE": ["CWE-1321"],
        "CVE": ["CVE-2018-16487"]
      },
      "upgradePath": ["", "4.17.5"],
      "isUpgradable": true
    },
    {
      "id": "npm:minimist:20200224",
      "title": "Prototype Pollution",
      "severity": "medium",
      "packageName": "minimist",
      "version": "0.0.8",
      "description": "",
      "identifiers": {
        "CWE": [],
        "CVE": []
      },
      "upgradePath": [],
      "isUpgradable": false
    }
  ],
  "packageManager": "npm"
}`)

	findings, err := ParseSnyk(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "npm:lodash:20180130" {
		t.Errorf("expected npm:lodash:20180130, got %s", f1.RuleID)
	}
	if f1.Severity != "high" {
		t.Errorf("expected high, got %s", f1.Severity)
	}
	if f1.CWEID != "CWE-1321" {
		t.Errorf("expected CWE-1321, got %s", f1.CWEID)
	}
	if f1.Remediation != "Upgrade lodash to 4.17.5" {
		t.Errorf("unexpected remediation: %s", f1.Remediation)
	}
	if f1.FilePath != "lodash@4.17.4" {
		t.Errorf("expected lodash@4.17.4, got %s", f1.FilePath)
	}

	f2 := findings[1]
	if f2.Severity != "medium" {
		t.Errorf("expected medium, got %s", f2.Severity)
	}
	if f2.Remediation != "" {
		t.Errorf("expected empty remediation for non-upgradable, got %s", f2.Remediation)
	}
}

func TestParseSnyk_Empty(t *testing.T) {
	findings, err := ParseSnyk([]byte(`{"vulnerabilities": []}`), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestParseSnyk_InvalidJSON(t *testing.T) {
	_, err := ParseSnyk([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
