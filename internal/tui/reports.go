package tui

import (
	"fmt"
	"pos/internal/models"
	"pos/internal/store"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// --- Day Close Report ---

type dayCloseLoadedMsg *models.DayReport

type dayCloseModel struct {
	store  *store.Store
	report *models.DayReport
}

func newDayCloseModel(s *store.Store) dayCloseModel {
	return dayCloseModel{store: s}
}

func (m dayCloseModel) load() tea.Cmd {
	return func() tea.Msg {
		r, _ := m.store.DayCloseReport()
		return dayCloseLoadedMsg(r)
	}
}

func (m dayCloseModel) update(msg tea.Msg) (dayCloseModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dayCloseLoadedMsg:
		m.report = msg
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "enter" {
			return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
		}
	}
	return m, nil
}

func (m dayCloseModel) view() string {
	var b strings.Builder

	b.WriteString(screenTitleStyle.Render("Cierre del Día"))
	b.WriteString("\n\n")

	if m.report == nil {
		b.WriteString(dimStyle.Render("  Cargando...\n"))
		return b.String()
	}

	r := m.report
	b.WriteString(fmt.Sprintf("  Fecha: %s\n\n", r.Date))

	info := fmt.Sprintf(
		"  Ventas realizadas:  %s\n"+
			"  Ingresos totales:   %14s\n"+
			"  Costo de productos: %14s\n"+
			"  Ganancia:           %14s\n"+
			"  Merma del día:      %s unidades",
		fmtI(r.TotalSales), fmtP(r.TotalIncome), fmtP(r.TotalCost), fmtP(r.Profit), fmtQ(r.TotalShrinkage))
	b.WriteString(boxStyle.Render(info))
	b.WriteString("\n\n")

	if len(r.TopProducts) > 0 {
		b.WriteString(subtitleStyle.Render("  Productos más vendidos:"))
		b.WriteString("\n")
		header := fmt.Sprintf("  %-30s %10s %12s", "Producto", "Cantidad", "Total")
		b.WriteString(tableHeaderStyle.Render(header) + "\n")
		for _, ps := range r.TopProducts {
			b.WriteString(fmt.Sprintf("  %-30s %10s %12s\n",
				truncate(ps.ProductName, 30), fmtQ(ps.Quantity), fmtP(ps.Total)))
		}
	}

	b.WriteString("\n")
	b.WriteString("  " + hKey(hkNav, "esc/enter", "volver al menú"))
	b.WriteString("\n")
	return b.String()
}

// --- Reorder Report ---

type reorderLoadedMsg []models.ReorderItem

type reorderModel struct {
	store  *store.Store
	items  []models.ReorderItem
	scroll int
}

func newReorderModel(s *store.Store) reorderModel {
	return reorderModel{store: s}
}

func (m reorderModel) load() tea.Cmd {
	return func() tea.Msg {
		items, _ := m.store.ReorderReport()
		return reorderLoadedMsg(items)
	}
}

func (m reorderModel) update(msg tea.Msg) (reorderModel, tea.Cmd) {
	switch msg := msg.(type) {
	case reorderLoadedMsg:
		m.items = msg
		m.scroll = 0
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter":
			return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
		case "down", "j":
			if m.items != nil && m.scroll < len(m.items)-1 {
				m.scroll++
			}
		case "up", "k":
			if m.scroll > 0 {
				m.scroll--
			}
		}
	}
	return m, nil
}

func (m reorderModel) view() string {
	var b strings.Builder

	b.WriteString(screenTitleStyle.Render("Reporte de Reorden"))
	b.WriteString("\n\n")

	if m.items == nil {
		b.WriteString(dimStyle.Render("  Cargando...\n"))
		return b.String()
	}

	if len(m.items) == 0 {
		b.WriteString(successStyle.Render("  Todos los productos tienen stock suficiente"))
		b.WriteString("\n")
	} else {
		maxVisible := 20
		total := len(m.items)

		start := m.scroll
		end := start + maxVisible
		if end > total {
			end = total
		}

		var tbl strings.Builder
		header := fmt.Sprintf(" %-10s %-22s %10s %10s %10s",
			"Código", "Nombre", "Stock", "Mínimo", "Faltan")
		tbl.WriteString(tableHeaderRow.Render(header) + "\n")

		for i := start; i < end; i++ {
			item := m.items[i]
			p := item.Product
			line := fmt.Sprintf(" %-10s %-22s %10s %10s %10s",
				truncate(p.Code, 10), truncate(p.Name, 22),
				p.StockLabel(), p.MinStockLabel(), p.FormatQty(item.Deficit))
			if p.Stock == 0 {
				tbl.WriteString(errorStyle.Render(line) + "\n")
			} else {
				tbl.WriteString(warnStyle.Render(line) + "\n")
			}
		}
		b.WriteString(tableBox("214").Render(tbl.String()))

		b.WriteString(fmt.Sprintf("\n  %s",
			warnStyle.Render(fmt.Sprintf("Total: %s productos necesitan reorden", fmtI(total)))))
		if total > maxVisible {
			b.WriteString(dimStyle.Render(fmt.Sprintf("  (mostrando %s-%s)",
				fmtI(start+1), fmtI(end))))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("  " + hKey(hkNav, "↑↓", "desplazar") + hSep() + hKey(hkNav, "esc/enter", "volver al menú"))
	b.WriteString("\n")
	return b.String()
}
