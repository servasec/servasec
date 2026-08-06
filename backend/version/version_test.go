package version

import "testing"

func TestString(t *testing.T) {
	cases := []struct{ in, want string }{
		{"v2.4.3", "2.4.3"},
		{"2.4.3", "2.4.3"},
		{"develop", "develop"},
		{"v0.3.0-114-g9737dd6", "0.3.0-114-g9737dd6"},
	}
	for _, c := range cases {
		Version = c.in
		if got := NormalizeVersion(); got != c.want {
			t.Fatalf("String() for %q = %q, want %q", c.in, got, c.want)
		}
	}
}
