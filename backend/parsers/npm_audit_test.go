package parsers

import (
	"testing"
)

func TestParseNpmAudit(t *testing.T) {
	data := []byte(`{
  "vulnerabilities": {
    "lodash": {
      "name": "lodash",
      "severity": "high",
      "range": "<4.17.5",
      "via": [
        {
          "title": "Prototype Pollution in lodash",
          "url": "https://npmjs.com/advisories/1086",
          "range": "<4.17.5"
        }
      ],
      "fixAvailable": {
        "name": "lodash",
        "version": "4.17.5",
        "isSemVerMajor": false
      }
    },
    "minimist": {
      "name": "minimist",
      "severity": "critical",
      "range": "<0.2.4",
      "via": [
        "Prototype Pollution in minimist"
      ],
      "fixAvailable": false
    }
  },
  "dependencyCount": 50,
  "packageManager": "npm"
}`)

	findings, err := ParseNpmAudit(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	// Map iteration order is not guaranteed, so find by filepath
	var lodashFinding, minimistFinding *FindingInput
	for i := range findings {
		switch {
		case findings[i].FilePath == "lodash@<4.17.5":
			lodashFinding = &findings[i]
		case findings[i].FilePath == "minimist@<0.2.4":
			minimistFinding = &findings[i]
		}
	}
	if lodashFinding == nil {
		t.Fatal("lodash finding not found")
	}
	if minimistFinding == nil {
		t.Fatal("minimist finding not found")
	}

	if lodashFinding.Severity != "high" {
		t.Errorf("expected high, got %s", lodashFinding.Severity)
	}
	if lodashFinding.Remediation != "Upgrade lodash to 4.17.5" {
		t.Errorf("unexpected remediation: %s", lodashFinding.Remediation)
	}

	if minimistFinding.Severity != "critical" {
		t.Errorf("expected critical, got %s", minimistFinding.Severity)
	}
	if minimistFinding.Remediation != "" {
		t.Errorf("expected empty remediation for non-fixable, got %s", minimistFinding.Remediation)
	}
}

func TestParseNpmAudit_Empty(t *testing.T) {
	findings, err := ParseNpmAudit([]byte(`{"vulnerabilities": {}}`), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestParseNpmAudit_InvalidJSON(t *testing.T) {
	_, err := ParseNpmAudit([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
