package store

import (
	"fmt"
	"pos/internal/models"
	"time"
)

var timeNow = time.Now

func timeDate(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
}

func (s *Store) DayCloseReport() (*models.DayReport, error) {
	r := &models.DayReport{}

	s.db.QueryRow(`SELECT DATE('now', 'localtime')`).Scan(&r.Date)

	s.db.QueryRow(
		`SELECT COALESCE(COUNT(*), 0), COALESCE(SUM(total), 0)
		 FROM sales WHERE DATE(created_at) = DATE('now', 'localtime')`,
	).Scan(&r.TotalSales, &r.TotalIncome)

	// Use cost_per_unit snapshot from sale_items for accurate cost
	s.db.QueryRow(
		`SELECT COALESCE(SUM(si.quantity * si.cost_per_unit), 0)
		 FROM sale_items si
		 JOIN sales s ON si.sale_id = s.id
		 WHERE DATE(s.created_at) = DATE('now', 'localtime')`,
	).Scan(&r.TotalCost)

	r.Profit = r.TotalIncome - r.TotalCost

	rows, err := s.db.Query(
		`SELECT si.product_name, SUM(si.quantity), SUM(si.subtotal)
		 FROM sale_items si
		 JOIN sales s ON si.sale_id = s.id
		 WHERE DATE(s.created_at) = DATE('now', 'localtime')
		 GROUP BY si.product_name
		 ORDER BY SUM(si.subtotal) DESC
		 LIMIT 20`,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ps models.ProductSales
			rows.Scan(&ps.ProductName, &ps.Quantity, &ps.Total)
			r.TopProducts = append(r.TopProducts, ps)
		}
	}

	s.db.QueryRow(
		`SELECT COALESCE(SUM(quantity), 0) FROM shrinkage
		 WHERE DATE(created_at) = DATE('now', 'localtime')`,
	).Scan(&r.TotalShrinkage)

	return r, nil
}

func (s *Store) InventoryReport() (*models.InventoryReport, error) {
	r := &models.InventoryReport{}

	rows, err := s.db.Query(
		`SELECT id, code, name, description, category, purchase_price, sale_price, stock, min_stock,
		 unit_type, measurement_unit, units_per_purchase, created_at, updated_at
		 FROM products ORDER BY category, name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	catMap := make(map[string]*models.CategorySummary)

	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.Category,
			&p.PurchasePrice, &p.SalePrice, &p.Stock, &p.MinStock,
			&p.UnitType, &p.MeasurementUnit, &p.UnitsPerPurchase,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		costVal := p.Stock * p.CostPerUnit()
		saleVal := p.Stock * p.SalePrice

		r.Items = append(r.Items, models.InventoryItem{Product: p, CostValue: costVal, SaleValue: saleVal})
		r.TotalProducts++
		r.TotalUnits += p.Stock
		r.TotalCostValue += costVal
		r.TotalSaleValue += saleVal

		cat := p.Category
		if cat == "" {
			cat = "Sin categoría"
		}
		if cs, ok := catMap[cat]; ok {
			cs.Products++
			cs.Units += p.Stock
			cs.CostValue += costVal
			cs.SaleValue += saleVal
		} else {
			catMap[cat] = &models.CategorySummary{
				Category: cat, Products: 1, Units: p.Stock, CostValue: costVal, SaleValue: saleVal,
			}
		}
	}

	for _, cs := range catMap {
		r.Categories = append(r.Categories, *cs)
	}
	r.PotentialProfit = r.TotalSaleValue - r.TotalCostValue

	return r, nil
}

func (s *Store) MonthlyFinanceReport(year, month int) (*models.MonthlyFinanceReport, error) {
	start := fmt.Sprintf("%04d-%02d-01 00:00:00", year, month)

	// Next month start
	ny, nm := year, month+1
	if nm > 12 {
		nm = 1
		ny++
	}
	end := fmt.Sprintf("%04d-%02d-01 00:00:00", ny, nm)

	// Previous month bounds
	py, pm := year, month-1
	if pm < 1 {
		pm = 12
		py--
	}
	prevStart := fmt.Sprintf("%04d-%02d-01 00:00:00", py, pm)

	// Days in month
	daysInMonth := daysIn(year, month)

	r := &models.MonthlyFinanceReport{
		Year:        year,
		Month:       month,
		DaysInMonth: daysInMonth,
	}

	// Daily sales
	r.DailySales = make([]models.DailySales, daysInMonth)
	for i := range r.DailySales {
		r.DailySales[i].Day = i + 1
	}

	rows, err := s.db.Query(
		`SELECT CAST(strftime('%d', created_at) AS INTEGER) AS day, COUNT(*), COALESCE(SUM(total), 0)
		 FROM sales WHERE created_at >= ? AND created_at < ?
		 GROUP BY day ORDER BY day`, start, end,
	)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var day, count int
			var total float64
			rows.Scan(&day, &count, &total)
			if day >= 1 && day <= daysInMonth {
				r.DailySales[day-1] = models.DailySales{Day: day, Count: count, Total: total}
				r.TotalSales += count
				r.TotalIncome += total
			}
		}
	}

	// Total cost
	s.db.QueryRow(
		`SELECT COALESCE(SUM(si.quantity * si.cost_per_unit), 0)
		 FROM sale_items si JOIN sales s ON si.sale_id = s.id
		 WHERE s.created_at >= ? AND s.created_at < ?`, start, end,
	).Scan(&r.TotalCost)

	r.Profit = r.TotalIncome - r.TotalCost

	// Average daily (days elapsed so far)
	elapsed := daysInMonth
	now := timeNow()
	if year == now.Year() && month == int(now.Month()) {
		elapsed = now.Day()
	}
	if elapsed > 0 {
		r.AvgDailySales = r.TotalIncome / float64(elapsed)
	}

	// Shrinkage
	s.db.QueryRow(
		`SELECT COALESCE(SUM(quantity), 0) FROM shrinkage
		 WHERE created_at >= ? AND created_at < ?`, start, end,
	).Scan(&r.TotalShrinkage)

	// Previous month
	s.db.QueryRow(
		`SELECT COALESCE(COUNT(*), 0), COALESCE(SUM(total), 0)
		 FROM sales WHERE created_at >= ? AND created_at < ?`, prevStart, start,
	).Scan(&r.PrevSales, &r.PrevIncome)

	s.db.QueryRow(
		`SELECT COALESCE(SUM(si.quantity * si.cost_per_unit), 0)
		 FROM sale_items si JOIN sales s ON si.sale_id = s.id
		 WHERE s.created_at >= ? AND s.created_at < ?`, prevStart, start,
	).Scan(&r.PrevCost)

	r.PrevProfit = r.PrevIncome - r.PrevCost
	r.HasPrev = r.PrevSales > 0

	return r, nil
}

func daysIn(year, month int) int {
	// First day of next month, minus one day
	ny, nm := year, month+1
	if nm > 12 {
		nm = 1
		ny++
	}
	t := timeDate(ny, nm, 1).AddDate(0, 0, -1)
	return t.Day()
}

func (s *Store) ReorderReport() ([]models.ReorderItem, error) {
	rows, err := s.db.Query(
		`SELECT id, code, name, description, category, purchase_price, sale_price, stock, min_stock,
		 unit_type, measurement_unit, units_per_purchase, created_at, updated_at
		 FROM products WHERE stock <= min_stock ORDER BY (min_stock - stock) DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ReorderItem
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.Category,
			&p.PurchasePrice, &p.SalePrice, &p.Stock, &p.MinStock,
			&p.UnitType, &p.MeasurementUnit, &p.UnitsPerPurchase,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, models.ReorderItem{
			Product: p,
			Deficit: p.MinStock - p.Stock,
		})
	}
	return items, nil
}
