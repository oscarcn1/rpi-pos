package store

import (
	"fmt"
	"pos/internal/models"
)

func (s *Store) ListSalesDesc(limit int) ([]models.Sale, error) {
	rows, err := s.db.Query(
		`SELECT id, total, payment, change_amount, created_at
		 FROM sales ORDER BY created_at DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []models.Sale
	for rows.Next() {
		var sale models.Sale
		if err := rows.Scan(&sale.ID, &sale.Total, &sale.Payment, &sale.ChangeAmount, &sale.CreatedAt); err != nil {
			return nil, err
		}
		sales = append(sales, sale)
	}
	return sales, nil
}

func (s *Store) GetSaleDetail(saleID int64) (*models.SaleDetail, error) {
	detail := &models.SaleDetail{}

	err := s.db.QueryRow(
		`SELECT id, total, payment, change_amount, created_at FROM sales WHERE id=?`, saleID,
	).Scan(&detail.Sale.ID, &detail.Sale.Total, &detail.Sale.Payment, &detail.Sale.ChangeAmount, &detail.Sale.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(
		`SELECT id, sale_id, product_id, product_name, quantity, unit_price, subtotal, cost_per_unit
		 FROM sale_items WHERE sale_id=?`, saleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.SaleItem
		if err := rows.Scan(&item.ID, &item.SaleID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.UnitPrice, &item.Subtotal, &item.CostPerUnit); err != nil {
			return nil, err
		}
		detail.Items = append(detail.Items, item)
	}
	return detail, nil
}

// ReturnedQtyForSaleItem returns how much of a sale item has already been returned.
func (s *Store) ReturnedQtyForSaleItem(saleItemID int64) float64 {
	var qty float64
	s.db.QueryRow(
		`SELECT COALESCE(SUM(quantity), 0) FROM return_items WHERE sale_item_id=?`, saleItemID,
	).Scan(&qty)
	return qty
}

func (s *Store) CreateReturn(ret *models.Return) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("iniciar transacción: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO returns (sale_id, total, reason) VALUES (?, ?, ?)`,
		ret.SaleID, ret.Total, ret.Reason,
	)
	if err != nil {
		return fmt.Errorf("crear devolución: %w", err)
	}
	ret.ID, _ = res.LastInsertId()

	for i := range ret.Items {
		item := &ret.Items[i]
		item.ReturnID = ret.ID
		_, err := tx.Exec(
			`INSERT INTO return_items (return_id, sale_item_id, product_id, product_name, quantity, unit_price, subtotal)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			item.ReturnID, item.SaleItemID, item.ProductID, item.ProductName,
			item.Quantity, item.UnitPrice, item.Subtotal,
		)
		if err != nil {
			return fmt.Errorf("crear item de devolución: %w", err)
		}

		// Restore stock
		_, err = tx.Exec(`UPDATE products SET stock = stock + ? WHERE id = ?`, item.Quantity, item.ProductID)
		if err != nil {
			return fmt.Errorf("restaurar stock: %w", err)
		}
	}

	return tx.Commit()
}

func (s *Store) DayReturnsReport(date string) (*models.DayReturnsReport, error) {
	r := &models.DayReturnsReport{}
	r.Date = date

	s.db.QueryRow(
		`SELECT COALESCE(COUNT(*), 0), COALESCE(SUM(total), 0)
		 FROM returns WHERE DATE(created_at) = ?`, date,
	).Scan(&r.TotalReturns, &r.TotalAmount)

	rows, err := s.db.Query(
		`SELECT r.id, r.sale_id, r.total, r.reason,
		 (SELECT COUNT(*) FROM return_items WHERE return_id = r.id), r.created_at
		 FROM returns r WHERE DATE(r.created_at) = ?
		 ORDER BY r.created_at DESC`, date,
	)
	if err != nil {
		return r, nil
	}
	defer rows.Close()

	for rows.Next() {
		var rs models.ReturnSummary
		rows.Scan(&rs.ID, &rs.SaleID, &rs.Total, &rs.Reason, &rs.ItemCount, &rs.CreatedAt)
		r.Returns = append(r.Returns, rs)
	}

	return r, nil
}
