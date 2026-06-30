package parsers

import (
	"testing"
)

func TestParseNuclei_Lines(t *testing.T) {
	data := []byte(`{"template-id":"CVE-2021-44228","info":{"name":"Apache Log4j RCE","severity":"critical","description":"Log4j JNDI RCE","remediation":"Upgrade to 2.17.0","classification":{"cve-id":["CVE-2021-44228"],"cwe-id":["CWE-502"]}},"type":"http","host":"https://example.com","matched-at":"https://example.com/api?x=${jndi:ldap://x}"}
{"template-id":"tech-detect","info":{"name":"Nginx Detected","severity":"info"},"type":"http","host":"https://example.com","matched-at":"https://example.com/"}`)

	findings, err := ParseNuclei(data, "out.jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "CVE-2021-44228" {
		t.Errorf("expected RuleID CVE-2021-44228, got %s", f1.RuleID)
	}
	if f1.Title != "Apache Log4j RCE" {
		t.Errorf("expected title from info.name, got %s", f1.Title)
	}
	if f1.Severity != "critical" {
		t.Errorf("expected critical, got %s", f1.Severity)
	}
	if f1.CWEID != "CWE-502" {
		t.Errorf("expected CWE-502, got %s", f1.CWEID)
	}
	if f1.FilePath != "https://example.com/api?x=${jndi:ldap://x}" {
		t.Errorf("expected matched-at URL as FilePath, got %s", f1.FilePath)
	}
	if f1.Remediation != "Upgrade to 2.17.0" {
		t.Errorf("expected remediation, got %s", f1.Remediation)
	}

	f2 := findings[1]
	if f2.Severity != "info" {
		t.Errorf("expected info, got %s", f2.Severity)
	}
}

func TestParseNuclei_Array(t *testing.T) {
	data := []byte(`[
  {"template-id":"weak-cipher","info":{"name":"Weak Cipher","severity":"medium"},"matched-at":"https://t.example/login"}
]`)

	findings, err := ParseNuclei(data, "out.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != "medium" {
		t.Errorf("expected medium, got %s", findings[0].Severity)
	}
	if findings[0].FilePath != "https://t.example/login" {
		t.Errorf("expected matched-at, got %s", findings[0].FilePath)
	}
}

// Older nuclei versions emit cve-id/cwe-id as a bare string instead of an array.
// extractCWE must normalize both forms.
func TestParseNuclei_CWEAsString(t *testing.T) {
	data := []byte(`{"template-id":"xss","info":{"name":"Reflected XSS","severity":"high","classification":{"cwe-id":"CWE-79"}},"matched-at":"https://t.example/q"}`)

	findings, err := ParseNuclei(data, "out.jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].CWEID != "CWE-79" {
		t.Errorf("expected CWE-79 from string form, got %s", findings[0].CWEID)
	}
}

// "unknown" (and any non-standard) severity should default to info.
func TestParseNuclei_UnknownSeverityDefaultsToInfo(t *testing.T) {
	data := []byte(`{"template-id":"misc","info":{"name":"Misc","severity":"unknown"},"matched-at":"https://t.example/"}`)

	findings, err := ParseNuclei(data, "out.jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findings[0].Severity != "info" {
		t.Errorf("expected info default, got %s", findings[0].Severity)
	}
}

// When matched-at is absent, host is used as the location.
func TestParseNuclei_FallbackToHost(t *testing.T) {
	data := []byte(`{"template-id":"open-redirect","info":{"name":"Open Redirect","severity":"low"},"host":"https://t.example"}`)

	findings, err := ParseNuclei(data, "out.jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findings[0].FilePath != "https://t.example" {
		t.Errorf("expected host fallback, got %s", findings[0].FilePath)
	}
}

func TestParseNuclei_EmptyInput(t *testing.T) {
	if _, err := ParseNuclei([]byte("   \n  "), "out.jsonl"); err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseNuclei_InvalidJSON(t *testing.T) {
	if _, err := ParseNuclei([]byte(`not json`), "out.jsonl"); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDetectScannerType_Nuclei(t *testing.T) {
	// Auto-detection works on the JSON array export form (valid JSON as a whole).
	data := []byte(`[{"template-id":"tech-detect","matched-at":"https://example.com/","info":{"name":"x","severity":"info"}}]`)
	if got := DetectScannerType(data); got != "nuclei" {
		t.Errorf("expected nuclei, got %q", got)
	}
}
