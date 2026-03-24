package store

import (
	"fmt"
	"pos/internal/models"
)

func (s *Store) CreateSale(sale *models.Sale) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("iniciar transacción: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO sales (total, payment, change_amount) VALUES (?, ?, ?)`,
		sale.Total, sale.Payment, sale.ChangeAmount,
	)
	if err != nil {
		return fmt.Errorf("crear venta: %w", err)
	}
	sale.ID, _ = res.LastInsertId()

	for i := range sale.Items {
		item := &sale.Items[i]
		item.SaleID = sale.ID
		_, err := tx.Exec(
			`INSERT INTO sale_items (sale_id, product_id, product_name, quantity, unit_price, subtotal, cost_per_unit)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			item.SaleID, item.ProductID, item.ProductName, item.Quantity, item.UnitPrice, item.Subtotal, item.CostPerUnit,
		)
		if err != nil {
			return fmt.Errorf("crear item de venta: %w", err)
		}

		_, err = tx.Exec(`UPDATE products SET stock = stock - ? WHERE id = ?`, item.Quantity, item.ProductID)
		if err != nil {
			return fmt.Errorf("actualizar stock: %w", err)
		}
	}

	return tx.Commit()
}

func (s *Store) GetTodaySales() ([]models.Sale, error) {
	rows, err := s.db.Query(
		`SELECT id, total, payment, change_amount, created_at
		 FROM sales WHERE DATE(created_at) = DATE('now', 'localtime')
		 ORDER BY created_at DESC`,
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
