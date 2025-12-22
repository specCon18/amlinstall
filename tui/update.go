package tui

import (
	"context"
	"errors"
	"strings"
	"time"

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

func refreshTagsCmd(owner, repo string) tea.Cmd {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	remote := ghrel.GitRemoteURL(owner, repo)

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

func downloadCmd(owner, repo, tag, asset, out, token string) tea.Cmd {
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)
	tag = strings.TrimSpace(tag)
	asset = strings.TrimSpace(asset)
	out = strings.TrimSpace(out)

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := ghrel.DownloadReleaseAssetByTag(ctx, owner, repo, tag, asset, out, token); err != nil {
			return downloadErrMsg{err: err}
		}
		return downloadDoneMsg{out: out}
	}
}

func (m model) Init() tea.Cmd {
	// Start focused.
	return tea.Batch(m.spin.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Resize tag list to fit better; keep sane minimums.
		w := max(msg.Width-4, 40)
		h := max(msg.Height-14, 6)
		m.tags.SetSize(w, h)
		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		// Global quit.
		if key == "q" || key == "ctrl+c" {
			return m, tea.Quit
		}

		// Clear current error/status with escape (does not quit).
		if key == "esc" {
			m.clearError()
			m.status = "r: refresh tags   d: download   tab: next   shift+tab: prev   q: quit"
			return m, nil
		}

// Refresh tags (Ctrl+R).
if key == "ctrl+r" {
	if m.loadingTags {
		return m, nil
	}
	if err := m.validateRefresh(); err != nil {
		m.setError(err)
		return m, nil
	}
	m.clearError()
	m.loadingTags = true
	m.status = "Refreshing tags…"
	return m, refreshTagsCmd(m.owner.Value(), m.repo.Value())
}

// Download (Ctrl+D).
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
		m.owner.Value(),
		m.repo.Value(),
		m.selectedTag,
		m.asset.Value(),
		m.resolveOutput(),
		m.resolveToken(),
	)
}


		// Focus navigation.
		if key == "tab" {
			m.focus = (m.focus + 1) % (focusToken + 1)
			m.applyFocus()
			return m, nil
		}
		if key == "shift+tab" {
			if m.focus == 0 {
				m.focus = focusToken
			} else {
				m.focus--
			}
			m.applyFocus()
			return m, nil
		}


		// If tags list focused, let it handle keys (including enter).
if m.focus == focusTags {
    var cmd tea.Cmd
    m.tags, cmd = m.tags.Update(msg)

    // Manual selection on Enter.
    if key == "enter" {
        if it, ok := m.tags.SelectedItem().(tagItem); ok {
            m.selectedTag = it.value
            m.status = "Selected tag: " + m.selectedTag
        }
    }
    return m, cmd
}

		// Otherwise, route to the focused input.
		return m.updateFocusedInput(msg)

	case tagsLoadedMsg:
		m.loadingTags = false

		items := make([]list.Item, 0, len(msg.tags))
		for _, t := range msg.tags {
			items = append(items, tagItem{value: t})
		}
		m.tags.SetItems(items)

		if len(msg.tags) == 0 {
			m.selectedTag = ""
			m.status = "No tags found."
			m.setError(errors.New("no tags found for this repository"))
			return m, nil
		}

		// Preserve selection if possible; otherwise select first.
		if m.selectedTag != "" {
			found := false
			for i, t := range msg.tags {
				if t == m.selectedTag {
					m.tags.Select(i)
					found = true
					break
				}
			}
			if !found {
				m.tags.Select(0)
				m.selectedTag = msg.tags[0]
			}
		} else {
			m.tags.Select(0)
			m.selectedTag = msg.tags[0]
		}

		m.status = "Loaded tags. Selected: " + m.selectedTag
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
		// Spinner tick, etc.
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	}
}

func (m *model) updateFocusedInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.focus {
	case focusOwner:
		m.owner, cmd = m.owner.Update(msg)
	case focusRepo:
		m.repo, cmd = m.repo.Update(msg)
	case focusAsset:
		m.asset, cmd = m.asset.Update(msg)
	case focusOutput:
		m.output, cmd = m.output.Update(msg)
	case focusToken:
		m.token, cmd = m.token.Update(msg)
	default:
		return *m, nil
	}
	return *m, cmd
}

func (m *model) applyFocus() {
	// Inputs
	m.owner.Blur()
	m.repo.Blur()
	m.asset.Blur()
	m.output.Blur()
	m.token.Blur()

	switch m.focus {
	case focusOwner:
		m.owner.Focus()
	case focusRepo:
		m.repo.Focus()
	case focusTags:
		// list focus is implicit; nothing to do
	case focusAsset:
		m.asset.Focus()
	case focusOutput:
		m.output.Focus()
	case focusToken:
		m.token.Focus()
	}
}

