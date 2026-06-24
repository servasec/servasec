package parsers

import (
	"testing"
)

func TestParseTrufflehog_Lines(t *testing.T) {
	data := []byte(`{"SourceMetadata":{"Data":{"Filesystem":{"file":"src/config.json"}}},"SourceType":"filesystem","SourceName":"trufflehog","DetectorType":0,"DetectorName":"AWS","Verified":false,"Raw":"AKIAIOSFODNN7EXAMPLE","Redacted":"AKIAIOSFODNN7EX****","Commit":""}
{"SourceMetadata":{"Data":{"Git":{"commit":"abc123","file":"src/.env"}}},"SourceType":"git","SourceName":"trufflehog","DetectorType":1,"DetectorName":"GitHub Token","Verified":true,"Raw":"ghp_xxxxxxxxxxxxxxxxxxxx","Redacted":"ghp_xxxx****","Commit":"abc123","FileMetadata":{"Filename":"src/.env"}}`)

	findings, err := ParseTrufflehog(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "AWS" {
		t.Errorf("expected AWS, got %s", f1.RuleID)
	}
	if f1.Severity != "medium" {
		t.Errorf("expected medium (unverified), got %s", f1.Severity)
	}
	if f1.FilePath != "src/config.json" {
		t.Errorf("expected src/config.json, got %s", f1.FilePath)
	}

	f2 := findings[1]
	if f2.RuleID != "GitHub Token" {
		t.Errorf("expected GitHub Token, got %s", f2.RuleID)
	}
	if f2.Severity != "high" {
		t.Errorf("expected high (verified), got %s", f2.Severity)
	}
	if f2.FilePath != "src/.env" {
		t.Errorf("expected src/.env, got %s", f2.FilePath)
	}
}

func TestParseTrufflehog_Array(t *testing.T) {
	data := []byte(`[
  {
    "SourceMetadata":{"Data":{"Filesystem":{"file":"src/key.txt"}}},
    "DetectorName":"AWS",
    "DetectorType":0,
    "Verified":true,
    "Raw":"AKIAIOSFODNN7EXAMPLE",
    "Redacted":"AKIAIOSFODNN7EX****"
  }
]`)

	findings, err := ParseTrufflehog(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Severity != "high" {
		t.Errorf("expected high, got %s", findings[0].Severity)
	}
}

func TestParseTrufflehog_EmptyLine(t *testing.T) {
	data := []byte("\n\n")
	_, err := ParseTrufflehog(data, "test.json")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseTrufflehog_InvalidJSON(t *testing.T) {
	data := []byte(`not json`)
	_, err := ParseTrufflehog(data, "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseTrufflehog_EmptyLines(t *testing.T) {
	data := []byte("")
	_, err := ParseTrufflehog(data, "test.json")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseTrufflehog_S3Path(t *testing.T) {
	data := []byte(`{"SourceMetadata":{"Data":{"S3":{"bucket":"my-bucket","key":"config/secret.txt"}}},"DetectorName":"AWS","Verified":true}`)

	findings, err := ParseTrufflehog(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].FilePath != "s3://my-bucket/config/secret.txt" {
		t.Errorf("expected s3 path, got %s", findings[0].FilePath)
	}
}
