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

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(" Cierre del Día "))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("Resumen de ventas, costos y ganancias del día. Revísalo antes de cerrar."))
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
	store *store.Store
	items []models.ReorderItem
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
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "enter" {
			return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
		}
	}
	return m, nil
}

func (m reorderModel) view() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(" Reporte de Reorden "))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("Productos con stock igual o menor al mínimo. Usa esta lista para hacer pedidos."))
	b.WriteString("\n\n")

	if m.items == nil {
		b.WriteString(dimStyle.Render("  Cargando...\n"))
		return b.String()
	}

	if len(m.items) == 0 {
		b.WriteString(successStyle.Render("  Todos los productos tienen stock suficiente"))
		b.WriteString("\n")
	} else {
		var tbl strings.Builder
		header := fmt.Sprintf(" %-10s %-22s %10s %10s %10s",
			"Código", "Nombre", "Stock", "Mínimo", "Faltan")
		tbl.WriteString(tableHeaderRow.Render(header) + "\n")

		for _, item := range m.items {
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
		b.WriteString(tableBoxStyle.Render(tbl.String()))

		b.WriteString(fmt.Sprintf("\n  %s\n",
			warnStyle.Render(fmt.Sprintf("Total: %s productos necesitan reorden", fmtI(len(m.items))))))
	}

	b.WriteString("\n")
	b.WriteString("  " + hKey(hkNav, "esc/enter", "volver al menú"))
	b.WriteString("\n")
	return b.String()
}
