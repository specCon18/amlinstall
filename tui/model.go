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
	focusOwner focusTarget = iota
	focusRepo
	focusTags
	focusAsset
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
	owner  textinput.Model
	repo   textinput.Model
	asset  textinput.Model
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
	owner := textinput.New()
	owner.Placeholder = "LavaGang"
	owner.Prompt = "Owner:  "
	owner.CharLimit = 200
	owner.Width = 40

	repo := textinput.New()
	repo.Placeholder = "MelonLoader"
	repo.Prompt = "Repo:   "
	repo.CharLimit = 200
	repo.Width = 40

	asset := textinput.New()
	asset.Placeholder = "MelonLoader.x64.zip"
	asset.Prompt = "Asset:  "
	asset.CharLimit = 500
	asset.Width = 40

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
		owner:  owner,
		repo:   repo,
		asset:  asset,
		output: output,
		token:  token,
		tags:   l,
		focus:  focusOwner,
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
	asset := strings.TrimSpace(m.asset.Value())
	if asset == "" {
		return ""
	}
	return filepath.Join(".", "downloads", asset)
}

func (m *model) validateRefresh() error {
	if strings.TrimSpace(m.owner.Value()) == "" || strings.TrimSpace(m.repo.Value()) == "" {
		return errors.New("owner and repo are required to refresh tags")
	}
	return nil
}

func (m *model) validateDownload() error {
	if strings.TrimSpace(m.owner.Value()) == "" {
		return errors.New("owner is required")
	}
	if strings.TrimSpace(m.repo.Value()) == "" {
		return errors.New("repo is required")
	}
	if strings.TrimSpace(m.selectedTag) == "" {
		return errors.New("select a tag (refresh with 'r' and choose one)")
	}
	if strings.TrimSpace(m.asset.Value()) == "" {
		return errors.New("asset is required")
	}
	if strings.TrimSpace(m.resolveOutput()) == "" {
		return errors.New("output is required (or provide asset so default output can be derived)")
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

