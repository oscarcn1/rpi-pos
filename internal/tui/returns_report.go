package tui

import (
	"fmt"
	"pos/internal/models"
	"pos/internal/store"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type dayReturnsLoadedMsg *models.DayReturnsReport

type dayReturnsModel struct {
	store  *store.Store
	report *models.DayReturnsReport
	date   time.Time
	scroll int
}

func newDayReturnsModel(s *store.Store) dayReturnsModel {
	return dayReturnsModel{store: s, date: time.Now()}
}

func (m dayReturnsModel) load() tea.Cmd {
	date := m.date.Format("2006-01-02")
	return func() tea.Msg {
		r, _ := m.store.DayReturnsReport(date)
		return dayReturnsLoadedMsg(r)
	}
}

func (m dayReturnsModel) update(msg tea.Msg) (dayReturnsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case dayReturnsLoadedMsg:
		m.report = msg
		m.scroll = 0
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter":
			return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
		case "left":
			m.date = m.date.AddDate(0, 0, -1)
			m.report = nil
			return m, m.load()
		case "right":
			next := m.date.AddDate(0, 0, 1)
			if !next.After(time.Now()) {
				m.date = next
				m.report = nil
				return m, m.load()
			}
		case "down", "j":
			if m.report != nil && m.scroll < len(m.report.Returns)-1 {
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

func (m dayReturnsModel) view() string {
	var b strings.Builder

	b.WriteString(subtitleStyle.Render("  Devoluciones del Día"))
	b.WriteString("\n\n")

	if m.report == nil {
		b.WriteString(dimStyle.Render("  Cargando...\n"))
		return b.String()
	}

	r := m.report
	isToday := m.date.Format("2006-01-02") == time.Now().Format("2006-01-02")
	dateLabel := r.Date
	if isToday {
		dateLabel += " (hoy)"
	}
	b.WriteString(fmt.Sprintf("  Fecha: %s\n\n", dateLabel))

	summary := fmt.Sprintf(
		"  Devoluciones realizadas:  %s\n"+
			"  Monto total devuelto:    %s",
		fmtI(r.TotalReturns), fmtP(r.TotalAmount))
	b.WriteString(boxStyle.Render(summary))
	b.WriteString("\n\n")

	if len(r.Returns) == 0 {
		b.WriteString(successStyle.Render("  No hay devoluciones en este día"))
		b.WriteString("\n")
	} else {
		maxVisible := 14
		total := len(r.Returns)

		start := m.scroll
		end := start + maxVisible
		if end > total {
			end = total
		}

		var tbl strings.Builder
		header := fmt.Sprintf(" %-5s %-6s %-28s %8s %10s",
			"#", "Venta", "Razón", "Arts.", "Monto")
		tbl.WriteString(tableHeaderRow.Render(header) + "\n")

		for i := start; i < end; i++ {
			rs := r.Returns[i]
			reason := rs.Reason
			if reason == "" {
				reason = "(sin razón)"
			}
			line := fmt.Sprintf(" %-5d %-6d %-28s %8s %10s",
				rs.ID, rs.SaleID, truncate(reason, 28), fmtI(rs.ItemCount), fmtP(rs.Total))
			tbl.WriteString(warnStyle.Render(line) + "\n")
		}
		b.WriteString(tableBoxStyle.Render(tbl.String()))

		if total > maxVisible {
			b.WriteString("\n" + dimStyle.Render(fmt.Sprintf("  Mostrando %s-%s de %s devoluciones",
				fmtI(start+1), fmtI(end), fmtI(total))))
		}
	}

	b.WriteString("\n")
	b.WriteString("  " + hKey(hkNav, "↑↓", "desplazar") + hSep() + hKey(hkNav, "← →", "cambiar día") + hSep() + hKey(hkNav, "esc/enter", "volver al menú"))
	b.WriteString("\n")
	return b.String()
}
