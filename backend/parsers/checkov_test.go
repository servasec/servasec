package parsers

import (
	"testing"
)

func TestParseCheckov(t *testing.T) {
	data := []byte(`{
  "results": {
    "failed_checks": [
      {
        "check_id": "CKV_AWS_18",
        "check": {
          "id": "CKV_AWS_18",
          "name": "Ensure the S3 bucket has access logging enabled",
          "description": "S3 bucket access logging should be enabled",
          "severity": "LOW"
        },
        "file_path": "terraform/s3.tf",
        "file_line_range": [1, 20],
        "resource": "aws_s3_bucket.data_bucket"
      },
      {
        "check_id": "CKV_AWS_21",
        "check": {
          "id": "CKV_AWS_21",
          "name": "Ensure all data stored in S3 have versioning enabled",
          "description": "",
          "severity": "HIGH"
        },
        "file_path": "s3/main.tf",
        "file_line_range": [5],
        "resource": "aws_s3_bucket.assets"
      }
    ],
    "passed_checks": [
      {
        "check_id": "CKV_AWS_20",
        "check": {
          "id": "CKV_AWS_20",
          "name": "S3 Bucket has an ACL defined which allows public access",
          "severity": "HIGH"
        },
        "file_path": "terraform/s3.tf",
        "file_line_range": [1, 20],
        "resource": "aws_s3_bucket.data_bucket"
      }
    ],
    "skipped_checks": []
  }
}`)

	findings, err := ParseCheckov(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings (failed only), got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "CKV_AWS_18" {
		t.Errorf("expected CKV_AWS_18, got %s", f1.RuleID)
	}
	if f1.Severity != "low" {
		t.Errorf("expected low, got %s", f1.Severity)
	}
	if f1.FilePath != "/terraform/s3.tf" {
		t.Errorf("expected /terraform/s3.tf, got %s", f1.FilePath)
	}
	if f1.LineStart == nil || *f1.LineStart != 1 {
		t.Errorf("expected line 1, got %v", f1.LineStart)
	}
	if f1.LineEnd == nil || *f1.LineEnd != 20 {
		t.Errorf("expected line 20, got %v", f1.LineEnd)
	}

	f2 := findings[1]
	if f2.Severity != "high" {
		t.Errorf("expected high, got %s", f2.Severity)
	}
	if f2.LineEnd != nil {
		t.Errorf("expected nil lineEnd for single line range, got %v", f2.LineEnd)
	}
}

func TestParseCheckov_Empty(t *testing.T) {
	findings, err := ParseCheckov([]byte(`{"results": {"failed_checks": [], "passed_checks": []}}`), "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(findings))
	}
}

func TestParseCheckov_InvalidJSON(t *testing.T) {
	_, err := ParseCheckov([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
