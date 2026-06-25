package services

import "testing"

func TestDedupeHash_StableAnd64Hex(t *testing.T) {
	l := 12
	a := DedupeHash("nuclei", "rule-1", "/a.go", &l, "high")
	b := DedupeHash("nuclei", "rule-1", "/a.go", &l, "high")
	if a != b {
		t.Fatalf("hash not stable: %s != %s", a, b)
	}
	if len(a) != 64 {
		t.Fatalf("expected 64 hex chars, got %d", len(a))
	}
}

func TestDedupeHash_DiffersByEachField(t *testing.T) {
	l := 12
	base := DedupeHash("nuclei", "rule-1", "/a.go", &l, "high")
	other := 13
	cases := map[string]string{
		"scanner":  DedupeHash("semgrep", "rule-1", "/a.go", &l, "high"),
		"rule":     DedupeHash("nuclei", "rule-2", "/a.go", &l, "high"),
		"file":     DedupeHash("nuclei", "rule-1", "/b.go", &l, "high"),
		"line":     DedupeHash("nuclei", "rule-1", "/a.go", &other, "high"),
		"severity": DedupeHash("nuclei", "rule-1", "/a.go", &l, "low"),
	}
	for field, h := range cases {
		if h == base {
			t.Errorf("changing %s should change the hash", field)
		}
	}
}

func TestDedupeHash_NilLine(t *testing.T) {
	a := DedupeHash("nuclei", "r", "/a", nil, "info")
	if a != DedupeHash("nuclei", "r", "/a", nil, "info") {
		t.Error("nil-line hash not stable")
	}
	zero := 0
	if a == DedupeHash("nuclei", "r", "/a", &zero, "info") {
		t.Error("nil line should be distinct from line 0")
	}
}
