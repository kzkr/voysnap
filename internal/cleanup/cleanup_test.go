package cleanup

import "testing"

func TestClean(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"leading space", " Hello world.", "Hello world."},
		{"collapse spaces", "too    many   spaces", "too many spaces"},
		{"trailing newline", "Hello world.\n", "Hello world."},
		{"empty", "   ", ""},
		{"already clean", "This is fine.", "This is fine."},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Clean(c.in); got != c.want {
				t.Errorf("Clean(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
