package tui

import (
	"fmt"
	"pos/internal/models"
	"pos/internal/store"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type inventoryLoadedMsg *models.InventoryReport

type inventoryModel struct {
	store  *store.Store
	report *models.InventoryReport
	scroll int
}

func newInventoryModel(s *store.Store) inventoryModel {
	return inventoryModel{store: s}
}

func (m inventoryModel) load() tea.Cmd {
	return func() tea.Msg {
		r, _ := m.store.InventoryReport()
		return inventoryLoadedMsg(r)
	}
}

func (m inventoryModel) update(msg tea.Msg) (inventoryModel, tea.Cmd) {
	switch msg := msg.(type) {
	case inventoryLoadedMsg:
		m.report = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter":
			return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
		case "down", "j":
			if m.report != nil && m.scroll < len(m.report.Items)-1 {
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

func (m inventoryModel) view() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(" Reporte de Inventario "))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("Vista completa del inventario con valores. Usa flechas para desplazarte."))
	b.WriteString("\n\n")

	if m.report == nil {
		b.WriteString(dimStyle.Render("  Cargando...\n"))
		return b.String()
	}

	r := m.report

	summary := fmt.Sprintf(
		"  Productos registrados: %s\n"+
			"  Valor total a costo:   %14s\n"+
			"  Valor total a venta:   %14s\n"+
			"  Ganancia potencial:    %14s",
		fmtI(r.TotalProducts), fmtP(r.TotalCostValue), fmtP(r.TotalSaleValue), fmtP(r.PotentialProfit))
	b.WriteString(boxStyle.Render(summary))
	b.WriteString("\n\n")

	if len(r.Categories) > 0 {
		b.WriteString(subtitleStyle.Render("  Por categoría:"))
		b.WriteString("\n")
		catHeader := fmt.Sprintf("  %-20s %6s %14s %14s",
			"Categoría", "Prods", "Costo", "Venta")
		b.WriteString(tableHeaderStyle.Render(catHeader) + "\n")
		for _, c := range r.Categories {
			b.WriteString(fmt.Sprintf("  %-20s %6s %14s %14s\n",
				truncate(c.Category, 20), fmtI(c.Products), fmtP(c.CostValue), fmtP(c.SaleValue)))
		}
		b.WriteString("\n")
	}

	b.WriteString(subtitleStyle.Render("  Detalle de productos:"))
	b.WriteString("\n")
	header := fmt.Sprintf("  %-10s %-20s %10s %12s %12s",
		"Código", "Nombre", "Stock", "Val.Costo", "Val.Venta")
	b.WriteString(tableHeaderStyle.Render(header) + "\n")

	maxVisible := 15
	start := m.scroll
	end := start + maxVisible
	if end > len(r.Items) {
		end = len(r.Items)
	}

	for i := start; i < end; i++ {
		item := r.Items[i]
		p := item.Product
		line := fmt.Sprintf("  %-10s %-20s %10s %12s %12s",
			truncate(p.Code, 10), truncate(p.Name, 20),
			p.StockLabel(), fmtP(item.CostValue), fmtP(item.SaleValue))
		if p.Stock == 0 {
			b.WriteString(errorStyle.Render(line))
		} else if p.Stock <= p.MinStock {
			b.WriteString(warnStyle.Render(line))
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	if len(r.Items) > maxVisible {
		b.WriteString(dimStyle.Render(fmt.Sprintf("\n  Mostrando %s-%s de %s productos",
			fmtI(start+1), fmtI(end), fmtI(len(r.Items)))))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("  " + hKey(hkNav, "↑↓", "desplazar") + hSep() + hKey(hkNav, "esc/enter", "volver al menú"))
	b.WriteString("\n")
	return b.String()
}
