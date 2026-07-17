package parsers

import (
	"testing"
)

func TestParseBandit(t *testing.T) {
	data := []byte(`{
  "results": [
    {
      "test_id": "B105",
      "test_name": "hardcoded_password_string",
      "issue_text": "Possible hardcoded password: 'admin123'",
      "issue_severity": "MEDIUM",
      "issue_confidence": "MEDIUM",
      "filename": "app/config.py",
      "line_number": 12,
      "issue_cwe": {
        "id": "CWE-259",
        "link": "https://cwe.mitre.org/data/definitions/259.html"
      }
    },
    {
      "test_id": "B608",
      "test_name": "hardcoded_sql_expressions",
      "issue_text": "Possible SQL injection variable",
      "issue_severity": "HIGH",
      "issue_confidence": "LOW",
      "filename": "/src/db/query.py",
      "line_number": 45,
      "issue_cwe": {
        "id": "CWE-89"
      }
    }
  ],
  "metrics": {}
}`)

	findings, err := ParseBandit(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "B105" {
		t.Errorf("expected B105, got %s", f1.RuleID)
	}
	if f1.Severity != "medium" {
		t.Errorf("expected medium, got %s", f1.Severity)
	}
	if f1.CWEID != "CWE-259" {
		t.Errorf("expected CWE-259, got %s", f1.CWEID)
	}
	if f1.FilePath != "/app/config.py" {
		t.Errorf("expected /app/config.py, got %s", f1.FilePath)
	}
	if f1.LineStart == nil || *f1.LineStart != 12 {
		t.Errorf("expected line 12, got %v", f1.LineStart)
	}
	if f1.Title != "Possible hardcoded password: 'admin123'" {
		t.Errorf("unexpected title: %s", f1.Title)
	}

	f2 := findings[1]
	if f2.Severity != "high" {
		t.Errorf("expected high, got %s", f2.Severity)
	}
}

func TestParseBandit_Empty(t *testing.T) {
	findings, err := ParseBandit([]byte(`{"results": []}`), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestParseBandit_InvalidJSON(t *testing.T) {
	_, err := ParseBandit([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
