package parsers

import (
	"testing"
)

func TestParseGitleaks(t *testing.T) {
	data := []byte(`[
  {
    "Description": "AWS Access Key ID",
    "StartLine": 42,
    "EndLine": 42,
    "StartColumn": 1,
    "EndColumn": 41,
    "Match": "AKIAIOSFODNN7EXAMPLE",
    "Secret": "AKIAIOSFODNN7EXAMPLE",
    "File": "src/main.go",
    "Commit": "abc123def456",
    "Entropy": 4.5,
    "Author": "John Doe",
    "Email": "john@example.com",
    "Date": "2024-01-01T00:00:00Z",
    "Message": "added config",
    "Tags": ["aws", "access-key"],
    "RuleID": "aws-access-key",
    "Fingerprint": "abc123def456:src/main.go:42"
  },
  {
    "Description": "Generic Secret",
    "StartLine": 10,
    "EndLine": 10,
    "Match": "secret_value",
    "Secret": "secret_value",
    "File": "/etc/config.yaml",
    "Tags": ["generic"],
    "RuleID": "generic-secret",
    "Fingerprint": "abc:file:10"
  }
]`)

	findings, err := ParseGitleaks(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "aws-access-key" {
		t.Errorf("expected aws-access-key, got %s", f1.RuleID)
	}
	if f1.Severity != "critical" {
		t.Errorf("expected critical, got %s", f1.Severity)
	}
	if f1.Title != "AWS Access Key ID" {
		t.Errorf("expected 'AWS Access Key ID', got %s", f1.Title)
	}
	if f1.FilePath != "/src/main.go" {
		t.Errorf("expected /src/main.go, got %s", f1.FilePath)
	}
	if f1.LineStart == nil || *f1.LineStart != 42 {
		t.Errorf("expected line 42, got %v", f1.LineStart)
	}
	if f1.LineEnd == nil || *f1.LineEnd != 42 {
		t.Errorf("expected line 42, got %v", f1.LineEnd)
	}
	if f1.Description != "AWS Access Key ID: added config" {
		t.Errorf("unexpected description: %s", f1.Description)
	}

	f2 := findings[1]
	if f2.RuleID != "generic-secret" {
		t.Errorf("expected generic-secret, got %s", f2.RuleID)
	}
	if f2.Severity != "high" {
		t.Errorf("expected high, got %s", f2.Severity)
	}
	if f2.FilePath != "/etc/config.yaml" {
		t.Errorf("expected /etc/config.yaml, got %s", f2.FilePath)
	}
}

func TestParseGitleaks_InvalidJSON(t *testing.T) {
	_, err := ParseGitleaks([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseGitleaks_EmptyArray(t *testing.T) {
	findings, err := ParseGitleaks([]byte(`[]`), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}
