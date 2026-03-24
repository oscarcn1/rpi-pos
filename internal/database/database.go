package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("crear directorio: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("abrir base de datos: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migración: %w", err)
	}

	return db, nil
}

func migrate(db *sql.DB) error {
	// Base schema (runs only on fresh databases)
	schema := `
	CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		description TEXT DEFAULT '',
		category TEXT DEFAULT '',
		purchase_price REAL NOT NULL,
		sale_price REAL NOT NULL,
		stock REAL NOT NULL DEFAULT 0,
		min_stock REAL NOT NULL DEFAULT 0,
		unit_type TEXT NOT NULL DEFAULT 'unit',
		measurement_unit TEXT NOT NULL DEFAULT '',
		units_per_purchase REAL NOT NULL DEFAULT 1.0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sales (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		total REAL NOT NULL,
		payment REAL NOT NULL,
		change_amount REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sale_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sale_id INTEGER NOT NULL REFERENCES sales(id),
		product_id INTEGER NOT NULL REFERENCES products(id),
		product_name TEXT NOT NULL,
		quantity REAL NOT NULL,
		unit_price REAL NOT NULL,
		subtotal REAL NOT NULL,
		cost_per_unit REAL NOT NULL DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS shrinkage (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		product_id INTEGER NOT NULL REFERENCES products(id),
		quantity REAL NOT NULL,
		reason TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS schema_version (version INTEGER NOT NULL);

	CREATE INDEX IF NOT EXISTS idx_products_code ON products(code);
	CREATE INDEX IF NOT EXISTS idx_products_name ON products(name);
	CREATE INDEX IF NOT EXISTS idx_sales_created ON sales(created_at);
	CREATE INDEX IF NOT EXISTS idx_shrinkage_created ON shrinkage(created_at);
	`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Versioned migrations for existing databases
	version := 0
	_ = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)

	if version < 2 {
		if err := migrateV2(db); err != nil {
			return fmt.Errorf("migración v2: %w", err)
		}
	}

	return nil
}

func migrateV2(db *sql.DB) error {
	// Add new columns to existing tables (safe: ALTER TABLE ADD COLUMN is idempotent-ish)
	alters := []string{
		"ALTER TABLE products ADD COLUMN unit_type TEXT NOT NULL DEFAULT 'unit'",
		"ALTER TABLE products ADD COLUMN measurement_unit TEXT NOT NULL DEFAULT ''",
		"ALTER TABLE products ADD COLUMN units_per_purchase REAL NOT NULL DEFAULT 1.0",
		"ALTER TABLE sale_items ADD COLUMN cost_per_unit REAL NOT NULL DEFAULT 0",
	}

	for _, stmt := range alters {
		// Ignore "duplicate column" errors from already-migrated databases
		_, _ = db.Exec(stmt)
	}

	_, err := db.Exec("INSERT INTO schema_version (version) VALUES (2)")
	return err
}
