package acctest

import "testing"

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"26.1", "26.7", -1},
		{"26.1.6", "26.7", -1},
		{"26.7", "26.7", 0},
		{"26.7.9", "26.7", 1},
		{"26.7_1", "26.7", 0},
		{"27.1", "26.7", 1},
		{"25.7.11_2", "26.1", -1},
	}
	for _, c := range cases {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
