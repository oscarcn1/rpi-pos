package tui

import (
	"fmt"
	"pos/internal/models"
	"pos/internal/store"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type productsLoadedMsg []models.Product

type productsModel struct {
	store    *store.Store
	products []models.Product
	cursor   int
	message  string
	isError  bool
}

func newProductsModel(s *store.Store) productsModel {
	return productsModel{store: s}
}

func (m productsModel) loadProducts() tea.Cmd {
	return func() tea.Msg {
		products, err := m.store.ListProducts()
		if err != nil {
			return statusMsg("Error: " + err.Error())
		}
		return productsLoadedMsg(products)
	}
}

func (m productsModel) update(msg tea.Msg) (productsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case productsLoadedMsg:
		m.products = msg
		m.cursor = 0
		return m, nil

	case statusMsg:
		m.message = string(msg)
		m.isError = strings.HasPrefix(string(msg), "Error")
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.products)-1 {
				m.cursor++
			}
		case "n":
			return m, func() tea.Msg {
				return switchScreenMsg{screen: screenProductForm}
			}
		case "e":
			if len(m.products) > 0 {
				p := m.products[m.cursor]
				return m, func() tea.Msg {
					return switchScreenMsg{
						screen: screenProductForm,
						data:   &productFormData{editing: &p},
					}
				}
			}
		case "d":
			if len(m.products) > 0 {
				p := m.products[m.cursor]
				if err := m.store.DeleteProduct(p.ID); err != nil {
					m.message = "Error: " + err.Error()
					m.isError = true
				} else {
					m.message = fmt.Sprintf("'%s' eliminado", p.Name)
					m.isError = false
				}
				return m, m.loadProducts()
			}
		}
	}
	return m, nil
}

func (m productsModel) view() string {
	var b strings.Builder

	b.WriteString(subtitleStyle.Render("  Productos"))
	b.WriteString("\n\n")

	if m.message != "" {
		if m.isError {
			b.WriteString("  " + errorStyle.Render(m.message) + "\n\n")
		} else {
			b.WriteString("  " + successStyle.Render(m.message) + "\n\n")
		}
	}

	if len(m.products) == 0 {
		b.WriteString(dimStyle.Render("  No hay productos registrados\n"))
	} else {
		maxVisible := 20
		total := len(m.products)

		start := m.cursor - maxVisible/2
		if start < 0 {
			start = 0
		}
		end := start + maxVisible
		if end > total {
			end = total
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}

		var tbl strings.Builder
		header := fmt.Sprintf(" %-10s %-24s %10s %10s %8s",
			"Código", "Nombre", "P.Venta", "Stock", "Tipo")
		tbl.WriteString(tableHeaderRow.Render(header) + "\n")

		for i := start; i < end; i++ {
			p := m.products[i]
			stockStr := p.StockLabel()
			typeStr := "pza"
			if p.IsMeasured() {
				typeStr = p.MeasurementUnit
			}
			line := fmt.Sprintf(" %-10s %-24s %10s %10s %8s",
				truncate(p.Code, 10), truncate(p.Name, 24),
				fmtP(p.SalePrice), stockStr, typeStr)
			if i == m.cursor {
				tbl.WriteString(tableRowSelected.Render(padRight(line, 66)) + "\n")
			} else {
				tbl.WriteString(tableRowNormal.Render(line) + "\n")
			}
		}
		b.WriteString(tableBoxStyle.Render(tbl.String()))

		if total > maxVisible {
			b.WriteString("\n" + dimStyle.Render(fmt.Sprintf("  Mostrando %s-%s de %s productos",
				fmtI(start+1), fmtI(end), fmtI(total))))
		}
	}

	b.WriteString("\n")
	b.WriteString("  " + hKey(hkOk, "n", "nuevo") + hSep() + hKey(hkAct, "e", "editar") + hSep() + hKey(hkDel, "d", "eliminar") + hSep() + hKey(hkNav, "↑↓", "navegar") + hSep() + hKey(hkNav, "esc", "menú"))
	b.WriteString("\n")
	return b.String()
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		return string(r[:max-1]) + "…"
	}
	return s
}
