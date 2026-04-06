package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var menuItems = []struct {
	key    string
	label  string
	screen screen
}{
	{"1", "Nueva Venta", screenSale},
	{"2", "Productos", screenProducts},
	{"3", "Registrar Merma", screenShrinkage},
	{"4", "Buscar Producto", screenSearch},
	{"5", "Cierre del Día", screenDayClose},
	{"6", "Reporte de Reorden", screenReorder},
	{"7", "Reporte de Inventario", screenInventory},
	{"8", "Finanzas Mensuales", screenMonthlyFinance},
	{"9", "Devoluciones", screenReturns},
	{"0", "Devoluciones del Día", screenDayReturns},
}

var (
	logoBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57")).
			Padding(0, 2).
			Foreground(lipgloss.Color("75"))

	menuBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1).
			Width(40)

	menuSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("27")).
				Foreground(lipgloss.Color("255")).
				Bold(true)

	menuNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
)

const asciiLogo = `  ____   ___  ____
 |  _ \ / _ \/ ___|
 | |_) | | | \___ \
 |  __/| |_| |___) |
 |_|    \___/|____/
  Punto  de  Venta`

type menuModel struct {
	cursor int
}

func newMenuModel() menuModel {
	return menuModel{}
}

func (m menuModel) update(msg tea.Msg) (menuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(menuItems)-1 {
				m.cursor++
			}
		case "enter":
			return m, func() tea.Msg {
				return switchScreenMsg{screen: menuItems[m.cursor].screen}
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9", "0":
			idx := int(msg.String()[0] - '1')
			if msg.String() == "0" {
				idx = 9
			}
			if idx >= 0 && idx < len(menuItems) {
				return m, func() tea.Msg {
					return switchScreenMsg{screen: menuItems[idx].screen}
				}
			}
		}
	}
	return m, nil
}

func (m menuModel) view() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(logoBorder.Render(asciiLogo))
	b.WriteString("\n\n")
	b.WriteString(instructionStyle.Render("Selecciona una opción con los números o navega con las flechas."))
	b.WriteString("\n\n")

	// Build menu items inside a box
	var items strings.Builder
	for i, item := range menuItems {
		label := fmt.Sprintf("  [%s]  %-28s", item.key, item.label)
		if i == m.cursor {
			items.WriteString(menuSelectedStyle.Render(padRight(label, 38)))
		} else {
			items.WriteString(menuNormalStyle.Render(label))
		}
		items.WriteString("\n")
	}
	b.WriteString(menuBoxStyle.Render(items.String()))

	b.WriteString("\n\n")
	b.WriteString("  " + hKey(hkNav, "↑↓/jk", "navegar") + hSep() + hKey(hkOk, "enter/1-0", "seleccionar"))
	b.WriteString("\n")

	return b.String()
}
