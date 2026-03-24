package tui

import (
	"fmt"
	"pos/internal/models"
	"pos/internal/store"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type searchModel struct {
	store    *store.Store
	input    textinput.Model
	results  []models.Product
	cursor   int
	searched bool
}

func newSearchModel(s *store.Store) searchModel {
	ti := textinput.New()
	ti.Placeholder = "Buscar por código, nombre o categoría"
	ti.Focus()
	ti.CharLimit = 50
	return searchModel{store: s, input: ti}
}

func (m searchModel) update(msg tea.Msg) (searchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
		case "enter":
			q := strings.TrimSpace(m.input.Value())
			if q == "" {
				return m, nil
			}
			results, _ := m.store.FuzzySearchProducts(q)
			m.results = results
			m.searched = true
			m.cursor = 0
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m searchModel) view() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(" Buscar Producto "))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("Escribe parte del nombre, código o categoría. No necesitas acentos."))
	b.WriteString("\n\n")

	b.WriteString("  " + m.input.View() + "\n\n")

	if m.searched {
		if len(m.results) == 0 {
			b.WriteString(dimStyle.Render("  Sin resultados\n"))
		} else {
			b.WriteString(fmt.Sprintf("  %s\n", dimStyle.Render(fmt.Sprintf("%d resultado(s)", len(m.results)))))
			header := fmt.Sprintf("  %-10s %-20s %-10s %9s %10s %10s",
				"Código", "Nombre", "Categoría", "P.Venta", "Stock", "Tipo")
			b.WriteString(tableHeaderStyle.Render(header) + "\n")

			for i, p := range m.results {
				line := fmt.Sprintf("  %-10s %-20s %-10s %10s %10s %10s",
					truncate(p.Code, 10), truncate(p.Name, 20), truncate(p.Category, 10),
					fmtP(p.SalePrice), p.StockLabel(), p.UnitLabel())
				if i == m.cursor {
					b.WriteString(selectedStyle.Render(line))
				} else {
					if p.Stock <= p.MinStock {
						b.WriteString(warnStyle.Render(line))
					} else {
						b.WriteString(line)
					}
				}
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString("  " + hKey(hkOk, "enter", "buscar") + hSep() + hKey(hkNav, "↑↓", "navegar") + hSep() + hKey(hkNav, "esc", "menú"))
	b.WriteString("\n")
	return b.String()
}
