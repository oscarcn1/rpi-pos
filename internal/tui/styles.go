package tui

import "github.com/charmbracelet/lipgloss"

var (
	subtitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229"))

	menuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("229")).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240"))

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57")).
			Padding(0, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// Help key styles by action type
	hkNav = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))        // blue - navigation
	hkOk  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))        // green - confirm/positive
	hkDel = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))       // red - delete/cancel
	hkAct = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))       // yellow - actions
	hkDim = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))       // gray - separator

	instructionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Italic(true).
				PaddingLeft(2)

	// Table with border for list screens
	tableBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1)

	tableRowSelected = lipgloss.NewStyle().
				Background(lipgloss.Color("27")).
				Foreground(lipgloss.Color("255")).
				Bold(true)

	tableRowNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	tableHeaderRow = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			Underline(true)
)

func hSep() string { return hkDim.Render(" · ") }

func hKey(style lipgloss.Style, key, label string) string {
	return style.Render(key+": "+label)
}
