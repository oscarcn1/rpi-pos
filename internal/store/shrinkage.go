package store

import (
	"fmt"
	"pos/internal/models"
)

func (s *Store) CreateShrinkage(sh *models.Shrinkage) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("iniciar transacción: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO shrinkage (product_id, quantity, reason) VALUES (?, ?, ?)`,
		sh.ProductID, sh.Quantity, sh.Reason,
	)
	if err != nil {
		return fmt.Errorf("registrar merma: %w", err)
	}
	sh.ID, _ = res.LastInsertId()

	_, err = tx.Exec(`UPDATE products SET stock = stock - ? WHERE id = ?`, sh.Quantity, sh.ProductID)
	if err != nil {
		return fmt.Errorf("actualizar stock: %w", err)
	}

	return tx.Commit()
}
