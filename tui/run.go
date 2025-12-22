package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func Run() error {
	m := newModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
