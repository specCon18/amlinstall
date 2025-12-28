package tui

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"automelonloaderinstallergo/internal/ghrel"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
)

type tagsLoadedMsg struct {
	tags []string
}

type tagsErrMsg struct {
	err error
}

type downloadDoneMsg struct {
	out string
}

type downloadErrMsg struct {
	err error
}

const focusCount = int(focusToken) + 1

func normalizeTag(tag string) string {
	tag = strings.TrimSpace(tag)
	if len(tag) > 1 && (tag[0] == 'v' || tag[0] == 'V') {
		return tag[1:]
	}
	return tag
}

// semver-ish parsing and comparison.
// Supports:
//   - "1.2.3"
//   - "1.2"
//   - "1"
//   - "1.2.3-rc.1"
// Comparison:
//   - major/minor/patch descending
//   - release > prerelease for same version
//   - prerelease compared by semver rules (numeric identifiers < non-numeric, shorter wins if equal prefix)
type semverKey struct {
	ok         bool
	major      int
	minor      int
	patch      int
	pre        []string
	hasPre     bool
	origString string
}

func parseSemver(s string) semverKey {
	s = strings.TrimSpace(s)
	k := semverKey{origString: s}

	// Require a leading digit to consider it semver-like.
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

	parts := strings.Split(main, ".")
	if len(parts) > 3 {
		return k
	}

	parseInt := func(p string) (int, bool) {
		if p == "" {
			return 0, false
		}
		for _, r := range p {
			if !unicode.IsDigit(r) {
				return 0, false
			}
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return 0, false
		}
		return v, true
	}

	var ok bool
	if len(parts) >= 1 {
		k.major, ok = parseInt(parts[0])
		if !ok {
			return semverKey{origString: s}
		}
	}
	if len(parts) >= 2 {
		k.minor, ok = parseInt(parts[1])
		if !ok {
			return semverKey{origString: s}
		}
	}
	if len(parts) == 3 {
		k.patch, ok = parseInt(parts[2])
		if !ok {
			return semverKey{origString: s}
		}
	}

	if k.hasPre && pre != "" {
		k.pre = strings.Split(pre, ".")
	}
	k.ok = true
	return k
}

func cmpPrerelease(a, b []string) int {
	// Return: -1 if a<b, 0 if equal, +1 if a>b (per semver precedence rules)
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		ai := a[i]
		bi := b[i]

		ain, aIsNum := isNumeric(ai)
		bin, bIsNum := isNumeric(bi)

		switch {
		case aIsNum && bIsNum:
			if ain < bin {
				return -1
			}
			if ain > bin {
				return 1
			}
		case aIsNum && !bIsNum:
			// Numeric identifiers have lower precedence than non-numeric
			return -1
		case !aIsNum && bIsNum:
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
	// If equal so far, shorter prerelease has lower precedence? (Actually: smaller set has lower precedence)
	// Semver: If a has fewer identifiers and all equal so far, a has lower precedence.
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

func isNumeric(s string) (int, bool) {
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

func semverGreater(displayA, displayB string) bool {
	a := parseSemver(displayA)
	b := parseSemver(displayB)

	// Prefer semver-like values over non-semver-like values.
	if a.ok && !b.ok {
		return true
	}
	if !a.ok && b.ok {
		return false
	}
	if !a.ok && !b.ok {
		// Fall back to lexical descending if neither parses.
		return displayA > displayB
	}

	// Compare major/minor/patch descending.
	if a.major != b.major {
		return a.major > b.major
	}
	if a.minor != b.minor {
		return a.minor > b.minor
	}
	if a.patch != b.patch {
		return a.patch > b.patch
	}

	// Same base version: release > prerelease
	if a.hasPre != b.hasPre {
		return !a.hasPre && b.hasPre
	}
	if !a.hasPre && !b.hasPre {
		return false
	}

	// Both prerelease: compare prerelease identifiers (higher wins)
	return cmpPrerelease(a.pre, b.pre) > 0
}

func refreshTagsCmd() tea.Cmd {
	remote := ghrel.GitRemoteURL(hardOwner, hardRepo)

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		tags, err := ghrel.GetTagsViaGit(ctx, remote)
		if err != nil {
			return tagsErrMsg{err: err}
		}
		return tagsLoadedMsg{tags: tags}
	}
}

func downloadCmd(tag, out, token string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := ghrel.DownloadReleaseAssetByTag(
			ctx,
			hardOwner,
			hardRepo,
			tag, // RAW tag (do not normalize)
			hardAsset,
			out,
			token,
		); err != nil {
			return downloadErrMsg{err: err}
		}

		return downloadDoneMsg{out: out}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spin.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tags.SetSize(max(msg.Width-4, 40), max(msg.Height-14, 6))
		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		if key == "q" || key == "ctrl+c" {
			return m, tea.Quit
		}

		if key == "esc" {
			m.clearError()
			m.status = "Ready"
			return m, nil
		}

		if key == "ctrl+r" {
			if m.loadingTags {
				return m, nil
			}
			m.clearError()
			m.loadingTags = true
			m.status = "Refreshing tags…"
			return m, refreshTagsCmd()
		}

		if key == "ctrl+d" {
			if m.downloading {
				return m, nil
			}
			if err := m.validateDownload(); err != nil {
				m.setError(err)
				return m, nil
			}
			m.clearError()
			m.downloading = true
			m.status = "Downloading…"
			return m, downloadCmd(
				m.selectedTag,
				m.resolveOutput(),
				m.resolveToken(),
			)
		}

		if key == "tab" {
			m.focus = focusTarget((int(m.focus) + 1) % focusCount)
			m.applyFocus()
			return m, nil
		}
		if key == "shift+tab" {
			i := int(m.focus) - 1
			if i < 0 {
				i = focusCount - 1
			}
			m.focus = focusTarget(i)
			m.applyFocus()
			return m, nil
		}

		if m.focus == focusTags {
			var cmd tea.Cmd
			m.tags, cmd = m.tags.Update(msg)

			if key == "enter" {
				if it, ok := m.tags.SelectedItem().(tagItem); ok {
					m.selectedTag = it.raw
					// Show the normalized tag in status
					if it.isLatest {
						m.status = "Selected tag: " + it.display + " (latest)"
					} else {
						m.status = "Selected tag: " + it.display
					}
				}
			}
			return m, cmd
		}

		return m.updateFocusedInput(msg)

	case tagsLoadedMsg:
		m.loadingTags = false

		// Build tag items (raw + normalized display).
		items := make([]tagItem, 0, len(msg.tags))
		for _, t := range msg.tags {
			items = append(items, tagItem{
				raw:     t,
				display: normalizeTag(t),
			})
		}

		if len(items) == 0 {
			m.setError(errors.New("no tags found for this repository"))
			m.status = "No tags found."
			m.tags.SetItems(nil)
			return m, nil
		}

		// Sort semver descending by display value.
		sort.Slice(items, func(i, j int) bool {
			di := items[i].display
			dj := items[j].display
			if di == dj {
				// Tie-breaker: stable ordering by raw desc
				return items[i].raw > items[j].raw
			}
			return semverGreater(di, dj)
		})

		// Mark "latest" (first item after sort).
		items[0].isLatest = true

		// Convert to list.Items
		litems := make([]list.Item, 0, len(items))
		for _, it := range items {
			litems = append(litems, it)
		}
		m.tags.SetItems(litems)

		// Preserve selection by RAW tag if possible; otherwise select first.
		selectedIdx := 0
		if m.selectedTag != "" {
			found := false
			for i := range items {
				if items[i].raw == m.selectedTag {
					selectedIdx = i
					found = true
					break
				}
			}
			if !found {
				// fall back to latest
				m.selectedTag = items[0].raw
			}
		} else {
			m.selectedTag = items[0].raw
		}

		m.tags.Select(selectedIdx)

		// Status message
		if items[0].raw == m.selectedTag {
			m.status = "Loaded tags. Latest: " + items[0].display
		} else {
			m.status = "Loaded tags."
		}

		return m, nil

	case tagsErrMsg:
		m.loadingTags = false
		m.setError(msg.err)
		return m, nil

	case downloadDoneMsg:
		m.downloading = false
		m.status = "Downloaded: " + msg.out
		return m, nil

	case downloadErrMsg:
		m.downloading = false
		m.setError(msg.err)
		return m, nil

	default:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)

		// Auto refresh once at startup
		if m.initialRefresh {
			m.initialRefresh = false
			m.loadingTags = true
			m.status = "Refreshing tags…"
			return m, tea.Batch(cmd, refreshTagsCmd())
		}

		return m, cmd
	}
}

func (m *model) updateFocusedInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.focus {
	case focusOutput:
		m.output, cmd = m.output.Update(msg)
	case focusToken:
		m.token, cmd = m.token.Update(msg)
	}
	return *m, cmd
}

func (m *model) applyFocus() {
	m.output.Blur()
	m.token.Blur()

	switch m.focus {
	case focusOutput:
		m.output.Focus()
	case focusToken:
		m.token.Focus()
	}
}

