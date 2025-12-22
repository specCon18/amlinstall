package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true)
	sectionStyle := lipgloss.NewStyle().MarginTop(1)
	statusStyle := lipgloss.NewStyle().Faint(true)
	errorStyle := lipgloss.NewStyle().Bold(true)

	fmt.Fprintf(&b, "%s\n\n", headerStyle.Render("GitHub Release Asset Helper (TUI)"))

	// Inputs
	fmt.Fprintf(&b, "%s\n", m.owner.View())
	fmt.Fprintf(&b, "%s\n", m.repo.View())

	// Tag section
	tagHeader := "Tags"
	if m.loadingTags {
		tagHeader = fmt.Sprintf("Tags  %s Refreshing…", m.spin.View())
	}
	if m.selectedTag != "" {
		tagHeader = fmt.Sprintf("%s (selected: %s)", tagHeader, m.selectedTag)
	}

	fmt.Fprintf(&b, "%s\n", sectionStyle.Render(headerStyle.Render(tagHeader)))
	fmt.Fprintf(&b, "%s\n", m.tags.View())

	// Remaining inputs
	fmt.Fprintf(&b, "\n%s\n", m.asset.View())
	fmt.Fprintf(&b, "%s\n", m.output.View())
	fmt.Fprintf(&b, "%s\n", m.token.View())

	// Download indicator
	if m.downloading {
		fmt.Fprintf(&b, "\n%s\n", sectionStyle.Render(fmt.Sprintf("%s Downloading…", m.spin.View())))
	}

	// Status / errors
	fmt.Fprintf(&b, "\n%s\n", statusStyle.Render(m.status))
	if m.err != nil {
		fmt.Fprintf(&b, "%s\n", errorStyle.Render("Error: "+m.err.Error()))
	}

	// Footer help
	help := "r: refresh tags   d: download   tab: next field   shift+tab: prev field   enter: select tag   esc: clear status   q: quit"
	fmt.Fprintf(&b, "\n%s\n", statusStyle.Render(help))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

