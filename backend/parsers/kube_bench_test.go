package parsers

import (
	"testing"
)

func TestParseKubeBench(t *testing.T) {
	data := []byte(`{
  "controls": [
    {
      "id": "1",
      "text": "Master Node Security Configuration",
      "node_type": "master",
      "version": "1.5",
      "tests": [
        {
          "section": "1.1",
          "desc": "Master Node Configuration Files",
          "results": [
            {
              "test_number": "1.1.1",
              "test_desc": "Ensure that the API server pod specification file permissions are set to 644 or more restrictive",
              "status": "FAIL",
              "remediation": "Set file permissions to 644"
            },
            {
              "test_number": "1.1.2",
              "test_desc": "Ensure that the API server pod specification file ownership is set to root:root",
              "status": "PASS",
              "remediation": ""
            }
          ]
        }
      ]
    },
    {
      "id": "4",
      "text": "Worker Node Security Configuration",
      "node_type": "node",
      "version": "1.5",
      "tests": [
        {
          "section": "4.2",
          "desc": "Kubelet",
          "results": [
            {
              "test_number": "4.2.1",
              "test_desc": "Ensure that the --anonymous-auth argument is set to false",
              "status": "FAIL",
              "remediation": "Set --anonymous-auth=false"
            },
            {
              "test_number": "4.2.2",
              "test_desc": "Ensure that the --authorization-mode argument is not set to AlwaysAllow",
              "status": "WARN",
              "remediation": ""
            }
          ]
        }
      ]
    }
  ]
}`)

	findings, err := ParseKubeBench(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings (FAIL only), got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "1.1.1" {
		t.Errorf("expected 1.1.1, got %s", f1.RuleID)
	}
	if f1.Severity != "high" {
		t.Errorf("expected high for master node, got %s", f1.Severity)
	}
	if f1.FilePath != "kube-bench:master:1" {
		t.Errorf("expected kube-bench:master:1, got %s", f1.FilePath)
	}
	if f1.Remediation != "Set file permissions to 644" {
		t.Errorf("unexpected remediation: %s", f1.Remediation)
	}

	f2 := findings[1]
	if f2.RuleID != "4.2.1" {
		t.Errorf("expected 4.2.1, got %s", f2.RuleID)
	}
	if f2.Severity != "medium" {
		t.Errorf("expected medium for worker node, got %s", f2.Severity)
	}
}

func TestParseKubeBench_NoFailures(t *testing.T) {
	data := []byte(`{
  "controls": [
    {
      "id": "1",
      "text": "Master",
      "node_type": "master",
      "tests": [
        {
          "section": "1.1",
          "results": [
            {"test_number": "1.1.1", "test_desc": "Test", "status": "PASS", "remediation": ""}
          ]
        }
      ]
    }
  ]
}`)

	findings, err := ParseKubeBench(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when all PASS, got %d", len(findings))
	}
}

func TestParseKubeBench_InvalidJSON(t *testing.T) {
	_, err := ParseKubeBench([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
