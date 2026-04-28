package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuItem struct {
	key    string
	label  string
	screen screen
}

type menuCategory struct {
	name  string
	color lipgloss.Color
	items []menuItem
}

var menuCategories = []menuCategory{
	{"Ventas", lipgloss.Color("42"), []menuItem{
		{"1", "Nueva Venta", screenSale},
		{"9", "Devoluciones", screenReturns},
		{"0", "Devoluciones del Día", screenDayReturns},
	}},
	{"Reportes", lipgloss.Color("75"), []menuItem{
		{"5", "Cierre del Día", screenDayClose},
		{"6", "Reporte de Reorden", screenReorder},
		{"7", "Reporte de Inventario", screenInventory},
		{"8", "Finanzas Mensuales", screenMonthlyFinance},
	}},
	{"Sistema", lipgloss.Color("214"), []menuItem{
		{"2", "Productos", screenProducts},
		{"3", "Registrar Merma", screenShrinkage},
		{"4", "Buscar Producto", screenSearch},
		{"w", "Conectarse a WiFi", screenWifi},
	}},
}

// flatItems returns all menu items in order for cursor navigation
func flatItems() []menuItem {
	var items []menuItem
	for _, cat := range menuCategories {
		items = append(items, cat.items...)
	}
	return items
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
	items := flatItems()
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(items)-1 {
				m.cursor++
			}
		case "enter":
			return m, func() tea.Msg {
				return switchScreenMsg{screen: items[m.cursor].screen}
			}
		default:
			key := msg.String()
			for i, item := range items {
				if item.key == key {
					idx := i
					return m, func() tea.Msg {
						return switchScreenMsg{screen: items[idx].screen}
					}
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

	// Build menu items with category headers
	var content strings.Builder
	idx := 0
	for _, cat := range menuCategories {
		// Category header: ── Ventas ──────
		catLabel := fmt.Sprintf("── %s ", cat.name)
		fill := 36 - len(catLabel)
		if fill < 0 {
			fill = 0
		}
		catLine := catLabel + strings.Repeat("─", fill)
		catStyle := lipgloss.NewStyle().Bold(true).Foreground(cat.color)
		content.WriteString(catStyle.Render(catLine))
		content.WriteString("\n")

		for _, item := range cat.items {
			label := fmt.Sprintf("  [%s]  %-28s", item.key, item.label)
			if idx == m.cursor {
				content.WriteString(menuSelectedStyle.Render(padRight(label, 38)))
			} else {
				content.WriteString(menuNormalStyle.Render(label))
			}
			content.WriteString("\n")
			idx++
		}
	}
	b.WriteString(menuBoxStyle.Render(content.String()))

	b.WriteString("\n\n")
	b.WriteString("  " + hKey(hkNav, "↑↓/jk", "navegar") + hSep() + hKey(hkOk, "enter/1-0/w", "seleccionar"))
	b.WriteString("\n")

	return b.String()
}
