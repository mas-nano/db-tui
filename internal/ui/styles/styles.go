package styles

import (
	lipgloss "charm.land/lipgloss/v2"
)

var (
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		Padding(0, 1)

	StatusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB")).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1)

	Help = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280"))

	Selected = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	Modal = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Width(52)

	ModalTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED"))

	Dim = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF"))

	FormError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F87171"))

	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#374151")).
		Padding(1, 2)

	Pane = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#374151"))

	PaneFocused = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED"))

	PaneTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#9CA3AF"))

	ToastError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#B91C1C")).
			Padding(0, 1)

	ToastInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#1D4ED8")).
			Padding(0, 1)
)
