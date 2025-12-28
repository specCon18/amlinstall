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
	// owner/repo/asset inputs removed
	focusTags focusTarget = iota
	focusOutput
	focusToken
)



type tagItem struct {
	value string
}

func (t tagItem) Title() string       { return t.value }
func (t tagItem) Description() string { return "" }
func (t tagItem) FilterValue() string { return t.value }

type model struct {
	output textinput.Model
	token  textinput.Model

	tags list.Model

	selectedTag string

	focus focusTarget

	loadingTags bool
	downloading bool
	spin        spinner.Model

	status string
	err    error

	width  int
	height int
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

	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 40, 8)
	l.Title = "Tags"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	sp := spinner.New()

	m := model{
		output: output,
		token:  token,
		tags:   l,
		focus:  focusTags,
		spin:   sp,
		status: "ctrl+r: refresh tags   ctrl+d: download   tab: next   shift+tab: prev   q: quit",
	}

	m.applyFocus()
	return m
}

func (m *model) resolveToken() string {
	if strings.TrimSpace(m.token.Value()) != "" {
		return strings.TrimSpace(m.token.Value())
	}
	return strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
}

func (m *model) resolveOutput() string {
	out := strings.TrimSpace(m.output.Value())
	if out != "" {
		return out
	}
	// derive from hardcoded asset
	return filepath.Join(".", "downloads", hardAsset)
}

func (m *model) validateRefresh() error {
	// owner/repo are hardcoded, so always valid
	return nil
}

func (m *model) validateDownload() error {
	if strings.TrimSpace(m.selectedTag) == "" {
		return errors.New("select a tag (refresh with 'r' and choose one)")
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
