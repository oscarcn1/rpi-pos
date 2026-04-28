package tui

import (
	"fmt"
	"pos/internal/models"
	"pos/internal/store"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type returnState int

const (
	returnListSales    returnState = iota // browsing sales list
	returnViewSale                        // viewing sale detail, picking items
	returnConfirmFull                     // confirming full return
	returnPickItem                        // selecting item for partial return
	returnInputReason                     // entering reason
	returnDone                            // return completed
)

type salesLoadedMsg []models.Sale
type saleDetailLoadedMsg *models.SaleDetail

type returnsModel struct {
	store     *store.Store
	state     returnState
	sales     []models.Sale
	salesCur  int
	detail    *models.SaleDetail
	available []float64 // available qty per item (after previous returns)
	itemCur   int
	selected  []bool // items selected for return
	input     textinput.Model
	reason    string
	result    *models.Return
	message   string
	isError   bool
}

func newReturnsModel(s *store.Store) returnsModel {
	ti := textinput.New()
	ti.Placeholder = "Razón de la devolución"
	ti.CharLimit = 100
	return returnsModel{store: s, input: ti}
}

func (m returnsModel) loadSales() tea.Cmd {
	return func() tea.Msg {
		sales, _ := m.store.ListSalesDesc(100)
		return salesLoadedMsg(sales)
	}
}

func (m returnsModel) loadDetail(saleID int64) tea.Cmd {
	return func() tea.Msg {
		detail, _ := m.store.GetSaleDetail(saleID)
		return saleDetailLoadedMsg(detail)
	}
}

func (m returnsModel) update(msg tea.Msg) (returnsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case salesLoadedMsg:
		m.sales = msg
		m.salesCur = 0
		return m, nil

	case saleDetailLoadedMsg:
		m.detail = msg
		m.itemCur = 0
		// Calculate available quantities
		m.available = make([]float64, len(msg.Items))
		m.selected = make([]bool, len(msg.Items))
		for i, item := range msg.Items {
			returned := m.store.ReturnedQtyForSaleItem(item.ID)
			m.available[i] = item.Quantity - returned
		}
		m.state = returnViewSale
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case returnListSales:
			return m.updateListSales(msg)
		case returnViewSale:
			return m.updateViewSale(msg)
		case returnPickItem:
			return m.updatePickItem(msg)
		case returnInputReason:
			return m.updateInputReason(msg)
		case returnDone:
			if msg.String() == "enter" || msg.String() == "esc" {
				return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
			}
		}
	}

	if m.state == returnInputReason {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m returnsModel) updateListSales(msg tea.KeyMsg) (returnsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
	case "up", "k":
		if m.salesCur > 0 {
			m.salesCur--
		}
	case "down", "j":
		if m.salesCur < len(m.sales)-1 {
			m.salesCur++
		}
	case "enter":
		if len(m.sales) > 0 {
			sale := m.sales[m.salesCur]
			return m, m.loadDetail(sale.ID)
		}
	}
	return m, nil
}

func (m returnsModel) updateViewSale(msg tea.KeyMsg) (returnsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = returnListSales
		m.detail = nil
		m.message = ""
		return m, nil
	case "t":
		// Full return — check if anything is available
		hasAvailable := false
		for _, a := range m.available {
			if a > 0 {
				hasAvailable = true
				break
			}
		}
		if !hasAvailable {
			m.message = "Esta venta ya fue devuelta completamente"
			m.isError = true
			return m, nil
		}
		m.state = returnInputReason
		m.input.SetValue("")
		m.input.Placeholder = "Razón de la devolución total"
		m.input.Focus()
		// Pre-select all available items
		for i, a := range m.available {
			m.selected[i] = a > 0
		}
		m.message = ""
		return m, nil
	case "p":
		// Partial return
		hasAvailable := false
		for _, a := range m.available {
			if a > 0 {
				hasAvailable = true
				break
			}
		}
		if !hasAvailable {
			m.message = "Esta venta ya fue devuelta completamente"
			m.isError = true
			return m, nil
		}
		m.state = returnPickItem
		m.itemCur = 0
		for i := range m.selected {
			m.selected[i] = false
		}
		m.message = ""
		// Move cursor to first available item
		for i, a := range m.available {
			if a > 0 {
				m.itemCur = i
				break
			}
		}
		return m, nil
	}
	return m, nil
}

func (m returnsModel) updatePickItem(msg tea.KeyMsg) (returnsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = returnViewSale
		m.message = ""
		return m, nil
	case "up", "k":
		for i := m.itemCur - 1; i >= 0; i-- {
			if m.available[i] > 0 {
				m.itemCur = i
				break
			}
		}
	case "down", "j":
		for i := m.itemCur + 1; i < len(m.detail.Items); i++ {
			if m.available[i] > 0 {
				m.itemCur = i
				break
			}
		}
	case " ":
		if m.available[m.itemCur] > 0 {
			m.selected[m.itemCur] = !m.selected[m.itemCur]
		}
	case "enter":
		anySelected := false
		for _, s := range m.selected {
			if s {
				anySelected = true
				break
			}
		}
		if !anySelected {
			m.message = "Selecciona al menos un artículo con Espacio"
			m.isError = true
			return m, nil
		}
		m.state = returnInputReason
		m.input.SetValue("")
		m.input.Placeholder = "Razón de la devolución"
		m.input.Focus()
		m.message = ""
		return m, nil
	}
	return m, nil
}

func (m returnsModel) updateInputReason(msg tea.KeyMsg) (returnsModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = returnViewSale
		m.message = ""
		return m, nil
	case "enter":
		reason := strings.TrimSpace(m.input.Value())

		ret := &models.Return{
			SaleID: m.detail.Sale.ID,
			Reason: reason,
		}

		for i, item := range m.detail.Items {
			if m.selected[i] && m.available[i] > 0 {
				ri := models.ReturnItem{
					SaleItemID:  item.ID,
					ProductID:   item.ProductID,
					ProductName: item.ProductName,
					Quantity:    m.available[i],
					UnitPrice:   item.UnitPrice,
					Subtotal:    m.available[i] * item.UnitPrice,
				}
				ret.Total += ri.Subtotal
				ret.Items = append(ret.Items, ri)
			}
		}

		if err := m.store.CreateReturn(ret); err != nil {
			m.message = "Error: " + err.Error()
			m.isError = true
			return m, nil
		}

		m.result = ret
		m.state = returnDone
		m.message = ""
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// --- View ---

var (
	returnedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Strikethrough(true)
	checkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
)

func (m returnsModel) view() string {
	var b strings.Builder

	b.WriteString(screenTitleStyle.Render("Devoluciones"))
	b.WriteString("\n\n")

	if m.message != "" {
		if m.isError {
			b.WriteString("  " + errorStyle.Render(m.message) + "\n\n")
		} else {
			b.WriteString("  " + successStyle.Render(m.message) + "\n\n")
		}
	}

	switch m.state {
	case returnListSales:
		m.viewSalesList(&b)
	case returnViewSale, returnPickItem:
		m.viewSaleDetail(&b)
	case returnInputReason:
		m.viewSaleDetail(&b)
		b.WriteString("  " + m.input.View() + "\n")
	case returnDone:
		m.viewDone(&b)
	}

	b.WriteString("\n")
	b.WriteString(m.viewHelp())
	b.WriteString("\n")

	return b.String()
}

func (m returnsModel) viewSalesList(b *strings.Builder) {
	if len(m.sales) == 0 {
		b.WriteString(dimStyle.Render("  No hay ventas registradas\n"))
		return
	}

	start := m.salesCur - 5
	if start < 0 {
		start = 0
	}
	end := start + 10
	if end > len(m.sales) {
		end = len(m.sales)
	}

	var tbl strings.Builder
	header := fmt.Sprintf(" %-6s %-20s %14s",
		"#", "Fecha", "Total")
	tbl.WriteString(tableHeaderRow.Render(header) + "\n")

	for i := start; i < end; i++ {
		s := m.sales[i]
		dateStr := s.CreatedAt.Format("2006-01-02 15:04")
		line := fmt.Sprintf(" %-6d %-20s %14s",
			s.ID, dateStr, fmtP(s.Total))
		if i == m.salesCur {
			tbl.WriteString(tableRowSelected.Render(padRight(line, 44)) + "\n")
		} else {
			tbl.WriteString(tableRowNormal.Render(line) + "\n")
		}
	}
	b.WriteString(tableBox("75").Render(tbl.String()))

	if len(m.sales) > 10 {
		b.WriteString("\n" + dimStyle.Render(fmt.Sprintf("  Mostrando %s-%s de %s ventas",
			fmtI(start+1), fmtI(end), fmtI(len(m.sales)))))
	}
	b.WriteString("\n")
}

func (m returnsModel) viewSaleDetail(b *strings.Builder) {
	if m.detail == nil {
		return
	}

	s := m.detail.Sale
	b.WriteString(fmt.Sprintf("  Venta #%d — %s — Total: %s\n\n",
		s.ID, s.CreatedAt.Format("2006-01-02 15:04"), fmtP(s.Total)))

	var tbl strings.Builder
	header := fmt.Sprintf("   %-26s %8s %10s %10s %10s",
		"Producto", "Cant.", "Precio", "Subtotal", "Disponible")
	tbl.WriteString(tableHeaderRow.Render(header) + "\n")

	for i, item := range m.detail.Items {
		avail := m.available[i]
		availStr := fmtQ(avail)
		if avail <= 0 {
			availStr = "devuelto"
		}

		prefix := "  "
		if m.state == returnPickItem {
			if m.selected[i] {
				prefix = checkStyle.Render("✓ ")
			}
		}

		line := fmt.Sprintf("%s %-26s %8s %10s %10s %10s",
			prefix,
			truncate(item.ProductName, 26), fmtQ(item.Quantity),
			fmtP(item.UnitPrice), fmtP(item.Subtotal), availStr)

		if avail <= 0 {
			tbl.WriteString(returnedStyle.Render(line) + "\n")
		} else if m.state == returnPickItem && i == m.itemCur {
			tbl.WriteString(tableRowSelected.Render(padRight(line, 72)) + "\n")
		} else if m.selected[i] {
			tbl.WriteString(successStyle.Render(line) + "\n")
		} else {
			tbl.WriteString(tableRowNormal.Render(line) + "\n")
		}
	}
	b.WriteString(tableBox("75").Render(tbl.String()))
	b.WriteString("\n")

	// Show return total if items selected
	if m.state == returnPickItem || m.state == returnInputReason {
		total := 0.0
		count := 0
		for i, item := range m.detail.Items {
			if m.selected[i] && m.available[i] > 0 {
				total += m.available[i] * item.UnitPrice
				count++
			}
		}
		if count > 0 {
			b.WriteString(fmt.Sprintf("\n  %s %s (%s artículos)\n\n",
				warnStyle.Render("Monto a devolver:"), fmtP(total), fmtI(count)))
		}
	}
}

func (m returnsModel) viewDone(b *strings.Builder) {
	if m.result == nil {
		return
	}

	r := m.result
	info := fmt.Sprintf(
		"  DEVOLUCIÓN REGISTRADA\n\n"+
			"  Venta original: #%d\n"+
			"  Artículos devueltos: %s\n"+
			"  Monto devuelto: %s\n"+
			"  Razón: %s\n\n"+
			"  El stock ha sido restaurado.",
		r.SaleID, fmtI(len(r.Items)), fmtP(r.Total), r.Reason)
	b.WriteString(boxStyle.Render(info))
	b.WriteString("\n")
}

func (m returnsModel) viewHelp() string {
	s := hSep()
	switch m.state {
	case returnListSales:
		return "  " + hKey(hkNav, "↑↓", "navegar") + s + hKey(hkOk, "enter", "ver venta") + s + hKey(hkNav, "esc", "menú")
	case returnViewSale:
		return "  " + hKey(hkAct, "t", "devolución total") + s + hKey(hkAct, "p", "devolución parcial") + s + hKey(hkNav, "esc", "volver")
	case returnPickItem:
		return "  " + hKey(hkNav, "↑↓", "navegar") + s + hKey(hkOk, "espacio", "marcar/desmarcar") + s + hKey(hkOk, "enter", "confirmar") + s + hKey(hkNav, "esc", "volver")
	case returnInputReason:
		return "  " + hKey(hkOk, "enter", "confirmar devolución") + s + hKey(hkDel, "esc", "cancelar")
	case returnDone:
		return "  " + hKey(hkNav, "enter/esc", "volver al menú")
	}
	return ""
}
