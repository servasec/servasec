package parsers

import (
	"testing"
)

func TestParseKubescape(t *testing.T) {
	data := []byte(`{
  "results": [
    {
      "resources": [
        {
          "name": "nginx-deployment",
          "kind": "Deployment",
          "namespace": "default"
        },
        {
          "name": "api-service",
          "kind": "Service",
          "namespace": "production"
        }
      ],
      "controls": [
        {
          "controlID": "C-0001",
          "name": "Allow privilege escalation",
          "severity": "high",
          "description": "Containers should not allow privilege escalation",
          "failedResources": 2
        },
        {
          "controlID": "C-0002",
          "name": "Root container",
          "severity": "medium",
          "description": "Containers should not run as root",
          "failedResources": 0
        }
      ]
    }
  ]
}`)

	findings, err := ParseKubescape(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings (only failed controls × resources), got %d", len(findings))
	}

	f1 := findings[0]
	if f1.RuleID != "C-0001" {
		t.Errorf("expected C-0001, got %s", f1.RuleID)
	}
	if f1.Severity != "high" {
		t.Errorf("expected high, got %s", f1.Severity)
	}
	if f1.FilePath != "default/Deployment/nginx-deployment" {
		t.Errorf("expected default/Deployment/nginx-deployment, got %s", f1.FilePath)
	}

	f2 := findings[1]
	if f2.FilePath != "production/Service/api-service" {
		t.Errorf("expected production/Service/api-service, got %s", f2.FilePath)
	}
}

func TestParseKubescape_NoFailedResources(t *testing.T) {
	data := []byte(`{
  "results": [
    {
      "resources": [{"name": "pod-1", "kind": "Pod", "namespace": "ns"}],
      "controls": [
        {
          "controlID": "C-0001",
          "name": "Test",
          "severity": "high",
          "description": "desc",
          "failedResources": 0
        }
      ]
    }
  ]
}`)

	findings, err := ParseKubescape(data, "test.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when failedResources=0, got %d", len(findings))
	}
}

func TestParseKubescape_InvalidJSON(t *testing.T) {
	_, err := ParseKubescape([]byte(`not json`), "test.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
