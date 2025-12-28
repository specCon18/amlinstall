package version

import "testing"

func TestNormalizeTag(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"v0.6.5", "0.6.5"},
		{"V1.2", "1.2"},
		{"1.2.3", "1.2.3"},
		{" v2.0 ", "2.0"},
		{"v", "v"},
	}
	for _, tc := range cases {
		if got := NormalizeTag(tc.in); got != tc.want {
			t.Fatalf("NormalizeTag(%q)=%q; want %q", tc.in, got, tc.want)
		}
	}
}

func TestGreater_SemverCore(t *testing.T) {
	if !Greater("1.10.0", "1.2.9") {
		t.Fatalf("expected 1.10.0 > 1.2.9")
	}
	if Greater("0.6.3", "0.6.4") {
		t.Fatalf("expected 0.6.3 < 0.6.4")
	}
	if !Greater("0.2.7.4", "0.2.7.3") {
		t.Fatalf("expected 0.2.7.4 > 0.2.7.3")
	}
}

func TestGreater_Prerelease(t *testing.T) {
	if !Greater("1.0.0", "1.0.0-beta.1") {
		t.Fatalf("expected release > prerelease")
	}
	if !Greater("1.0.0-beta.2", "1.0.0-beta.1") {
		t.Fatalf("expected beta.2 > beta.1")
	}
	if !Greater("1.0.0-beta.1", "1.0.0-1") {
		// semver: numeric identifiers have lower precedence than non-numeric
		t.Fatalf("expected beta.1 > 1")
	}
}

func TestGreater_FallbackLexical(t *testing.T) {
	// Non-version-like values fall back to lexical descending.
	if !Greater("zzz", "aaa") {
		t.Fatalf("expected lexical desc")
	}
}
