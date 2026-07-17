package parsers

import (
	"testing"
)

func TestParseGosec(t *testing.T) {
	data := []byte(`{
  "Issues": [
    {
      "severity": "HIGH",
      "confidence": "HIGH",
      "rule_id": "G101",
      "details": "Potential hardcoded credentials",
      "file": "/app/main.go",
      "code": "password = \"secret123\"",
      "line": "42",
      "cwe": {
        "ID": "CWE-798"
      }
    },
    {
      "severity": "MEDIUM",
      "confidence": "LOW",
      "rule_id": "G104",
      "details": "Errors unhandled",
      "file": "handler.go",
      "code": "resp, err := http.Get(url)",
      "line": "15",
      "cwe": null
    }
  ],
  "Stats": {
    "files": 2,
    "lines": 100,
    "nosec": 0,
    "found": 2
  }
}`)

	findings, err := ParseGosec(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "G101" {
		t.Errorf("expected G101, got %s", f1.RuleID)
	}
	if f1.Severity != "high" {
		t.Errorf("expected high, got %s", f1.Severity)
	}
	if f1.CWEID != "CWE-798" {
		t.Errorf("expected CWE-798, got %s", f1.CWEID)
	}
	if f1.FilePath != "/app/main.go" {
		t.Errorf("expected /app/main.go, got %s", f1.FilePath)
	}
	if f1.LineStart == nil || *f1.LineStart != 42 {
		t.Errorf("expected line 42, got %v", f1.LineStart)
	}

	f2 := findings[1]
	if f2.FilePath != "/handler.go" {
		t.Errorf("expected /handler.go, got %s", f2.FilePath)
	}
	if f2.CWEID != "" {
		t.Errorf("expected empty CWE, got %s", f2.CWEID)
	}
}

func TestParseGosec_Empty(t *testing.T) {
	findings, err := ParseGosec([]byte(`{"Issues": []}`), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestParseGosec_InvalidJSON(t *testing.T) {
	_, err := ParseGosec([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
