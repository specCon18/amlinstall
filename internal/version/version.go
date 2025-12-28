package version

import (
	"strconv"
	"strings"
	"unicode"
)

// NormalizeTag strips a single leading "v" or "V" from a git tag for display.
//
// Examples:
//   - "v0.6.5" -> "0.6.5"
//   - "V1.2"   -> "1.2"
//   - "1.2"    -> "1.2"
func NormalizeTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if len(tag) > 1 && (tag[0] == 'v' || tag[0] == 'V') {
		return tag[1:]
	}
	return tag
}

// versionKey supports any number of numeric dot segments:
//   - "0.2.7.4"
//   - "1.2.3"
//   - "1.2"
//   - "1"
// and semver-like prerelease ordering:
//   - release > prerelease (same core)
//   - prerelease identifiers compared per semver rules (numeric < non-numeric, etc.)
//
// This mirrors the baseline behavior previously embedded in the TUI.
// It is intentionally conservative ("version-like" must start with a digit).

type versionKey struct {
	ok     bool
	core   []int
	hasPre bool
	pre    []string
}

func parseVersion(s string) versionKey {
	s = strings.TrimSpace(s)
	var k versionKey

	// Require leading digit to treat as version-like.
	if s == "" || !unicode.IsDigit(rune(s[0])) {
		return k
	}

	main := s
	pre := ""
	if i := strings.IndexByte(s, '-'); i >= 0 {
		main = s[:i]
		pre = s[i+1:]
		k.hasPre = true
	}

	coreParts := strings.Split(main, ".")
	if len(coreParts) == 0 {
		return versionKey{}
	}

	k.core = make([]int, 0, len(coreParts))
	for _, p := range coreParts {
		if p == "" {
			return versionKey{}
		}
		for _, r := range p {
			if !unicode.IsDigit(r) {
				return versionKey{}
			}
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return versionKey{}
		}
		k.core = append(k.core, v)
	}

	if k.hasPre && pre != "" {
		k.pre = strings.Split(pre, ".")
	}

	k.ok = true
	return k
}

func isNumericIdent(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return 0, false
		}
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return v, true
}

func cmpPrerelease(a, b []string) int {
	// -1 if a<b, 0 if equal, +1 if a>b (semver precedence rules)
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		ai := a[i]
		bi := b[i]

		ain, aNum := isNumericIdent(ai)
		bin, bNum := isNumericIdent(bi)

		switch {
		case aNum && bNum:
			if ain < bin {
				return -1
			}
			if ain > bin {
				return 1
			}
		case aNum && !bNum:
			// numeric < non-numeric
			return -1
		case !aNum && bNum:
			return 1
		default:
			if ai < bi {
				return -1
			}
			if ai > bi {
				return 1
			}
		}
	}
	// If equal prefix, shorter prerelease has lower precedence.
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

// Greater returns true if aDisp should sort ahead of bDisp in descending order.
//
// The comparison is semver-like when both values are "version-like" (start with a digit).
// Otherwise, it falls back to lexical descending ordering.
func Greater(aDisp, bDisp string) bool {
	a := parseVersion(aDisp)
	b := parseVersion(bDisp)

	// Prefer version-like values over non-version-like values.
	if a.ok && !b.ok {
		return true
	}
	if !a.ok && b.ok {
		return false
	}
	if !a.ok && !b.ok {
		// fallback: lexical descending
		return aDisp > bDisp
	}

	// Compare core numeric segments, treating missing segments as 0.
	n := len(a.core)
	if len(b.core) > n {
		n = len(b.core)
	}
	for i := 0; i < n; i++ {
		av := 0
		if i < len(a.core) {
			av = a.core[i]
		}
		bv := 0
		if i < len(b.core) {
			bv = b.core[i]
		}
		if av != bv {
			return av > bv
		}
	}

	// Same core: release > prerelease
	if a.hasPre != b.hasPre {
		return !a.hasPre && b.hasPre
	}
	if !a.hasPre && !b.hasPre {
		return false
	}

	// Both prerelease: higher prerelease wins
	return cmpPrerelease(a.pre, b.pre) > 0
}
