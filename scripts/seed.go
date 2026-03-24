//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type product struct {
	code, name, desc, cat        string
	pp, sp                       float64
	stock, min                   float64
	unitType, measUnit           string
	unitsPerPurchase             float64
}

func u(code, name, desc, cat string, pp, sp float64, stock, min int) product {
	return product{code, name, desc, cat, pp, sp, float64(stock), float64(min), "unit", "", 1}
}

func m(code, name, desc, cat string, pp, sp, stock, min, upp float64, unit string) product {
	return product{code, name, desc, cat, pp, sp, stock, min, "measure", unit, upp}
}

func main() {
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".pos", "pos.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	products := []product{
		// Cocina
		u("COC001", "Sartén antiadherente 26cm", "Sartén con recubrimiento antiadherente", "Cocina", 85, 159, 15, 5),
		u("COC002", "Olla de presión 6L", "Olla de presión de aluminio 6 litros", "Cocina", 320, 589, 8, 3),
		u("COC003", "Juego de cuchillos x5", "Set de 5 cuchillos de acero inoxidable", "Cocina", 120, 229, 12, 4),
		u("COC004", "Licuadora 3 velocidades", "Licuadora de vaso de vidrio 1.5L", "Cocina", 280, 499, 6, 3),
		u("COC005", "Batería de cocina 9 pzas", "Juego de ollas y sartenes aluminio", "Cocina", 450, 849, 5, 2),
		u("COC006", "Tostador de pan 2 rebanadas", "Tostador eléctrico con control", "Cocina", 150, 279, 10, 3),
		u("COC007", "Tabla para picar bambú", "Tabla de cortar de bambú 30x40cm", "Cocina", 45, 89, 20, 8),
		u("COC008", "Colador de acero inoxidable", "Colador de malla fina 20cm", "Cocina", 35, 69, 18, 6),
		u("COC009", "Molde para pastel redondo", "Molde antiadherente 24cm", "Cocina", 55, 99, 14, 5),
		u("COC010", "Rallador 4 caras", "Rallador de acero inoxidable", "Cocina", 28, 55, 22, 8),
		// Limpieza
		u("LIM001", "Escoba de plástico", "Escoba resistente con mango largo", "Limpieza", 35, 65, 25, 10),
		u("LIM002", "Trapeador de microfibra", "Trapeador con mango telescópico", "Limpieza", 55, 99, 18, 8),
		u("LIM003", "Cubeta 12 litros", "Cubeta plástica resistente", "Limpieza", 22, 45, 30, 12),
		u("LIM004", "Recogedor con mango", "Recogedor de basura con mango largo", "Limpieza", 28, 55, 20, 8),
		u("LIM005", "Cepillo para baño", "Cepillo con base de plástico", "Limpieza", 25, 49, 15, 6),
		u("LIM006", "Guantes de hule par", "Guantes para limpieza talla mediana", "Limpieza", 12, 25, 40, 15),
		u("LIM007", "Jalador para vidrios", "Jalador de hule 30cm", "Limpieza", 30, 59, 12, 5),
		u("LIM008", "Bote de basura 20L", "Bote con tapa de balancín", "Limpieza", 65, 119, 10, 4),
		// Baño
		u("BAN001", "Cortina para baño", "Cortina de poliéster 180x180cm", "Baño", 75, 139, 12, 4),
		u("BAN002", "Tapete antiderrapante", "Tapete para baño con ventosas", "Baño", 45, 85, 15, 5),
		u("BAN003", "Organizador de ducha", "Organizador esquinero 3 niveles", "Baño", 90, 169, 8, 3),
		u("BAN004", "Porta cepillos de dientes", "Porta cepillos de cerámica", "Baño", 35, 69, 20, 6),
		u("BAN005", "Jabonera de cerámica", "Jabonera decorativa", "Baño", 30, 59, 18, 6),
		u("BAN006", "Toallero de barra 60cm", "Toallero de acero cromado", "Baño", 55, 99, 10, 4),
		// Organización
		u("ORG001", "Caja organizadora grande", "Caja plástica con tapa 50L", "Organización", 85, 159, 12, 4),
		u("ORG002", "Caja organizadora mediana", "Caja plástica con tapa 30L", "Organización", 55, 99, 15, 5),
		u("ORG003", "Ganchos para ropa x10", "Ganchos de plástico resistente", "Organización", 18, 35, 50, 20),
		u("ORG004", "Zapatera de 10 niveles", "Zapatera metálica desmontable", "Organización", 180, 339, 6, 2),
		u("ORG005", "Canasta de mimbre mediana", "Canasta decorativa multiusos", "Organización", 65, 119, 10, 4),
		u("ORG006", "Perchero de pie", "Perchero metálico 8 ganchos", "Organización", 150, 279, 5, 2),
		// Herramientas
		u("HER001", "Martillo de uña", "Martillo de acero con mango de fibra", "Herramientas", 45, 85, 12, 5),
		u("HER002", "Desarmador juego x6", "Set de desarmadores plano y cruz", "Herramientas", 55, 99, 10, 4),
		u("HER003", "Pinzas de presión", "Pinzas de presión 10 pulgadas", "Herramientas", 40, 75, 15, 5),
		u("HER004", "Cinta métrica 5m", "Flexómetro de 5 metros", "Herramientas", 25, 49, 20, 8),
		u("HER005", "Llave ajustable 10\"", "Llave perica de acero", "Herramientas", 65, 119, 8, 3),
		u("HER006", "Cinta de aislar negra", "Cinta aislante 18m", "Herramientas", 8, 18, 35, 15),
		// Iluminación
		u("ILU001", "Foco LED 9W luz blanca", "Foco ahorrador equivalente 60W", "Iluminación", 18, 35, 40, 15),
		u("ILU002", "Foco LED 9W luz cálida", "Foco ahorrador luz cálida", "Iluminación", 18, 35, 35, 15),
		u("ILU003", "Lámpara de escritorio", "Lámpara LED flexible con base", "Iluminación", 120, 229, 7, 3),
		u("ILU004", "Extensión eléctrica 3m", "Extensión con 3 contactos", "Iluminación", 35, 65, 20, 8),
		u("ILU005", "Multicontacto 6 entradas", "Multicontacto con supresor de picos", "Iluminación", 85, 159, 12, 5),
		u("ILU006", "Linterna LED recargable", "Linterna de mano 500 lumens", "Iluminación", 65, 119, 10, 4),
		// Textiles
		u("TEX001", "Juego de sábanas matrimonial", "Sábanas de microfibra 3 piezas", "Textiles", 180, 339, 8, 3),
		u("TEX002", "Almohada estándar", "Almohada de fibra hipoalergénica", "Textiles", 55, 99, 15, 5),
		u("TEX003", "Toalla de baño grande", "Toalla de algodón 70x140cm", "Textiles", 65, 119, 20, 8),
		u("TEX004", "Mantel rectangular", "Mantel de poliéster 150x250cm", "Textiles", 75, 139, 10, 4),
		u("TEX005", "Cobertor matrimonial", "Cobertor de fleece suave", "Textiles", 220, 399, 6, 2),
		u("TEX006", "Cojín decorativo 45x45", "Cojín con funda lavable", "Textiles", 45, 85, 18, 6),

		// === Productos por MEDIDA ===
		// Cuerdas y cables
		m("SOG001", "Soga de nylon 6mm", "Carrete de soga de nylon trenzada", "Cuerdas", 350, 5.50, 100, 20, 100, "metros"),
		m("SOG002", "Soga de algodón 10mm", "Carrete de soga de algodón", "Cuerdas", 280, 4.80, 50, 15, 50, "metros"),
		m("SOG003", "Cable eléctrico 12 AWG", "Cable THW calibre 12", "Cuerdas", 890, 12.00, 100, 30, 100, "metros"),
		m("SOG004", "Cadena galvanizada 3mm", "Cadena de eslabón corto galvanizada", "Cuerdas", 520, 8.50, 30, 10, 30, "metros"),
		m("SOG005", "Manguera para jardín 1/2\"", "Manguera flexible reforzada", "Cuerdas", 450, 7.00, 50, 15, 50, "metros"),
		// Telas y cintas
		m("TEL001", "Tela para cortina lisa", "Tela de poliéster para cortinas", "Telas", 180, 45.00, 25, 8, 25, "metros"),
		m("TEL002", "Malla mosquitera", "Malla de fibra de vidrio anti insectos", "Telas", 120, 18.00, 30, 10, 30, "metros"),
		m("TEL003", "Cinta de aislar jumbo", "Cinta aislante rollo industrial 50m", "Telas", 85, 2.50, 50, 15, 50, "metros"),
		// Materiales a granel
		m("GRA001", "Alambre galvanizado cal.16", "Rollo de alambre galvanizado", "Ferretería", 95, 3.50, 50, 15, 50, "metros"),
		m("GRA002", "Tubo termoencogible 6mm", "Rollo de termoencogible negro", "Ferretería", 65, 4.00, 20, 5, 20, "metros"),
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO products
		(code, name, description, category, purchase_price, sale_price, stock, min_stock,
		 unit_type, measurement_unit, units_per_purchase)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer stmt.Close()

	count := 0
	for _, p := range products {
		ut := p.unitType
		if ut == "" {
			ut = "unit"
		}
		upp := p.unitsPerPurchase
		if upp == 0 {
			upp = 1
		}
		res, err := stmt.Exec(p.code, p.name, p.desc, p.cat, p.pp, p.sp, p.stock, p.min,
			ut, p.measUnit, upp)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error insertando %s: %v\n", p.code, err)
			continue
		}
		n, _ := res.RowsAffected()
		count += int(n)
	}

	if err := tx.Commit(); err != nil {
		fmt.Fprintf(os.Stderr, "Error commit: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%d productos nuevos insertados\n", count)

	var total, measured int
	db.QueryRow("SELECT COUNT(*) FROM products").Scan(&total)
	db.QueryRow("SELECT COUNT(*) FROM products WHERE unit_type='measure'").Scan(&measured)
	fmt.Printf("Total: %d productos (%d por unidad, %d por medida)\n", total, total-measured, measured)
}
