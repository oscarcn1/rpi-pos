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

type productFormData struct {
	editing *models.Product
}

const (
	fieldCode = iota
	fieldName
	fieldDesc
	fieldCategory
	fieldUnitType
	fieldMeasurementUnit
	fieldUnitsPerPurchase
	fieldPurchasePrice
	fieldSalePrice
	fieldStock
	fieldMinStock
	fieldCount
)

var fieldLabels = [fieldCount]string{
	"Código", "Nombre", "Descripción", "Categoría",
	"Tipo venta", "Unidad medida", "Cantidad/compra",
	"P. Compra", "P. Venta", "Stock", "Stock Mínimo",
}

type productFormModel struct {
	store    *store.Store
	fields   [fieldCount]textinput.Model
	focus    int
	editing  *models.Product
	measured bool
	message  string
	isError  bool
}

func newProductFormModel(s *store.Store) productFormModel {
	m := productFormModel{store: s}
	m.initFields()
	return m
}

func newProductFormModelWith(s *store.Store, data *productFormData) productFormModel {
	m := productFormModel{store: s, editing: data.editing}
	m.initFields()
	if data.editing != nil {
		p := data.editing
		m.fields[fieldCode].SetValue(p.Code)
		m.fields[fieldName].SetValue(p.Name)
		m.fields[fieldDesc].SetValue(p.Description)
		m.fields[fieldCategory].SetValue(p.Category)
		if p.IsMeasured() {
			m.measured = true
			m.fields[fieldUnitType].SetValue("medida")
			m.fields[fieldMeasurementUnit].SetValue(p.MeasurementUnit)
			m.fields[fieldUnitsPerPurchase].SetValue(fmt.Sprintf("%.1f", p.UnitsPerPurchase))
		} else {
			m.fields[fieldUnitType].SetValue("unidad")
		}
		m.fields[fieldPurchasePrice].SetValue(fmt.Sprintf("%.2f", p.PurchasePrice))
		m.fields[fieldSalePrice].SetValue(fmt.Sprintf("%.2f", p.SalePrice))
		m.fields[fieldStock].SetValue(fmt.Sprintf("%.2f", p.Stock))
		m.fields[fieldMinStock].SetValue(fmt.Sprintf("%.2f", p.MinStock))
	}
	return m
}

func (m *productFormModel) initFields() {
	for i := range m.fields {
		ti := textinput.New()
		ti.Placeholder = fieldLabels[i]
		ti.CharLimit = 50
		m.fields[i] = ti
	}
	m.fields[fieldCode].CharLimit = 100
	m.fields[fieldUnitType].Placeholder = "unidad o medida (espacio para cambiar)"
	m.fields[fieldUnitType].SetValue("unidad")
	m.fields[fieldMeasurementUnit].Placeholder = "m, kg, L, etc."
	m.fields[fieldUnitsPerPurchase].Placeholder = "ej: 100 (metros por rollo)"
	m.fields[fieldPurchasePrice].CharLimit = 12
	m.fields[fieldSalePrice].CharLimit = 12
	m.fields[fieldStock].CharLimit = 12
	m.fields[fieldMinStock].CharLimit = 12
	m.fields[0].Focus()
}

func (m productFormModel) update(msg tea.Msg) (productFormModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return switchScreenMsg{screen: screenProducts}
			}
		case "tab", "down":
			m.fields[m.focus].Blur()
			m.focus = (m.focus + 1) % fieldCount
			// Skip measurement fields for unit products
			if !m.measured && (m.focus == fieldMeasurementUnit || m.focus == fieldUnitsPerPurchase) {
				m.focus = fieldPurchasePrice
			}
			m.fields[m.focus].Focus()
			return m, nil
		case "shift+tab", "up":
			m.fields[m.focus].Blur()
			m.focus = (m.focus - 1 + fieldCount) % fieldCount
			if !m.measured && (m.focus == fieldMeasurementUnit || m.focus == fieldUnitsPerPurchase) {
				m.focus = fieldUnitType
			}
			m.fields[m.focus].Focus()
			return m, nil
		case " ":
			if m.focus == fieldUnitType {
				m.measured = !m.measured
				if m.measured {
					m.fields[fieldUnitType].SetValue("medida")
				} else {
					m.fields[fieldUnitType].SetValue("unidad")
					m.fields[fieldMeasurementUnit].SetValue("")
					m.fields[fieldUnitsPerPurchase].SetValue("")
				}
				return m, nil
			}
		case "enter":
			if m.focus < fieldCount-1 {
				m.fields[m.focus].Blur()
				m.focus++
				if !m.measured && (m.focus == fieldMeasurementUnit || m.focus == fieldUnitsPerPurchase) {
					m.focus = fieldPurchasePrice
				}
				m.fields[m.focus].Focus()
				return m, nil
			}
			return m.save()
		case "ctrl+s":
			return m.save()
		}
	}

	if m.focus == fieldUnitType {
		// Don't pass keys to textinput for the toggle field
		return m, nil
	}

	var cmd tea.Cmd
	m.fields[m.focus], cmd = m.fields[m.focus].Update(msg)
	return m, cmd
}

func (m productFormModel) save() (productFormModel, tea.Cmd) {
	code := strings.TrimSpace(m.fields[fieldCode].Value())
	name := strings.TrimSpace(m.fields[fieldName].Value())
	if code == "" || name == "" {
		m.message = "Código y nombre son obligatorios"
		m.isError = true
		return m, nil
	}

	pp, err := strconv.ParseFloat(stripCommas(strings.TrimSpace(m.fields[fieldPurchasePrice].Value())), 64)
	if err != nil || pp < 0 {
		m.message = "Precio de compra inválido"
		m.isError = true
		return m, nil
	}

	sp, err := strconv.ParseFloat(stripCommas(strings.TrimSpace(m.fields[fieldSalePrice].Value())), 64)
	if err != nil || sp < 0 {
		m.message = "Precio de venta inválido"
		m.isError = true
		return m, nil
	}

	stock, _ := strconv.ParseFloat(stripCommas(strings.TrimSpace(m.fields[fieldStock].Value())), 64)
	minStock, _ := strconv.ParseFloat(stripCommas(strings.TrimSpace(m.fields[fieldMinStock].Value())), 64)

	p := &models.Product{
		Code:          code,
		Name:          name,
		Description:   strings.TrimSpace(m.fields[fieldDesc].Value()),
		Category:      strings.TrimSpace(m.fields[fieldCategory].Value()),
		PurchasePrice: pp,
		SalePrice:     sp,
		Stock:         stock,
		MinStock:      minStock,
		UnitType:      "unit",
	}

	if m.measured {
		mu := strings.TrimSpace(m.fields[fieldMeasurementUnit].Value())
		if mu == "" {
			m.message = "La unidad de medida es obligatoria para productos por medida"
			m.isError = true
			return m, nil
		}
		upp, err := strconv.ParseFloat(stripCommas(strings.TrimSpace(m.fields[fieldUnitsPerPurchase].Value())), 64)
		if err != nil || upp <= 0 {
			m.message = "La cantidad por compra debe ser mayor a 0"
			m.isError = true
			return m, nil
		}
		p.UnitType = "measure"
		p.MeasurementUnit = mu
		p.UnitsPerPurchase = upp
	}

	if m.editing != nil {
		p.ID = m.editing.ID
		err = m.store.UpdateProduct(p)
	} else {
		err = m.store.CreateProduct(p)
	}

	if err != nil {
		m.message = "Error: " + err.Error()
		m.isError = true
		return m, nil
	}

	return m, func() tea.Msg {
		return switchScreenMsg{screen: screenProducts}
	}
}

func (m productFormModel) view() string {
	var b strings.Builder

	title := "Nuevo Producto"
	if m.editing != nil {
		title = "Editar Producto"
	}
	b.WriteString(screenTitleStyle.Render(title))
	if m.measured {
		b.WriteString("  " + dimStyle.Render("P.Compra por unidad de compra · P.Venta por unidad de medida"))
	}
	b.WriteString("\n\n")

	if m.message != "" {
		if m.isError {
			b.WriteString("  " + errorStyle.Render(m.message) + "\n\n")
		} else {
			b.WriteString("  " + successStyle.Render(m.message) + "\n\n")
		}
	}

	for i := range m.fields {
		// Hide measurement fields for unit products
		if !m.measured && (i == fieldMeasurementUnit || i == fieldUnitsPerPurchase) {
			continue
		}
		indicator := "  "
		if i == m.focus {
			indicator = "▸ "
		}
		label := fieldLabels[i]
		if i == fieldUnitType {
			val := "[ Unidad ]"
			if m.measured {
				val = "[ Medida ]"
			}
			b.WriteString(fmt.Sprintf("  %s%-16s %s %s\n", indicator, label+":", val,
				dimStyle.Render("(espacio para cambiar)")))
			continue
		}
		b.WriteString(fmt.Sprintf("  %s%-16s %s\n", indicator, label+":", m.fields[i].View()))
	}

	b.WriteString("\n")
	b.WriteString("  " + hKey(hkNav, "tab/↑↓", "navegar campos") + hSep() + hKey(hkOk, "ctrl+s", "guardar") + hSep() + hKey(hkDel, "esc", "cancelar"))
	b.WriteString("\n")
	return b.String()
}
