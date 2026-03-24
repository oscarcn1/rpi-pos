package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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
}

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
		case "q", "Q":
			return m, tea.Quit
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
		case "1", "2", "3", "4", "5", "6", "7":
			idx := int(msg.String()[0] - '1')
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
	b.WriteString(titleStyle.Render(" POS - Punto de Venta "))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  Tienda de Enseres Domésticos"))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("Selecciona una opción con los números o navega con las flechas."))
	b.WriteString("\n\n")

	for i, item := range menuItems {
		label := fmt.Sprintf("[%s] %s", item.key, item.label)
		if i == m.cursor {
			b.WriteString(selectedStyle.Render(label))
		} else {
			b.WriteString(menuItemStyle.Render(label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("  " + hKey(hkNav, "↑↓/jk", "navegar") + hSep() + hKey(hkOk, "enter/1-7", "seleccionar") + hSep() + hKey(hkDel, "q", "salir"))
	b.WriteString("\n")

	return b.String()
}
