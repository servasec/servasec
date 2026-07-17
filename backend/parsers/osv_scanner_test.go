package parsers

import (
	"testing"
)

func TestParseOSVScanner(t *testing.T) {
	data := []byte(`{
  "results": [
    {
      "source": {
        "path": "/project/go.mod",
        "type": "lockfile"
      },
      "packages": [
        {
          "package": {
            "name": "github.com/gogo/protobuf",
            "version": "1.3.1",
            "ecosystem": "Go"
          },
          "vulnerabilities": [
            {
              "id": "GHSA-c3h9-896r-86jm",
              "aliases": ["CVE-2021-3121"],
              "summary": "Protobuf DoS vulnerability",
              "details": "A denial of service vulnerability in protobuf.",
              "severity": []
            }
          ]
        }
      ]
    }
  ]
}`)

	findings, err := ParseOSVScanner(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}

	f := findings[0]
	if f.RuleID != "CVE-2021-3121" {
		t.Errorf("expected CVE-2021-3121 (alias preferred), got %s", f.RuleID)
	}
	if f.Title != "Protobuf DoS vulnerability" {
		t.Errorf("unexpected title: %s", f.Title)
	}
	if f.FilePath != "/project/go.mod (github.com/gogo/protobuf@1.3.1)" {
		t.Errorf("unexpected filePath: %s", f.FilePath)
	}
}

func TestParseOSVScanner_MultipleVulns(t *testing.T) {
	data := []byte(`{
  "results": [
    {
      "source": {"path": "Cargo.lock", "type": "lockfile"},
      "packages": [
        {
          "package": {"name": "regex", "version": "1.5.1", "ecosystem": "crates.io"},
          "vulnerabilities": [
            {"id": "GHSA-aaaa", "summary": "Vuln A", "details": ""},
            {"id": "GHSA-bbbb", "summary": "Vuln B", "details": ""}
          ]
        }
      ]
    }
  ]
}`)

	findings, err := ParseOSVScanner(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestParseOSVScanner_InvalidJSON(t *testing.T) {
	_, err := ParseOSVScanner([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
