package store

import (
	"fmt"
	"pos/internal/models"
	"pos/internal/search"
	"strings"
)

const productColumns = `id, code, name, description, category, purchase_price, sale_price, stock, min_stock, unit_type, measurement_unit, units_per_purchase, created_at, updated_at`

func (s *Store) CreateProduct(p *models.Product) error {
	if p.UnitType == "" {
		p.UnitType = "unit"
	}
	if p.UnitsPerPurchase == 0 {
		p.UnitsPerPurchase = 1.0
	}
	res, err := s.db.Exec(
		`INSERT INTO products (code, name, description, category, purchase_price, sale_price, stock, min_stock, unit_type, measurement_unit, units_per_purchase)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Code, p.Name, p.Description, p.Category, p.PurchasePrice, p.SalePrice, p.Stock, p.MinStock,
		p.UnitType, p.MeasurementUnit, p.UnitsPerPurchase,
	)
	if err != nil {
		return fmt.Errorf("crear producto: %w", err)
	}
	p.ID, _ = res.LastInsertId()
	return nil
}

func (s *Store) UpdateProduct(p *models.Product) error {
	_, err := s.db.Exec(
		`UPDATE products SET code=?, name=?, description=?, category=?,
		 purchase_price=?, sale_price=?, stock=?, min_stock=?,
		 unit_type=?, measurement_unit=?, units_per_purchase=?,
		 updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		p.Code, p.Name, p.Description, p.Category, p.PurchasePrice, p.SalePrice, p.Stock, p.MinStock,
		p.UnitType, p.MeasurementUnit, p.UnitsPerPurchase, p.ID,
	)
	if err != nil {
		return fmt.Errorf("actualizar producto: %w", err)
	}
	return nil
}

func (s *Store) DeleteProduct(id int64) error {
	_, err := s.db.Exec(`DELETE FROM products WHERE id=?`, id)
	return err
}

func (s *Store) GetProductByCode(code string) (*models.Product, error) {
	p := &models.Product{}
	err := s.db.QueryRow(
		`SELECT `+productColumns+` FROM products WHERE code=?`, code,
	).Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.Category, &p.PurchasePrice, &p.SalePrice,
		&p.Stock, &p.MinStock, &p.UnitType, &p.MeasurementUnit, &p.UnitsPerPurchase, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) GetProductByID(id int64) (*models.Product, error) {
	p := &models.Product{}
	err := s.db.QueryRow(
		`SELECT `+productColumns+` FROM products WHERE id=?`, id,
	).Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.Category, &p.PurchasePrice, &p.SalePrice,
		&p.Stock, &p.MinStock, &p.UnitType, &p.MeasurementUnit, &p.UnitsPerPurchase, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) SearchProducts(query string) ([]models.Product, error) {
	q := "%" + strings.ToLower(query) + "%"
	rows, err := s.db.Query(
		`SELECT `+productColumns+` FROM products WHERE LOWER(code) LIKE ? OR LOWER(name) LIKE ? OR LOWER(category) LIKE ?
		 ORDER BY name LIMIT 50`, q, q, q,
	)
	if err != nil {
		return nil, fmt.Errorf("buscar productos: %w", err)
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (s *Store) ListProducts() ([]models.Product, error) {
	rows, err := s.db.Query(
		`SELECT ` + productColumns + ` FROM products ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("listar productos: %w", err)
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (s *Store) FuzzySearchProducts(query string) ([]models.Product, error) {
	p, err := s.GetProductByCode(strings.ToUpper(strings.TrimSpace(query)))
	if err == nil {
		return []models.Product{*p}, nil
	}

	all, err := s.ListProducts()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(all))
	for i, p := range all {
		names[i] = p.Name
	}

	matches := search.Rank(query, names)
	results := make([]models.Product, 0, len(matches))
	for _, m := range matches {
		results = append(results, all[m.Index])
	}
	return results, nil
}

func scanProducts(rows interface {
	Next() bool
	Scan(dest ...any) error
}) ([]models.Product, error) {
	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.Category,
			&p.PurchasePrice, &p.SalePrice, &p.Stock, &p.MinStock,
			&p.UnitType, &p.MeasurementUnit, &p.UnitsPerPurchase,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, nil
}
