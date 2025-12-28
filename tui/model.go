package tui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
)

type focusTarget int

const (
	focusTags focusTarget = iota
	focusOutput
	focusToken
)

type tagItem struct {
	raw      string // exact git tag, e.g. "v0.6.5"
	display  string // UI value, e.g. "0.6.5"
	isLatest bool   // visually mark the first (highest) semver
}

func (t tagItem) Title() string {
	if t.isLatest {
		return t.display + "  (latest)"
	}
	return t.display
}
func (t tagItem) Description() string { return "" }
func (t tagItem) FilterValue() string { return t.display }

type model struct {
	output textinput.Model
	token  textinput.Model

	tags list.Model

	selectedTag string // RAW tag value

	focus focusTarget

	loadingTags bool
	downloading bool
	spin        spinner.Model

	status string
	err    error

	width  int
	height int

	initialRefresh bool
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
	l.Title = "Tags"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	sp := spinner.New()

	m := model{
		output:         output,
		token:          token,
		tags:           l,
		focus:          focusTags,
		spin:           sp,
		status:         "Ready",
		initialRefresh: true,
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
	if strings.TrimSpace(m.selectedTag) == "" {
		return errors.New("select a tag (refresh with 'ctrl+r' and choose one)")
	}
	if strings.TrimSpace(m.resolveOutput()) == "" {
		return errors.New("output is required")
	}
	return nil
}

func (m *model) setError(err error) {
	m.err = err
	if err != nil {
		m.status = err.Error()
	}
}

func (m *model) clearError() {
	m.err = nil
}

