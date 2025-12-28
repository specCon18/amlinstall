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

// versionKey supports ANY number of numeric dot segments:
//   - "0.2.7.4"
//   - "1.2.3"
//   - "1.2"
//   - "1"
// and semver-like prerelease ordering:
//   - release > prerelease (same core)
//   - prerelease identifiers compared per semver rules (numeric < non-numeric, etc.)
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

func versionGreater(aDisp, bDisp string) bool {
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
			tag, // RAW tag
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
			m.status = "Refreshing version list…"
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
					if it.isLatest {
						m.status = "Selected version: " + it.display + " (latest)"
					} else {
						m.status = "Selected version: " + it.display
					}
				}
			}
			return m, cmd
		}

		return m.updateFocusedInput(msg)

	case tagsLoadedMsg:
		m.loadingTags = false

		items := make([]tagItem, 0, len(msg.tags))
		for _, t := range msg.tags {
			items = append(items, tagItem{
				raw:     t,
				display: normalizeTag(t),
			})
		}

		if len(items) == 0 {
			m.setError(errors.New("no versions found for this repository"))
			m.status = "No tags found."
			m.tags.SetItems(nil)
			return m, nil
		}

		// FIXED: sort versions descending, supporting 4+ numeric segments like 0.2.7.4
		sort.Slice(items, func(i, j int) bool {
			di := items[i].display
			dj := items[j].display
			if di == dj {
				return items[i].raw > items[j].raw
			}
			return versionGreater(di, dj)
		})

		items[0].isLatest = true

		litems := make([]list.Item, 0, len(items))
		for _, it := range items {
			litems = append(litems, it)
		}
		m.tags.SetItems(litems)

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
				m.selectedTag = items[0].raw
				selectedIdx = 0
			}
		} else {
			m.selectedTag = items[0].raw
			selectedIdx = 0
		}

		m.tags.Select(selectedIdx)

		selectedDisplay := items[selectedIdx].display
		if selectedIdx == 0 {
			m.status = "Selected version: " + selectedDisplay + " (latest)"
		} else {
			m.status = "Selected version: " + selectedDisplay
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

		if m.initialRefresh {
			m.initialRefresh = false
			m.loadingTags = true
			m.status = "Refreshing version list…"
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

