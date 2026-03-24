package tui

import (
	"fmt"
	"pos/internal/models"
	"pos/internal/store"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type dayReturnsLoadedMsg *models.DayReturnsReport

type dayReturnsModel struct {
	store  *store.Store
	report *models.DayReturnsReport
}

func newDayReturnsModel(s *store.Store) dayReturnsModel {
	return dayReturnsModel{store: s}
}

func (m dayReturnsModel) load() tea.Cmd {
	return func() tea.Msg {
		r, _ := m.store.DayReturnsReport()
		return dayReturnsLoadedMsg(r)
	}
}

func (m dayReturnsModel) update(msg tea.Msg) (dayReturnsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dayReturnsLoadedMsg:
		m.report = msg
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "enter" {
			return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
		}
	}
	return m, nil
}

func (m dayReturnsModel) view() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(" Devoluciones del Día "))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("Resumen de todas las devoluciones realizadas hoy."))
	b.WriteString("\n\n")

	if m.report == nil {
		b.WriteString(dimStyle.Render("  Cargando...\n"))
		return b.String()
	}

	r := m.report
	b.WriteString(fmt.Sprintf("  Fecha: %s\n\n", r.Date))

	summary := fmt.Sprintf(
		"  Devoluciones realizadas:  %s\n"+
			"  Monto total devuelto:    %s",
		fmtI(r.TotalReturns), fmtP(r.TotalAmount))
	b.WriteString(boxStyle.Render(summary))
	b.WriteString("\n\n")

	if len(r.Returns) == 0 {
		b.WriteString(successStyle.Render("  No hay devoluciones hoy"))
		b.WriteString("\n")
	} else {
		var tbl strings.Builder
		header := fmt.Sprintf(" %-5s %-6s %-28s %8s %10s",
			"#", "Venta", "Razón", "Arts.", "Monto")
		tbl.WriteString(tableHeaderRow.Render(header) + "\n")

		for _, rs := range r.Returns {
			reason := rs.Reason
			if reason == "" {
				reason = "(sin razón)"
			}
			line := fmt.Sprintf(" %-5d %-6d %-28s %8s %10s",
				rs.ID, rs.SaleID, truncate(reason, 28), fmtI(rs.ItemCount), fmtP(rs.Total))
			tbl.WriteString(warnStyle.Render(line) + "\n")
		}
		b.WriteString(tableBoxStyle.Render(tbl.String()))
	}

	b.WriteString("\n")
	b.WriteString("  " + hKey(hkNav, "esc/enter", "volver al menú"))
	b.WriteString("\n")
	return b.String()
}
