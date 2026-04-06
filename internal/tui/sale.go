package tui

import (
	"fmt"
	"pos/internal/models"
	"pos/internal/store"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const saleWidth = 72

type saleState int

const (
	saleSearching    saleState = iota
	salePickResult
	saleInputQty
	saleCartManage
	saleInputPayment
	saleDone
)

type saleModel struct {
	store   *store.Store
	state   saleState
	input   textinput.Model
	cart    []models.SaleItem
	total   float64
	results []models.Product
	resCur  int
	cartCur int
	current *models.Product
	message string
	isError bool
	payment float64
	change  float64
}

func newSaleModel(s *store.Store) saleModel {
	ti := textinput.New()
	ti.Placeholder = "Buscar producto (nombre o código)"
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = saleWidth - 6
	return saleModel{
		store: s,
		state: saleSearching,
		input: ti,
	}
}

func (m saleModel) update(msg tea.Msg) (saleModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case saleSearching:
			return m.updateSearching(msg)
		case salePickResult:
			return m.updatePickResult(msg)
		case saleInputQty:
			return m.updateInputQty(msg)
		case saleCartManage:
			return m.updateCartManage(msg)
		case saleInputPayment:
			return m.updateInputPayment(msg)
		case saleDone:
			return m.updateDone(msg)
		}
	}

	if m.state == saleSearching || m.state == saleInputQty || m.state == saleInputPayment {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m saleModel) updateSearching(msg tea.KeyMsg) (saleModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if len(m.cart) == 0 {
			return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
		}
		m.cart = nil
		m.total = 0
		m.message = "Venta cancelada"
		m.isError = false
		m.results = nil
		return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }

	case "enter":
		query := strings.TrimSpace(m.input.Value())
		if query == "" {
			return m, nil
		}
		results, err := m.store.FuzzySearchProducts(query)
		if err != nil || len(results) == 0 {
			m.message = fmt.Sprintf("No se encontró '%s'", query)
			m.isError = true
			m.results = nil
			return m, nil
		}
		if len(results) == 1 {
			p := results[0]
			if strings.EqualFold(p.Code, query) {
				return m.selectProduct(&p)
			}
		}
		m.results = results
		m.resCur = 0
		m.state = salePickResult
		m.message = ""
		return m, nil

	case "f2":
		if len(m.cart) > 0 {
			return m.goToPayment()
		}

	case "f3", "tab":
		if len(m.cart) > 0 {
			m.state = saleCartManage
			m.cartCur = len(m.cart) - 1
			m.message = ""
			m.results = nil
			return m, nil
		}

	case "down":
		if len(m.results) > 0 {
			m.state = salePickResult
			m.resCur = 0
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m saleModel) updatePickResult(msg tea.KeyMsg) (saleModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = saleSearching
		m.results = nil
		m.input.Focus()
		return m, nil
	case "up", "k":
		if m.resCur > 0 {
			m.resCur--
		} else {
			m.state = saleSearching
			m.input.Focus()
			return m, nil
		}
	case "down", "j":
		if m.resCur < len(m.results)-1 {
			m.resCur++
		}
	case "enter":
		if len(m.results) > 0 {
			p := m.results[m.resCur]
			return m.selectProduct(&p)
		}
	}
	return m, nil
}

func (m saleModel) selectProduct(p *models.Product) (saleModel, tea.Cmd) {
	if p.Stock <= 0 {
		m.message = fmt.Sprintf("'%s' sin existencias", p.Name)
		m.isError = true
		m.state = saleSearching
		m.input.SetValue("")
		m.input.Focus()
		return m, nil
	}
	inCart := 0.0
	for _, item := range m.cart {
		if item.ProductID == p.ID {
			inCart += item.Quantity
		}
	}
	available := p.Stock - inCart
	if available <= 0 {
		m.message = fmt.Sprintf("'%s' ya tiene todas las existencias en el carrito", p.Name)
		m.isError = true
		m.state = saleSearching
		m.input.SetValue("")
		m.input.Focus()
		return m, nil
	}

	m.current = p
	m.state = saleInputQty
	if p.IsMeasured() {
		m.input.SetValue("")
		m.input.Placeholder = fmt.Sprintf("Cantidad en %s (ej: 2.5)", p.MeasurementUnit)
		m.message = fmt.Sprintf("%s — %s/%s (disponible: %s)", p.Name, fmtP(p.SalePrice), p.MeasurementUnit, p.FormatQty(available))
	} else {
		m.input.SetValue("1")
		m.input.Placeholder = "Cantidad (piezas)"
		m.message = fmt.Sprintf("%s — %s/pza (disponible: %s)", p.Name, fmtP(p.SalePrice), fmtQ(available))
	}
	m.input.Focus()
	m.isError = false
	return m, nil
}

func (m saleModel) updateInputQty(msg tea.KeyMsg) (saleModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = saleSearching
		m.current = nil
		m.input.SetValue("")
		m.input.Placeholder = "Buscar producto (nombre o código)"
		m.input.Focus()
		m.message = ""
		return m, nil

	case "enter":
		val := stripCommas(strings.TrimSpace(m.input.Value()))
		var qty float64

		if m.current.IsMeasured() {
			var err error
			qty, err = strconv.ParseFloat(val, 64)
			if err != nil || qty <= 0 {
				m.message = "Cantidad inválida (usa decimales, ej: 2.5)"
				m.isError = true
				return m, nil
			}
		} else {
			intQty, err := strconv.Atoi(val)
			if err != nil || intQty <= 0 {
				m.message = "Cantidad inválida (debe ser un número entero)"
				m.isError = true
				return m, nil
			}
			qty = float64(intQty)
		}

		inCart := 0.0
		for _, item := range m.cart {
			if item.ProductID == m.current.ID {
				inCart += item.Quantity
			}
		}
		available := m.current.Stock - inCart
		if qty > available {
			m.message = fmt.Sprintf("Solo hay %s disponible(s)", m.current.FormatQty(available))
			m.isError = true
			return m, nil
		}

		item := models.SaleItem{
			ProductID:   m.current.ID,
			ProductName: m.current.Name,
			Quantity:    qty,
			UnitPrice:   m.current.SalePrice,
			Subtotal:    qty * m.current.SalePrice,
			CostPerUnit: m.current.CostPerUnit(),
		}
		m.cart = append(m.cart, item)
		m.total += item.Subtotal
		qtyLabel := m.current.FormatQty(qty)
		m.current = nil
		m.results = nil
		m.state = saleSearching
		m.input.SetValue("")
		m.input.Placeholder = "Buscar producto (nombre o código)"
		m.input.Focus()
		m.message = fmt.Sprintf("+ %s x %s = %s", qtyLabel, item.ProductName, fmtP(item.Subtotal))
		m.isError = false
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m saleModel) updateCartManage(msg tea.KeyMsg) (saleModel, tea.Cmd) {
	switch msg.String() {
	case "esc", "tab", "f3":
		m.state = saleSearching
		m.input.Focus()
		m.message = ""
		return m, nil
	case "up", "k":
		if m.cartCur > 0 {
			m.cartCur--
		}
	case "down", "j":
		if m.cartCur < len(m.cart)-1 {
			m.cartCur++
		}
	case "d", "delete", "backspace":
		if len(m.cart) > 0 {
			removed := m.cart[m.cartCur]
			m.total -= removed.Subtotal
			m.cart = append(m.cart[:m.cartCur], m.cart[m.cartCur+1:]...)
			if m.cartCur >= len(m.cart) && m.cartCur > 0 {
				m.cartCur--
			}
			m.message = fmt.Sprintf("- %s eliminado", removed.ProductName)
			m.isError = false
			if len(m.cart) == 0 {
				m.state = saleSearching
				m.input.Focus()
			}
		}
	case "x":
		m.cart = nil
		m.total = 0
		m.state = saleSearching
		m.input.SetValue("")
		m.input.Focus()
		m.message = "Venta cancelada"
		m.isError = false
		return m, nil
	}
	return m, nil
}

func (m saleModel) goToPayment() (saleModel, tea.Cmd) {
	m.state = saleInputPayment
	m.input.SetValue("")
	m.input.Placeholder = fmt.Sprintf("Pago (total: %s)", fmtP(m.total))
	m.input.Focus()
	m.message = ""
	m.results = nil
	return m, nil
}

func (m saleModel) updateInputPayment(msg tea.KeyMsg) (saleModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = saleSearching
		m.input.SetValue("")
		m.input.Placeholder = "Buscar producto (nombre o código)"
		m.input.Focus()
		m.message = ""
		return m, nil

	case "enter":
		val := stripCommas(strings.TrimSpace(m.input.Value()))
		pay, err := strconv.ParseFloat(val, 64)
		if err != nil || pay < m.total {
			m.message = fmt.Sprintf("Monto insuficiente (mínimo: %s)", fmtP(m.total))
			m.isError = true
			return m, nil
		}
		m.payment = pay
		m.change = pay - m.total
		sale := &models.Sale{
			Items:        m.cart,
			Total:        m.total,
			Payment:      m.payment,
			ChangeAmount: m.change,
		}
		if err := m.store.CreateSale(sale); err != nil {
			m.message = "Error al guardar: " + err.Error()
			m.isError = true
			return m, nil
		}
		m.state = saleDone
		m.message = ""
		m.isError = false
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m saleModel) updateDone(msg tea.KeyMsg) (saleModel, tea.Cmd) {
	switch msg.String() {
	case "enter":
		return newSaleModel(m.store), nil
	case "esc":
		return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
	}
	return m, nil
}

// --- View ---

var (
	saleBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Width(saleWidth).
			Padding(0, 1)

	cartHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63")).
			Width(saleWidth).
			Align(lipgloss.Center).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("63"))

	cartSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("52")).
				Foreground(lipgloss.Color("231")).
				Bold(true)

	totalBarStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color("229")).
			Width(saleWidth + 4).
			Align(lipgloss.Right).
			Padding(0, 2)

	resultSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("27")).
				Foreground(lipgloss.Color("255")).
				Bold(true)

	resultNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	resultHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Bold(true)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("57")).
			Width(saleWidth).
			Padding(0, 1)
)

func (m saleModel) view() string {
	var b strings.Builder

	b.WriteString(subtitleStyle.Render("  Nueva Venta"))
	b.WriteString("\n\n")

	if m.state == saleDone {
		m.viewReceipt(&b)
		m.viewCart(&b)
		return b.String()
	}

	if m.state != saleCartManage {
		b.WriteString(inputBoxStyle.Render(m.input.View()))
		b.WriteString("\n")
		b.WriteString(m.viewHelp())
		b.WriteString("\n")
	} else {
		b.WriteString(m.viewHelp())
		b.WriteString("\n")
	}

	if m.message != "" {
		b.WriteString("\n")
		if m.isError {
			b.WriteString("  " + errorStyle.Render(m.message))
		} else {
			b.WriteString("  " + successStyle.Render(m.message))
		}
		b.WriteString("\n")
	}

	if (m.state == saleSearching || m.state == salePickResult) && len(m.results) > 0 {
		b.WriteString("\n")

		maxVisible := 8
		total := len(m.results)

		start := m.resCur - maxVisible/2
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

		header := fmt.Sprintf(" %-8s %-26s %12s %12s",
			"Código", "Producto", "Precio", "Stock")
		b.WriteString("  " + resultHeaderStyle.Render(header) + "\n")

		for i := start; i < end; i++ {
			p := m.results[i]
			priceLabel := fmt.Sprintf("%s/%s", fmtP(p.SalePrice), p.UnitLabel())
			row := fmt.Sprintf(" %-8s %-26s %12s %12s",
				truncate(p.Code, 8), truncate(p.Name, 26), priceLabel, p.StockLabel())

			if m.state == salePickResult && i == m.resCur {
				padded := padRight(row, saleWidth)
				b.WriteString("  " + resultSelectedStyle.Render(padded) + "\n")
			} else if p.Stock <= 0 {
				b.WriteString("  " + dimStyle.Render(row) + "\n")
			} else {
				b.WriteString("  " + resultNormalStyle.Render(row) + "\n")
			}
		}

		if total > maxVisible {
			b.WriteString("  " + dimStyle.Render(fmt.Sprintf("Mostrando %s-%s de %s",
				fmtI(start+1), fmtI(end), fmtI(total))) + "\n")
		}
	}

	b.WriteString("\n")
	m.viewCart(&b)

	return b.String()
}

func (m saleModel) viewCart(b *strings.Builder) {
	if len(m.cart) == 0 {
		b.WriteString(cartHeaderStyle.Render("Artículos en la venta"))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  (vacío)") + "\n")
		return
	}

	b.WriteString(cartHeaderStyle.Render(fmt.Sprintf("Artículos en la venta (%d)", len(m.cart))))
	b.WriteString("\n")

	var lines strings.Builder
	for i, item := range m.cart {
		line := fmt.Sprintf(" %2d. %-26s %6s x %9s = %10s",
			i+1, truncate(item.ProductName, 26), fmtQ(item.Quantity), fmtP(item.UnitPrice), fmtP(item.Subtotal))

		if m.state == saleCartManage && i == m.cartCur {
			lines.WriteString(cartSelectedStyle.Render(padRight(line, saleWidth)))
		} else {
			lines.WriteString(line)
		}
		lines.WriteString("\n")
	}

	b.WriteString(saleBoxStyle.Render(lines.String()))
	b.WriteString("\n")
	b.WriteString(totalBarStyle.Render(fmt.Sprintf("TOTAL: %s ", fmtP(m.total))))
	b.WriteString("\n")
}

func (m saleModel) viewReceipt(b *strings.Builder) {
	receipt := fmt.Sprintf(
		"  VENTA COMPLETADA\n\n"+
			"  Total:   %12s\n"+
			"  Pago:    %12s\n"+
			"  Cambio:  %12s",
		fmtP(m.total), fmtP(m.payment), fmtP(m.change))
	b.WriteString(boxStyle.Render(receipt))
	b.WriteString("\n\n")
	b.WriteString("  " + hKey(hkOk, "enter", "nueva venta") + hSep() + hKey(hkNav, "esc", "menú"))
	b.WriteString("\n\n")
}

func (m saleModel) viewHelp() string {
	s := hSep()
	switch m.state {
	case saleSearching:
		h := "  " + hKey(hkAct, "enter", "buscar")
		if len(m.results) > 0 {
			h += s + hKey(hkNav, "↓", "resultados")
		}
		if len(m.cart) > 0 {
			h += s + hKey(hkOk, "F2", "cobrar") + s + hKey(hkAct, "F3", "editar carrito") + s + hKey(hkDel, "esc", "cancelar")
		} else {
			h += s + hKey(hkNav, "esc", "menú")
		}
		return h
	case salePickResult:
		return "  " + hKey(hkNav, "↑↓", "navegar") + s + hKey(hkOk, "enter", "seleccionar") + s + hKey(hkNav, "esc", "volver")
	case saleInputQty:
		return "  " + hKey(hkOk, "enter", "agregar al carrito") + s + hKey(hkDel, "esc", "cancelar")
	case saleCartManage:
		return "  " + hKey(hkNav, "↑↓", "navegar") + s + hKey(hkDel, "d", "eliminar artículo") + s + hKey(hkDel, "x", "cancelar venta") + s + hKey(hkNav, "esc/F3", "volver")
	case saleInputPayment:
		return "  " + hKey(hkOk, "enter", "confirmar pago") + s + hKey(hkNav, "esc", "seguir agregando")
	}
	return ""
}

func padRight(s string, width int) string {
	r := []rune(s)
	if len(r) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(r))
}
