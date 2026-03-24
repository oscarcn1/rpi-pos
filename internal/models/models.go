package models

import (
	"fmt"
	"strings"
	"time"
)

type Product struct {
	ID               int64
	Code             string
	Name             string
	Description      string
	Category         string
	PurchasePrice    float64
	SalePrice        float64
	Stock            float64
	MinStock         float64
	UnitType         string  // "unit" or "measure"
	MeasurementUnit  string  // "metros", "kilos", "litros", etc.
	UnitsPerPurchase float64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (p Product) IsMeasured() bool { return p.UnitType == "measure" }

func (p Product) CostPerUnit() float64 {
	if p.IsMeasured() && p.UnitsPerPurchase > 0 {
		return p.PurchasePrice / p.UnitsPerPurchase
	}
	return p.PurchasePrice
}

func (p Product) StockLabel() string {
	if p.IsMeasured() {
		return FmtQty(p.Stock) + " " + p.MeasurementUnit
	}
	return FmtQty(p.Stock)
}

func (p Product) MinStockLabel() string {
	if p.IsMeasured() {
		return FmtQty(p.MinStock) + " " + p.MeasurementUnit
	}
	return FmtQty(p.MinStock)
}

func (p Product) UnitLabel() string {
	if p.IsMeasured() {
		return p.MeasurementUnit
	}
	return "pza"
}

func (p Product) FormatQty(qty float64) string {
	if p.IsMeasured() {
		return FmtQty(qty) + " " + p.MeasurementUnit
	}
	return FmtQty(qty)
}

// FmtQty formats a quantity with commas: whole numbers without decimals, otherwise 2 decimals.
func FmtQty(n float64) string {
	if n == float64(int64(n)) {
		return addCommas(fmt.Sprintf("%.0f", n))
	}
	s := fmt.Sprintf("%.2f", n)
	parts := strings.Split(s, ".")
	return addCommas(parts[0]) + "." + parts[1]
}

// FmtPrice formats a price with commas and $ sign.
func FmtPrice(n float64) string {
	return "$" + fmtDecimal(n)
}

func fmtDecimal(n float64) string {
	neg := ""
	if n < 0 {
		neg = "-"
		n = -n
	}
	s := fmt.Sprintf("%.2f", n)
	parts := strings.Split(s, ".")
	return neg + addCommas(parts[0]) + "." + parts[1]
}

func addCommas(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (n-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

type Sale struct {
	ID           int64
	Items        []SaleItem
	Total        float64
	Payment      float64
	ChangeAmount float64
	CreatedAt    time.Time
}

type SaleItem struct {
	ID          int64
	SaleID      int64
	ProductID   int64
	ProductName string
	Quantity    float64
	UnitPrice   float64
	Subtotal    float64
	CostPerUnit float64
}

type Shrinkage struct {
	ID          int64
	ProductID   int64
	ProductName string
	Quantity    float64
	Reason      string
	CreatedAt   time.Time
}

type DayReport struct {
	Date           string
	TotalSales     int
	TotalIncome    float64
	TotalCost      float64
	Profit         float64
	TopProducts    []ProductSales
	TotalShrinkage float64
}

type ProductSales struct {
	ProductName string
	Quantity    float64
	Total       float64
}

type ReorderItem struct {
	Product Product
	Deficit float64
}

type InventoryReport struct {
	TotalProducts   int
	TotalUnits      float64
	TotalCostValue  float64
	TotalSaleValue  float64
	PotentialProfit float64
	Categories      []CategorySummary
	Items           []InventoryItem
}

type CategorySummary struct {
	Category  string
	Products  int
	Units     float64
	CostValue float64
	SaleValue float64
}

type InventoryItem struct {
	Product   Product
	CostValue float64
	SaleValue float64
}
