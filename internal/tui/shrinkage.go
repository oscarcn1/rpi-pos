package tui

import (
	"fmt"
	"pos/internal/models"
	"pos/internal/store"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type shrinkageState int

const (
	shrinkageCode shrinkageState = iota
	shrinkageQty
	shrinkageReason
	shrinkageDone
)

type shrinkageModel struct {
	store   *store.Store
	state   shrinkageState
	input   textinput.Model
	product *models.Product
	qty     float64
	message string
	isError bool
}

func newShrinkageModel(s *store.Store) shrinkageModel {
	ti := textinput.New()
	ti.Placeholder = "Código del producto"
	ti.Focus()
	ti.CharLimit = 50
	return shrinkageModel{store: s, input: ti}
}

func (m shrinkageModel) update(msg tea.Msg) (shrinkageModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.state == shrinkageDone || m.state == shrinkageCode {
				return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
			}
			m.state = shrinkageCode
			m.product = nil
			m.input.SetValue("")
			m.input.Placeholder = "Código del producto"
			m.message = ""
			return m, nil

		case "enter":
			return m.handleEnter()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m shrinkageModel) handleEnter() (shrinkageModel, tea.Cmd) {
	val := strings.TrimSpace(m.input.Value())

	switch m.state {
	case shrinkageCode:
		if val == "" {
			return m, nil
		}
		p, err := m.store.GetProductByCode(val)
		if err != nil {
			m.message = fmt.Sprintf("Producto '%s' no encontrado", val)
			m.isError = true
			m.input.SetValue("")
			return m, nil
		}
		m.product = p
		m.state = shrinkageQty
		m.input.SetValue("")
		if p.IsMeasured() {
			m.input.Placeholder = fmt.Sprintf("Cantidad de merma en %s (ej: 0.5)", p.MeasurementUnit)
		} else {
			m.input.Placeholder = "Cantidad de merma (piezas)"
		}
		m.message = fmt.Sprintf("%s (stock actual: %s)", p.Name, p.StockLabel())
		m.isError = false

	case shrinkageQty:
		var qty float64
		val = stripCommas(val)
		if m.product.IsMeasured() {
			var err error
			qty, err = strconv.ParseFloat(val, 64)
			if err != nil || qty <= 0 {
				m.message = "Cantidad inválida (usa decimales, ej: 0.5)"
				m.isError = true
				return m, nil
			}
		} else {
			intQty, err := strconv.Atoi(val)
			if err != nil || intQty <= 0 {
				m.message = "Cantidad inválida"
				m.isError = true
				return m, nil
			}
			qty = float64(intQty)
		}
		if qty > m.product.Stock {
			m.message = fmt.Sprintf("Solo hay %s en existencia", m.product.StockLabel())
			m.isError = true
			return m, nil
		}
		m.qty = qty
		m.state = shrinkageReason
		m.input.SetValue("")
		m.input.Placeholder = "Razón (dañado, error de medición, caducado, etc.)"
		m.message = ""

	case shrinkageReason:
		sh := &models.Shrinkage{
			ProductID: m.product.ID,
			Quantity:  m.qty,
			Reason:    val,
		}
		if err := m.store.CreateShrinkage(sh); err != nil {
			m.message = "Error: " + err.Error()
			m.isError = true
			return m, nil
		}
		m.state = shrinkageDone
		m.message = fmt.Sprintf("Merma registrada: %s de %s", m.product.FormatQty(m.qty), m.product.Name)
		m.isError = false

	case shrinkageDone:
		return m, func() tea.Msg { return switchScreenMsg{screen: screenMenu} }
	}

	return m, nil
}

func (m shrinkageModel) view() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(titleStyle.Render(" Registrar Merma "))
	b.WriteString("\n")

	switch m.state {
	case shrinkageCode:
		b.WriteString(instructionStyle.Render("Escribe el código del producto que tuvo pérdida o daño."))
	case shrinkageQty:
		if m.product != nil && m.product.IsMeasured() {
			b.WriteString(instructionStyle.Render("Escribe la cantidad perdida. Acepta decimales para productos por medida."))
		} else {
			b.WriteString(instructionStyle.Render("Escribe la cantidad de piezas perdidas o dañadas."))
		}
	case shrinkageReason:
		b.WriteString(instructionStyle.Render("Describe brevemente la razón de la merma."))
	}
	b.WriteString("\n\n")

	if m.message != "" {
		if m.isError {
			b.WriteString("  " + errorStyle.Render(m.message) + "\n")
		} else {
			b.WriteString("  " + successStyle.Render(m.message) + "\n")
		}
	}

	if m.state == shrinkageDone {
		b.WriteString("\n")
		b.WriteString("  " + hKey(hkOk, "enter", "continuar") + hSep() + hKey(hkNav, "esc", "menú"))
	} else {
		b.WriteString("\n  " + m.input.View() + "\n\n")
		b.WriteString("  " + hKey(hkOk, "enter", "confirmar") + hSep() + hKey(hkDel, "esc", "cancelar/menú"))
	}

	b.WriteString("\n")
	return b.String()
}
