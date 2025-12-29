package tui

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"automelonloaderinstallergo/internal/releases"
	"automelonloaderinstallergo/internal/version"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type versionsLoadedMsg struct {
	versions []string
}

type versionsErrMsg struct {
	err error
}

type versionsCanceledMsg struct{}

type downloadDoneMsg struct {
	out string
}

type downloadErrMsg struct {
	err error
}

type downloadCanceledMsg struct{}

// initRefreshMsg triggers the startup auto-refresh flow.
type initRefreshMsg struct{}

const focusCount = int(focusToken) + 1

func retryWithBackoff(ctx context.Context, attempts int, baseDelay time.Duration, fn func() error) error {
	delay := baseDelay
	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := fn()
		if err == nil {
			return nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		if i == attempts-1 {
			return err
		}
		t := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}
		delay *= 2
	}
	return ctx.Err()
}

func refreshVersionsCmd(ctx context.Context, src releases.Source, token string) tea.Cmd {
	return func() tea.Msg {
		var versions []string
		err := retryWithBackoff(ctx, 3, 250*time.Millisecond, func() error {
			v, e := src.ListTags(ctx, hardOwner, hardRepo, token)
			if e == nil {
				versions = v
			}
			return e
		})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return versionsCanceledMsg{}
			}
			return versionsErrMsg{err: fmt.Errorf("refresh versions: %w", err)}
		}
		return versionsLoadedMsg{versions: versions}
	}
}

func downloadCmd(ctx context.Context, src releases.Source, tag, out, token string) tea.Cmd {
	return func() tea.Msg {
		err := retryWithBackoff(ctx, 3, 500*time.Millisecond, func() error {
			return src.DownloadAsset(ctx, hardOwner, hardRepo, tag /* raw tag */, hardAsset, out, token)
		})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return downloadCanceledMsg{}
			}
			return downloadErrMsg{err: fmt.Errorf("download asset: %w", err)}
		}
		return downloadDoneMsg{out: out}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spin.Tick,
		func() tea.Msg { return initRefreshMsg{} },
	)
}

func (m *model) startRefresh() tea.Cmd {
	// Cancel/replace policy: starting a refresh cancels any in-flight work.
	m.cancelDownload()
	m.downloading = false
	m.cancelRefresh()

	if err := m.validateRefresh(); err != nil {
		m.SetError(err)
		return nil
	}

	m.ClearBanner()
	m.loadingVersions = true
	m.SetStatus("Refreshing version list…")

	baseCtx, cancel := context.WithCancel(context.Background())
	m.refreshCancel = cancel
	ctx, timeoutCancel := context.WithTimeout(baseCtx, 30*time.Second)

	inner := refreshVersionsCmd(ctx, m.src, m.resolveToken())
	return func() tea.Msg {
		defer timeoutCancel()
		return inner()
	}
}

func (m *model) startDownload() tea.Cmd {
	// Cancel/replace policy: starting a download cancels any in-flight work.
	m.cancelRefresh()
	m.loadingVersions = false
	m.cancelDownload()

	if err := m.validateDownload(); err != nil {
		m.SetError(err)
		return nil
	}

	m.ClearBanner()
	m.downloading = true
	m.SetStatus("Downloading…")

	baseCtx, cancel := context.WithCancel(context.Background())
	m.downloadCancel = cancel
	ctx, timeoutCancel := context.WithTimeout(baseCtx, 2*time.Minute)

	inner := downloadCmd(
		ctx,
		m.src,
		m.selectedVersionTag,
		m.resolveOutput(),
		m.resolveToken(),
	)
	return func() tea.Msg {
		defer timeoutCancel()
		return inner()
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case initRefreshMsg:
		return m, m.startRefresh()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.versions.SetSize(max(msg.Width-4, 40), max(msg.Height-14, 6))
		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		if key == "q" || key == "ctrl+c" {
			m.cancelRefresh()
			m.cancelDownload()
			return m, tea.Quit
		}

		if key == "esc" {
			m.ClearBanner()
			m.SetStatus("Ready")
			return m, nil
		}

		if key == "ctrl+r" {
			return m, m.startRefresh()
		}


		if key == "ctrl+d" {
			return m, m.startDownload()
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

		if m.focus == focusVersions {
			var cmd tea.Cmd
			m.versions, cmd = m.versions.Update(msg)

			if key == "enter" {
				if it, ok := m.versions.SelectedItem().(versionItem); ok {
					m.selectedVersionTag = it.raw
					if it.isLatest {
						m.SetStatus("Selected version: " + it.display + " (latest)")
					} else {
						m.SetStatus("Selected version: " + it.display)
					}
				}
			}
			return m, cmd
		}

		return m.updateFocusedInput(msg)

	case versionsLoadedMsg:
		m.loadingVersions = false
		m.refreshCancel = nil

		items := make([]versionItem, 0, len(msg.versions))
		for _, t := range msg.versions {
			items = append(items, versionItem{
				raw:     t,
				display: version.NormalizeTag(t),
			})
		}

		if len(items) == 0 {
			m.SetError(errors.New("no versions found for this repository"))
			m.SetStatus("No versions found.")
			m.versions.SetItems(nil)
			return m, nil
		}

		// FIXED: sort versions descending, supporting 4+ numeric segments like 0.2.7.4
		sort.Slice(items, func(i, j int) bool {
			di := items[i].display
			dj := items[j].display
			if di == dj {
				return items[i].raw > items[j].raw
			}
			return version.Greater(di, dj)
		})

		items[0].isLatest = true

		litems := make([]list.Item, 0, len(items))
		for _, it := range items {
			litems = append(litems, it)
		}
		m.versions.SetItems(litems)

		selectedIdx := 0
		if m.selectedVersionTag != "" {
			found := false
			for i := range items {
				if items[i].raw == m.selectedVersionTag {
					selectedIdx = i
					found = true
					break
				}
			}
			if !found {
				m.selectedVersionTag = items[0].raw
				selectedIdx = 0
			}
		} else {
			m.selectedVersionTag = items[0].raw
			selectedIdx = 0
		}

		m.versions.Select(selectedIdx)

		selectedDisplay := items[selectedIdx].display
		if selectedIdx == 0 {
			m.SetStatus("Selected version: " + selectedDisplay + " (latest)")
		} else {
			m.SetStatus("Selected version: " + selectedDisplay)
		}

		return m, nil

	case versionsErrMsg:
		m.loadingVersions = false
		m.refreshCancel = nil
		m.SetError(msg.err)
		return m, nil

	case versionsCanceledMsg:
		m.loadingVersions = false
		m.refreshCancel = nil
		m.SetStatus("Refresh canceled.")
		return m, nil

	case downloadDoneMsg:
		m.downloading = false
		m.downloadCancel = nil
		m.SetStatus("Downloaded: " + msg.out)
		return m, nil

	case downloadErrMsg:
		m.downloading = false
		m.downloadCancel = nil
		m.SetError(msg.err)
		return m, nil

	case downloadCanceledMsg:
		m.downloading = false
		m.downloadCancel = nil
		m.SetStatus("Download canceled.")
		return m, nil

	default:
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)

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
