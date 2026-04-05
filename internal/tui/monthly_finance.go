package tui

import (
	"fmt"
	"pos/internal/models"
	"pos/internal/store"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var monthNames = []string{
	"", "Enero", "Febrero", "Marzo", "Abril", "Mayo", "Junio",
	"Julio", "Agosto", "Septiembre", "Octubre", "Noviembre", "Diciembre",
}

var (
	barStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	barDimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	monthNavStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")).
			Padding(0, 1)
	pctUpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	pctDownStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	chartLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

type monthlyFinanceLoadedMsg *models.MonthlyFinanceReport

type monthlyFinanceModel struct {
	store  *store.Store
	report *models.MonthlyFinanceReport
	year   int
	month  int
}

func newMonthlyFinanceModel(s *store.Store) monthlyFinanceModel {
	now := time.Now()
	return monthlyFinanceModel{
		store: s,
		year:  now.Year(),
		month: int(now.Month()),
	}
}

func (m monthlyFinanceModel) load() tea.Cmd {
	return func() tea.Msg {
		r, _ := m.store.MonthlyFinanceReport(m.year, m.month)
		return monthlyFinanceLoadedMsg(r)
	}
}

func (m monthlyFinanceModel) update(msg tea.Msg) (monthlyFinanceModel, tea.Cmd) {
	switch msg := msg.(type) {
	case monthlyFinanceLoadedMsg:
		m.report = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter":
			return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
		case "left", "h":
			m.month--
			if m.month < 1 {
				m.month = 12
				m.year--
			}
			m.report = nil
			return m, m.load()
		case "right", "l":
			now := time.Now()
			ny, nm := m.year, m.month+1
			if nm > 12 {
				nm = 1
				ny++
			}
			if ny < now.Year() || (ny == now.Year() && nm <= int(now.Month())) {
				m.month = nm
				m.year = ny
				m.report = nil
				return m, m.load()
			}
		}
	}
	return m, nil
}

func (m monthlyFinanceModel) view() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(" Finanzas Mensuales "))
	b.WriteString("\n")
	b.WriteString(instructionStyle.Render("Resumen financiero del mes con gráfica de ventas diarias."))
	b.WriteString("\n\n")

	// Month navigation
	nav := fmt.Sprintf("◀  %s %d  ▶", monthNames[m.month], m.year)
	b.WriteString("  " + monthNavStyle.Render(nav))
	b.WriteString("\n\n")

	if m.report == nil {
		b.WriteString(dimStyle.Render("  Cargando...\n"))
		b.WriteString("\n")
		b.WriteString("  " + hKey(hkNav, "←→/hl", "cambiar mes") + hSep() + hKey(hkNav, "esc/enter", "volver al menú"))
		b.WriteString("\n")
		return b.String()
	}

	r := m.report

	// Summary box
	summary := fmt.Sprintf(
		"  Ventas realizadas:  %10s\n"+
			"  Ingresos totales:   %10s\n"+
			"  Costo total:        %10s\n"+
			"  Ganancia:           %10s\n"+
			"  Promedio diario:    %10s\n"+
			"  Merma del mes:      %10s unidades",
		fmtI(r.TotalSales), fmtP(r.TotalIncome), fmtP(r.TotalCost),
		fmtP(r.Profit), fmtP(r.AvgDailySales), fmtQ(r.TotalShrinkage))
	b.WriteString(boxStyle.Render(summary))
	b.WriteString("\n\n")

	// Comparison with previous month
	if r.HasPrev {
		b.WriteString("  " + subtitleStyle.Render("vs. mes anterior:") + "  ")
		b.WriteString(fmtPct("Ingresos", r.TotalIncome, r.PrevIncome))
		b.WriteString("  ")
		b.WriteString(fmtPct("Ganancia", r.Profit, r.PrevProfit))
		b.WriteString("  ")
		b.WriteString(fmtPct("Ventas", float64(r.TotalSales), float64(r.PrevSales)))
		b.WriteString("\n\n")
	}

	// Bar chart
	m.viewChart(&b, r)

	b.WriteString("\n")
	b.WriteString("  " + hKey(hkNav, "←→/hl", "cambiar mes") + hSep() + hKey(hkNav, "esc/enter", "volver al menú"))
	b.WriteString("\n")
	return b.String()
}

func (m monthlyFinanceModel) viewChart(b *strings.Builder, r *models.MonthlyFinanceReport) {
	b.WriteString(subtitleStyle.Render("  Ventas diarias:"))
	b.WriteString("\n\n")

	if r.TotalSales == 0 {
		b.WriteString(dimStyle.Render("  Sin ventas este mes\n"))
		return
	}

	// Find max daily total for normalization
	maxDaily := 0.0
	for _, ds := range r.DailySales {
		if ds.Total > maxDaily {
			maxDaily = ds.Total
		}
	}

	blocks := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	chartHeight := 6

	// Render vertical bar chart (8 rows, top to bottom)
	for row := chartHeight; row >= 1; row-- {
		line := "  "
		// Y-axis label on leftmost row positions
		if row == chartHeight {
			line += chartLabelStyle.Render(fmt.Sprintf("%8s │", fmtP(maxDaily)))
		} else if row == chartHeight/2 {
			line += chartLabelStyle.Render(fmt.Sprintf("%8s │", fmtP(maxDaily/2)))
		} else if row == 1 {
			line += chartLabelStyle.Render(fmt.Sprintf("%8s │", "$0"))
		} else {
			line += chartLabelStyle.Render("         │")
		}

		for _, ds := range r.DailySales {
			level := 0
			if maxDaily > 0 {
				level = int(ds.Total / maxDaily * float64(chartHeight))
				if ds.Total > 0 && level == 0 {
					level = 1
				}
			}

			if level >= row {
				line += barStyle.Render("█")
			} else if level == row-1 && ds.Total > 0 {
				// Partial block for transition
				frac := (ds.Total/maxDaily*float64(chartHeight) - float64(level))
				idx := int(frac * float64(len(blocks)-1))
				if idx < 0 {
					idx = 0
				}
				if idx >= len(blocks) {
					idx = len(blocks) - 1
				}
				line += barStyle.Render(string(blocks[idx]))
			} else {
				line += barDimStyle.Render("·")
			}
		}
		b.WriteString(line + "\n")
	}

	// X-axis
	b.WriteString("  " + chartLabelStyle.Render("         └"))
	for i := 0; i < r.DaysInMonth; i++ {
		b.WriteString(chartLabelStyle.Render("─"))
	}
	b.WriteString("\n")

	// Day labels
	b.WriteString("  " + chartLabelStyle.Render("          "))
	for i := 1; i <= r.DaysInMonth; i++ {
		if i == 1 || i%5 == 0 || i == r.DaysInMonth {
			b.WriteString(chartLabelStyle.Render(fmt.Sprintf("%d", i%10)))
		} else {
			b.WriteString(" ")
		}
	}
	b.WriteString("\n")

	// Best day
	bestDay := 0
	bestTotal := 0.0
	for _, ds := range r.DailySales {
		if ds.Total > bestTotal {
			bestTotal = ds.Total
			bestDay = ds.Day
		}
	}
	if bestDay > 0 {
		b.WriteString(fmt.Sprintf("\n  %s día %d con %s en ventas",
			dimStyle.Render("Mejor día:"), bestDay, fmtP(bestTotal)))
		b.WriteString("\n")
	}
}

func fmtPct(label string, current, prev float64) string {
	if prev == 0 {
		return dimStyle.Render(label + " N/A")
	}
	pct := (current - prev) / prev * 100
	sign := "+"
	style := pctUpStyle
	if pct < 0 {
		sign = ""
		style = pctDownStyle
	}
	return style.Render(fmt.Sprintf("%s %s%.1f%%", label, sign, pct))
}
