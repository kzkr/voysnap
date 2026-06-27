package snippets

import "testing"

func TestExpand(t *testing.T) {
	m := map[string]string{
		"my email":  "hello@kzkr.dev",
		"my handle": "kzkr",
	}
	cases := []struct {
		in, want string
	}{
		{"Send it to my email please.", "Send it to hello@kzkr.dev please."},
		{"MY EMAIL is best", "hello@kzkr.dev is best"}, // case-insensitive
		{"no triggers here", "no triggers here"},
		{"ping my handle now", "ping kzkr now"},
		{"myemail should not match", "myemail should not match"}, // word boundary
	}
	for _, c := range cases {
		if got := Expand(c.in, m); got != c.want {
			t.Errorf("Expand(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestExpandEmpty(t *testing.T) {
	if got := Expand("unchanged", nil); got != "unchanged" {
		t.Errorf("got %q", got)
	}
}
