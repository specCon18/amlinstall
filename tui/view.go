package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	// Shrink overall UI width by 4 columns.
	w := m.width - 4
	if w <= 0 {
		w = 92
	}

	var (
		appPad = lipgloss.NewStyle().Padding(1, 2)

		muted = lipgloss.NewStyle().Faint(true)
		bold  = lipgloss.NewStyle().Bold(true)

		titleBar = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder())

		panelBase = lipgloss.NewStyle().
				Padding(1, 1).
				Border(lipgloss.RoundedBorder()).
				MarginTop(1)

		panelFocused = panelBase.Border(lipgloss.DoubleBorder()).Bold(true)
		
		settingsTitleStyle = lipgloss.NewStyle().Bold(true).PaddingBottom(1)

		fieldFocused = lipgloss.NewStyle().Bold(true)
		fieldBlurred = lipgloss.NewStyle().Faint(true)

		statusBox = lipgloss.NewStyle().
				Padding(0, 1).
				Border(lipgloss.RoundedBorder())

		errorBox = lipgloss.NewStyle().
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				Bold(true)

		footer = lipgloss.NewStyle().MarginTop(1)
	)

	gap := 2
	leftW := (w - 2*2 - gap) * 2 / 3
	rightW := (w - 2*2 - gap) - leftW
	if leftW < 40 {
		leftW = 40
	}
	if rightW < 34 {
		rightW = 34
	}

	// Right panel inner width must account for:
	// - 2 columns border (left+right)
	// - 2 columns padding (left+right), since panel padding is (1,1)
	rightInnerW := rightW - 4
	rightInnerW = max(rightInnerW, 10)

	title := "MelonLoader Automated Installer Linux Edition"
	sub := fmt.Sprintf("%s/%s  •  %s", hardOwner, hardRepo, hardAsset)
	if m.loadingTags {
		sub = fmt.Sprintf("%s  •  %s Refreshing Version List…", sub, m.spin.View())
	}
	if m.downloading {
		sub = fmt.Sprintf("%s  •  %s Downloading…", sub, m.spin.View())
	}

	header := titleBar.Width(w-2*2).Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			bold.Render(title),
			muted.Render(sub),
		),
	)

	tagsPanelStyle := panelBase
	if m.focus == focusTags {
		tagsPanelStyle = panelFocused
	}
	settingsPanelStyle := panelBase
	if m.focus == focusOutput || m.focus == focusToken {
		settingsPanelStyle = panelFocused
	}

	// Left panel: keep the list's own title ("Version") and remove our redundant header.
	// Optional: show selected tag as a subtle line below the list (non-redundant).
	var leftBody strings.Builder
	leftBody.WriteString(m.tags.View())

	if strings.TrimSpace(m.selectedTag) != "" {
		// Show raw tag (may include leading v). If you prefer normalized, say so.
		fmt.Fprintf(&leftBody, "\n%s", muted.Render("Selected: "+m.selectedTag))
	}

	tagsPanel := tagsPanelStyle.
		Width(leftW).
		Render(leftBody.String())

	// Right panel: inputs + status
	var rightBody strings.Builder

	settingsTitle := "Download Settings"
	if m.focus == focusOutput || m.focus == focusToken {
		settingsTitle = "▶ " + settingsTitle
	}

	fmt.Fprintf(&rightBody, "%s\n%s\n",
		settingsTitleStyle.Render(settingsTitle),
		muted.Render("Tab/Shift+Tab to change focus."),
	)

	outputView := m.output.View()
	tokenView := m.token.View()

	switch m.focus {
	case focusOutput:
		outputView = fieldFocused.Render(outputView)
		tokenView = fieldBlurred.Render(tokenView)

	case focusToken:
		outputView = fieldBlurred.Render(outputView)
		tokenView = fieldFocused.Render(tokenView)

	default:
		outputView = fieldBlurred.Render(outputView)
		tokenView = fieldBlurred.Render(tokenView)
	}

	fmt.Fprintf(&rightBody, "\n%s\n", outputView)
	fmt.Fprintf(&rightBody, "%s\n", tokenView)

	if strings.TrimSpace(m.status) != "" {
		fmt.Fprintf(&rightBody, "\n%s\n", statusBox.Width(rightInnerW).Render(m.status))
	}
	if m.err != nil {
		fmt.Fprintf(&rightBody, "\n%s\n", errorBox.Width(rightInnerW).Render("Error: "+m.err.Error()))
	}

	rightPanel := settingsPanelStyle.
		Width(rightW).
		Render(rightBody.String())

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		tagsPanel,
		lipgloss.NewStyle().Width(gap).Render(""),
		rightPanel,
	)

	footerLine := footer.Render(muted.Render(helpText))

	return appPad.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			content,
			footerLine,
		),
	)
}

