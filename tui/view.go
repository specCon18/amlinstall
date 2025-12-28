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
	kvStyle := lipgloss.NewStyle().Faint(true)

	fmt.Fprintf(&b, "%s\n\n", headerStyle.Render("MelonLoader Automated Installer Linux"))

	// Static config (replaces owner/repo/asset inputs)
	fmt.Fprintf(&b, "%s\n", kvStyle.Render(fmt.Sprintf("Owner: %s", hardOwner)))
	fmt.Fprintf(&b, "%s\n", kvStyle.Render(fmt.Sprintf("Repo:  %s", hardRepo)))
	fmt.Fprintf(&b, "%s\n", kvStyle.Render(fmt.Sprintf("Asset: %s", hardAsset)))

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
	fmt.Fprintf(&b, "\n%s\n", m.output.View())
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

	// Footer help (single source of truth)
	fmt.Fprintf(&b, "\n%s\n", statusStyle.Render(helpText))

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

