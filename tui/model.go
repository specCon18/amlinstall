package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"automelonloaderinstallergo/internal/releases"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

type focusTarget int

const (
	focusVersions focusTarget = iota
	focusOutput
	focusToken
)

type versionItem struct {
	raw      string // exact git tag, e.g. "v0.6.5"
	display  string // UI value, e.g. "0.6.5"
	isLatest bool   // visually mark the first (highest) semver
}

func (t versionItem) Title() string {
	if t.isLatest {
		return t.display + "  (latest)"
	}
	return t.display
}
func (t versionItem) Description() string { return "" }
func (t versionItem) FilterValue() string { return t.display }

type banner struct {
	status string
	err    error
}

type model struct {
	output textinput.Model
	token  textinput.Model

	versions list.Model

	selectedVersionTag string // RAW tag value

	focus focusTarget

	loadingVersions bool
	downloading     bool
	spin            spinner.Model

	banner banner

	width  int
	height int

	src releases.Source

	refreshCancel  context.CancelFunc
	downloadCancel context.CancelFunc
}

func (m *model) cancelRefresh() {
	if m.refreshCancel != nil {
		m.refreshCancel()
		m.refreshCancel = nil
	}
}

func (m *model) cancelDownload() {
	if m.downloadCancel != nil {
		m.downloadCancel()
		m.downloadCancel = nil
	}
}

func newModel() model {
	output := textinput.New()
	output.Placeholder = "./downloads/<asset>"
	output.Prompt = "Output: "
	output.CharLimit = 2000
	output.Width = 40

	token := textinput.New()
	token.Placeholder = "(optional; overrides GITHUB_TOKEN)"
	token.Prompt = "Token:  "
	token.CharLimit = 4000
	token.Width = 40
	token.EchoMode = textinput.EchoPassword
	token.EchoCharacter = '•'

	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 40, 8)
	l.Title = "Version"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	sp := spinner.New()

	m := model{
		output:         output,
		token:          token,
		versions:       l,
		focus:          focusVersions,
		spin:           sp,
		banner: banner{status: "Ready"},
		src:    releases.NewGitHubSource(),
	}

	m.applyFocus()
	return m
}

func (m *model) resolveToken() string {
	if v := strings.TrimSpace(m.token.Value()); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
}

func (m *model) resolveOutput() string {
	if out := strings.TrimSpace(m.output.Value()); out != "" {
		return out
	}
	return filepath.Join(".", "downloads", hardAsset)
}

func (m *model) validateRefresh() error {
	return nil
}

func (m *model) validateDownload() error {
	if strings.TrimSpace(m.selectedVersionTag) == "" {
		return errors.New("select a version (refresh with 'ctrl+r' and choose one)")
	}
	if strings.TrimSpace(m.resolveOutput()) == "" {
		return errors.New("output is required")
	}
	return nil
}

func (m *model) SetStatus(msg string) {
	m.banner.status = msg
	// A new status generally indicates a new UI state; clear any prior error.
	m.banner.err = nil
}

func (m *model) SetError(err error) {
	m.banner.err = err
	if err != nil {
		m.banner.status = err.Error()
	}
}

func (m *model) ClearBanner() {
	m.banner.status = ""
	m.banner.err = nil
}

func (m model) Status() string { return m.banner.status }
func (m model) Err() error     { return m.banner.err }
