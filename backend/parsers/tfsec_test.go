package parsers

import (
	"testing"
)

func TestParseTfsec(t *testing.T) {
	data := []byte(`{
  "results": {
    "passed": [
      {
        "rule_id": "aws-s3-enable-bucket-encryption",
        "severity": "MEDIUM",
        "description": "S3 bucket has encryption enabled",
        "resolution": "",
        "location": {
          "filename": "s3/main.tf",
          "start_line": 1,
          "end_line": 10
        }
      }
    ],
    "failed": [
      {
        "rule_id": "aws-s3-enable-bucket-encryption",
        "severity": "HIGH",
        "description": "Bucket does not have encryption enabled",
        "resolution": "Enable AES-256 or CMK encryption for the bucket",
        "location": {
          "filename": "s3/data.tf",
          "start_line": 12,
          "end_line": 18
        }
      },
      {
        "rule_id": "aws-vpc-no-public-ingress-sgr",
        "severity": "CRITICAL",
        "description": "Security group allows public ingress",
        "resolution": "Restrict ingress to specific CIDR blocks",
        "location": {
          "filename": "vpc/main.tf",
          "start_line": 5,
          "end_line": 15
        }
      }
    ]
  }
}`)

	findings, err := ParseTfsec(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings (failed only), got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "aws-s3-enable-bucket-encryption" {
		t.Errorf("expected aws-s3-enable-bucket-encryption, got %s", f1.RuleID)
	}
	if f1.Severity != "high" {
		t.Errorf("expected high, got %s", f1.Severity)
	}
	if f1.FilePath != "/s3/data.tf" {
		t.Errorf("expected /s3/data.tf, got %s", f1.FilePath)
	}
	if f1.LineStart == nil || *f1.LineStart != 12 {
		t.Errorf("expected line 12, got %v", f1.LineStart)
	}
	if f1.Remediation != "Enable AES-256 or CMK encryption for the bucket" {
		t.Errorf("unexpected remediation: %s", f1.Remediation)
	}

	f2 := findings[1]
	if f2.Severity != "critical" {
		t.Errorf("expected critical, got %s", f2.Severity)
	}
}

func TestParseTfsec_Empty(t *testing.T) {
	findings, err := ParseTfsec([]byte(`{"results": {"passed": [], "failed": []}}`), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestParseTfsec_InvalidJSON(t *testing.T) {
	_, err := ParseTfsec([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
