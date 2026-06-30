package parsers

import (
	"testing"
)

func TestParseSarif_Basic(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Semgrep",
          "rules": [
            {
              "id": "python.lang.correctness.path-traversal",
              "properties": {
                "tags": ["cwe-22", "security"]
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "python.lang.correctness.path-traversal",
          "ruleIndex": 0,
          "level": "error",
          "message": {
            "text": "Possible path traversal found in join()"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "file:///app/src/handler.py"
                },
                "region": {
                  "startLine": 42,
                  "endLine": 45
                }
              }
            }
          ],
          "properties": {
            "tags": ["cwe-22", "security"]
          }
        },
        {
          "ruleId": "python.lang.best-practice.logging",
          "ruleIndex": -1,
          "level": "note",
          "message": {
            "text": "Use of print() instead of logging"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "file:///app/src/helper.py"
                },
                "region": {
                  "startLine": 10
                }
              }
            }
          ]
        },
        {
          "ruleId": "python.lang.best-practice.todo",
          "ruleIndex": -1,
          "level": "none",
          "message": {
            "text": "TODO comment found"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "file:///app/src/main.py"
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`)

	findings, err := ParseSarif(data, "out.sarif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "python.lang.correctness.path-traversal" {
		t.Errorf("expected RuleID python.lang.correctness.path-traversal, got %s", f1.RuleID)
	}
	if f1.Title != "Possible path traversal found in join()" {
		t.Errorf("expected title from message.text, got %s", f1.Title)
	}
	if f1.Severity != "high" {
		t.Errorf("expected high (from error level), got %s", f1.Severity)
	}
	if f1.CWEID != "CWE-22" {
		t.Errorf("expected CWE-22 from properties.tags, got %s", f1.CWEID)
	}
	if f1.FilePath != "/app/src/handler.py" {
		t.Errorf("expected /app/src/handler.py, got %s", f1.FilePath)
	}
	if *f1.LineStart != 42 {
		t.Errorf("expected LineStart 42, got %d", *f1.LineStart)
	}
	if *f1.LineEnd != 45 {
		t.Errorf("expected LineEnd 45, got %d", *f1.LineEnd)
	}

	f2 := findings[1]
	if f2.Severity != "low" {
		t.Errorf("expected low (from note level), got %s", f2.Severity)
	}
	if *f2.LineStart != 10 {
		t.Errorf("expected LineStart 10, got %d", *f2.LineStart)
	}
	if f2.LineEnd != nil {
		t.Errorf("expected nil LineEnd, got %d", *f2.LineEnd)
	}

	f3 := findings[2]
	if f3.Severity != "info" {
		t.Errorf("expected info (from none level), got %s", f3.Severity)
	}
	if f3.LineStart != nil {
		t.Errorf("expected nil LineStart for missing region, got %d", *f3.LineStart)
	}
}

func TestParseSarif_PropertiesSeverity(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "test"
        }
      },
      "results": [
        {
          "ruleId": "R1",
          "level": "warning",
          "message": {"text": "properties.severity wins"},
          "properties": {
            "severity": "CRITICAL",
            "priority": "low"
          },
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "f.py"}, "region": {"startLine": 1}}}]
        },
        {
          "ruleId": "R2",
          "level": "none",
          "message": {"text": "properties.priority fallback"},
          "properties": {
            "priority": "high"
          },
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "f.py"}, "region": {"startLine": 2}}}]
        },
        {
          "ruleId": "R3",
          "level": "error",
          "message": {"text": "level fallback"},
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "f.py"}, "region": {"startLine": 3}}}]
        }
      ]
    }
  ]
}`)

	findings, err := ParseSarif(data, "out.sarif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}

	if findings[0].Severity != "high" {
		t.Errorf("expected high (from CRITICAL), got %s", findings[0].Severity)
	}
	if findings[1].Severity != "high" {
		t.Errorf("expected high (from priority=high), got %s", findings[1].Severity)
	}
	if findings[2].Severity != "high" {
		t.Errorf("expected high (from error level), got %s", findings[2].Severity)
	}
}

func TestParseSarif_RelativePath(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {"name": "test"}
      },
      "results": [
        {
          "ruleId": "R1",
          "level": "warning",
          "message": {"text": "relative path"},
          "locations": [{
            "physicalLocation": {
              "artifactLocation": {"uri": "src/main.go"},
              "region": {"startLine": 1}
            }
          }]
        }
      ]
    }
  ]
}`)

	findings, err := ParseSarif(data, "out.sarif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].FilePath != "/src/main.go" {
		t.Errorf("expected /src/main.go (prepended /), got %s", findings[0].FilePath)
	}
}

func TestParseSarif_EmptyBodyText(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {"driver": {"name": "test"}},
      "results": [
        {
          "ruleId": "R1",
          "level": "warning",
          "message": {"text": ""},
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "f.py"}, "region": {"startLine": 1}}}]
        }
      ]
    }
  ]
}`)

	findings, err := ParseSarif(data, "out.sarif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Title != "R1" {
		t.Errorf("expected RuleID as fallback title, got %s", findings[0].Title)
	}
}

func TestParseSarif_NoRuns(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": []
}`)

	_, err := ParseSarif(data, "empty.sarif")
	if err == nil {
		t.Fatal("expected error for empty runs, got nil")
	}
}

func TestParseSarif_InvalidJSON(t *testing.T) {
	_, err := ParseSarif([]byte(`not json`), "bad.sarif")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestParseSarif_NoResults(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {"driver": {"name": "test"}},
      "results": []
    }
  ]
}`)

	findings, err := ParseSarif(data, "empty.sarif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestExtractSarifToolName(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Semgrep"
        }
      },
      "results": []
    }
  ]
}`)

	name := ExtractSarifToolName(data)
	if name != "semgrep" {
		t.Errorf("expected lowercased 'semgrep', got '%s'", name)
	}
}

func TestExtractSarifToolName_NoRuns(t *testing.T) {
	name := ExtractSarifToolName([]byte(`{"runs":[]}`))
	if name != "" {
		t.Errorf("expected empty for no runs, got '%s'", name)
	}
}

func TestExtractSarifToolName_Opengrep(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Opengrep OSS"
        }
      },
      "results": []
    }
  ]
}`)

	name := ExtractSarifToolName(data)
	if name != "semgrep" {
		t.Errorf("expected aliased 'semgrep' for Opengrep OSS, got '%s'", name)
	}
}

func TestParseSarif_SeverityFallbackToRuleDefault(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "test",
          "rules": [
            {
              "id": "error-rule",
              "defaultConfiguration": {"level": "error"},
              "properties": {"tags": []}
            },
            {
              "id": "warning-rule",
              "defaultConfiguration": {"level": "warning"},
              "properties": {"tags": []}
            },
            {
              "id": "note-rule",
              "defaultConfiguration": {"level": "note"},
              "properties": {"tags": []}
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "error-rule",
          "message": {"text": "No level on result, rule says error"},
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "f.py"}, "region": {"startLine": 1}}}]
        },
        {
          "ruleId": "warning-rule",
          "message": {"text": "No level on result, rule says warning"},
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "f.py"}, "region": {"startLine": 2}}}]
        },
        {
          "ruleId": "note-rule",
          "message": {"text": "No level on result, rule says note"},
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "f.py"}, "region": {"startLine": 3}}}]
        }
      ]
    }
  ]
}`)

	findings, err := ParseSarif(data, "out.sarif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 3 {
		t.Fatalf("expected 3 findings, got %d", len(findings))
	}
	if findings[0].Severity != "high" {
		t.Errorf("expected high (rule default error), got %s", findings[0].Severity)
	}
	if findings[1].Severity != "medium" {
		t.Errorf("expected medium (rule default warning), got %s", findings[1].Severity)
	}
	if findings[2].Severity != "low" {
		t.Errorf("expected low (rule default note), got %s", findings[2].Severity)
	}
}

func TestParseSarif_SeverityFallbackMixed(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "test",
          "rules": [
            {
              "id": "explicit-rule",
              "defaultConfiguration": {"level": "note"}
            },
            {
              "id": "fallback-rule",
              "defaultConfiguration": {"level": "error"}
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "explicit-rule",
          "level": "warning",
          "message": {"text": "Explicit level beats rule default"},
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "f.py"}, "region": {"startLine": 1}}}]
        },
        {
          "ruleId": "fallback-rule",
          "message": {"text": "No level, uses rule default"},
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "f.py"}, "region": {"startLine": 2}}}]
        }
      ]
    }
  ]
}`)

	findings, err := ParseSarif(data, "out.sarif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	if findings[0].Severity != "medium" {
		t.Errorf("expected medium (explicit warning beats rule note), got %s", findings[0].Severity)
	}
	if findings[1].Severity != "high" {
		t.Errorf("expected high (fallback to rule error), got %s", findings[1].Severity)
	}
}

func TestDetectScannerType_Sarif(t *testing.T) {
	t.Run("with $schema", func(t *testing.T) {
		data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {"driver": {"name": "Semgrep"}},
      "results": [{"ruleId": "R1", "level": "error", "message": {"text": "test"}, "locations": []}]
    }
  ]
}`)

		if got := DetectScannerType(data); got != "sarif" {
			t.Errorf("expected sarif, got %s", got)
		}
	})

	t.Run("without $schema (Opengrep-style)", func(t *testing.T) {
		data := []byte(`{
  "version": "2.1.0",
  "runs": [
    {
      "tool": {"driver": {"name": "Opengrep OSS"}},
      "results": [{"ruleId": "R1", "message": {"text": "test"}, "locations": []}]
    }
  ]
}`)

		if got := DetectScannerType(data); got != "sarif" {
			t.Errorf("expected sarif, got %s", got)
		}
	})
}

func TestParseSarif_CWEfromRuleTags(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "CodeQL",
          "rules": [
            {
              "id": "py/path-injection",
              "properties": {
                "tags": ["cwe-22", "security", "external/cwe/cwe-022"]
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "py/path-injection",
          "ruleIndex": 0,
          "level": "error",
          "message": {"text": "Path injection"},
          "properties": {"tags": []},
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "f.py"}, "region": {"startLine": 1}}}]
        }
      ]
    }
  ]
}`)

	findings, err := ParseSarif(data, "out.sarif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].CWEID != "CWE-22" {
		t.Errorf("expected CWE-22 from rule tags, got %s", findings[0].CWEID)
	}
}

func TestParseSarif_MultiRun(t *testing.T) {
	data := []byte(`{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {"driver": {"name": "Semgrep"}},
      "results": [
        {
          "ruleId": "R1",
          "level": "error",
          "message": {"text": "Finding from run 1"},
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "a.py"}, "region": {"startLine": 1}}}]
        }
      ]
    },
    {
      "tool": {"driver": {"name": "CodeQL"}},
      "results": [
        {
          "ruleId": "R2",
          "level": "warning",
          "message": {"text": "Finding from run 2"},
          "locations": [{"physicalLocation": {"artifactLocation": {"uri": "b.py"}, "region": {"startLine": 5}}}]
        }
      ]
    }
  ]
}`)

	findings, err := ParseSarif(data, "out.sarif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	if findings[0].RuleID != "R1" || findings[1].RuleID != "R2" {
		t.Errorf("unexpected findings ordering: R1=%s R2=%s", findings[0].RuleID, findings[1].RuleID)
	}
}

func TestSarifSeverity_Mapping(t *testing.T) {
	tests := []struct {
		level    string
		severity string
		priority string
		want     string
	}{
		{"error", "", "", "high"},
		{"warning", "", "", "medium"},
		{"note", "", "", "low"},
		{"none", "", "", "info"},
		{"", "", "", "info"},
		{"warning", "CRITICAL", "", "high"},
		{"none", "", "high", "high"},
		{"error", "low", "", "low"},
	}

	for _, tt := range tests {
		got := sarifSeverity(tt.level, sarifProperties{Severity: tt.severity, Priority: tt.priority})
		if got != tt.want {
			t.Errorf("sarifSeverity(level=%q, severity=%q, priority=%q) = %q, want %q",
				tt.level, tt.severity, tt.priority, got, tt.want)
		}
	}
}

func TestSarifCWE_TagFormats(t *testing.T) {
	tests := []struct {
		tags []string
		want string
	}{
		{[]string{"cwe-79"}, "CWE-79"},
		{[]string{"CWE-22"}, "CWE-22"},
		{[]string{"CWE/22"}, "CWE-22"},
		{[]string{"cwe:502"}, "CWE-502"},
		{[]string{"security", "external/cwe/cwe-022"}, "CWE-022"},
		{[]string{"security"}, ""},
		{nil, ""},
	}

	for _, tt := range tests {
		got := sarifCWE(tt.tags)
		if got != tt.want {
			t.Errorf("sarifCWE(%v) = %q, want %q", tt.tags, got, tt.want)
		}
	}
}

func TestSarifFilePath_Formats(t *testing.T) {
	tests := []struct {
		uri  string
		want string
	}{
		{"file:///home/user/project/main.go", "/home/user/project/main.go"},
		{"src/relative.go", "/src/relative.go"},
		{"/abs/path.go", "/abs/path.go"},
		{"file:///C:/Users/test/file.cs", "/C:/Users/test/file.cs"},
	}

	for _, tt := range tests {
		got := sarifFilePath(tt.uri)
		if got != tt.want {
			t.Errorf("sarifFilePath(%q) = %q, want %q", tt.uri, got, tt.want)
		}
	}
}
