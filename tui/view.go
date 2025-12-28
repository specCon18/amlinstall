package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	// Layout sizing
	w := m.width
	if w <= 0 {
		w = 96 // reasonable default if WindowSizeMsg hasn't arrived yet
	}
	h := m.height
	_ = h // reserved if you later want dynamic vertical sizing

	// Styles
	var (
		// Base
		appPad  = lipgloss.NewStyle().Padding(1, 2)
		muted   = lipgloss.NewStyle().Faint(true)
		bold    = lipgloss.NewStyle().Bold(true)
		warn    = lipgloss.NewStyle().Bold(true)
		success = lipgloss.NewStyle().Bold(true)

		// Header
		titleBar = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderBottom(true)

		// Panels
		panel = lipgloss.NewStyle().
			Padding(1, 1).
			Border(lipgloss.RoundedBorder()).
			MarginTop(1)

		panelTitle = lipgloss.NewStyle().Bold(true)

		// Status boxes
		statusBox = lipgloss.NewStyle().
				Padding(0, 1).
				Border(lipgloss.RoundedBorder())

		errorBox = lipgloss.NewStyle().
				Padding(0, 1).
				Border(lipgloss.RoundedBorder()).
				Bold(true)

		// Footer
		footer = lipgloss.NewStyle().MarginTop(1)
	)

	// Compute column widths
	// Keep tags comfortably wide; controls narrow but usable.
	gap := 2
	leftW := (w - 2*2 - gap) * 2 / 3  // account roughly for outer padding
	rightW := (w - 2*2 - gap) - leftW // remaining
	if leftW < 40 {
		leftW = 40
	}
	if rightW < 34 {
		rightW = 34
	}

	// Header content
	title := "GitHub Release Asset Helper"
	sub := fmt.Sprintf("%s/%s  •  %s", hardOwner, hardRepo, hardAsset)
	if m.loadingTags {
		sub = fmt.Sprintf("%s  •  %s Refreshing tags…", sub, m.spin.View())
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

	// Left panel: tags
	tagHeader := "Tags"
	if m.selectedTag != "" {
		tagHeader = fmt.Sprintf("%s (selected: %s)", tagHeader, m.selectedTag)
	}

	tagsPanel := panel.
		Width(leftW).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				panelTitle.Render(tagHeader),
				m.tags.View(),
			),
		)

	// Right panel: inputs + status
	var rightBody strings.Builder
	fmt.Fprintf(&rightBody, "%s\n%s\n",
		panelTitle.Render("Download Settings"),
		muted.Render("Output path may be empty to use the default."),
	)
	fmt.Fprintf(&rightBody, "\n%s\n", m.output.View())
	fmt.Fprintf(&rightBody, "%s\n", m.token.View())

	// Status area
	if strings.TrimSpace(m.status) != "" {
		// Give a slight semantic cue based on common phrases
		st := m.status
		box := statusBox
		if strings.HasPrefix(strings.ToLower(st), "downloaded:") {
			box = box.Copy()
			st = success.Render(st)
		}
		fmt.Fprintf(&rightBody, "\n%s\n", box.Width(rightW-2).Render(st))
	}

	if m.err != nil {
		fmt.Fprintf(&rightBody, "\n%s\n", errorBox.Width(rightW-2).Render(warn.Render("Error: ")+m.err.Error()))
	}

	rightPanel := panel.
		Width(rightW).
		Render(rightBody.String())

	// Main content row
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		tagsPanel,
		lipgloss.NewStyle().Width(gap).Render(""),
		rightPanel,
	)

	// Footer help
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

