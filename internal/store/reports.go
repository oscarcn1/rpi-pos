package store

import (
	"pos/internal/models"
)

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
