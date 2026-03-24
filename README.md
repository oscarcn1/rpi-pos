# POS - Punto de Venta

Sistema de punto de venta para pequeños negocios: tiendas de abarrotes, papelerías, ferreterías, tiendas de ropa, y cualquier comercio minorista. Aplicación de terminal (TUI) escrita en Go con base de datos SQLite embebida, diseñada para ejecutarse en hardware de bajo consumo como Raspberry Pi 4.

## Características

- **Venta rápida** con búsqueda inteligente de productos por nombre o código (sin necesidad de acentos)
- **Dos tipos de producto**: por unidad (piezas) y por medida (metros, kilos, litros, etc.)
- **Control de inventario** con descuento automático de existencias al vender
- **Devoluciones** parciales o totales con restauración automática de stock
- **Registro de merma** con razón y cantidades decimales para productos por medida
- **Reportes**: cierre del día, reorden, inventario valorizado, finanzas mensuales con gráfica, devoluciones
- **Interfaz intuitiva** con instrucciones contextuales, colores por tipo de acción y formato de números con separador de miles
- **Un solo binario** de ~11 MB sin dependencias externas, sin servidor web, sin navegador

---

## Arquitectura

```
┌─────────────────────────────────────────────┐
│            Un solo binario Go               │
│                                             │
│  ┌────────────┐    ┌─────────────────────┐  │
│  │  TUI       │───▶│  Lógica de negocio  │  │
│  │ Bubble Tea │◀───│  (store)            │  │
│  └────────────┘    └──────────┬──────────┘  │
│                               │             │
│                    ┌──────────▼──────────┐  │
│                    │  SQLite (embebido)  │  │
│                    │  ~/.pos/pos.db      │  │
│                    └────────────────────-┘  │
└─────────────────────────────────────────────┘
```

### Stack tecnológico

| Componente | Tecnología | Descripción |
|---|---|---|
| **Lenguaje** | Go 1.26.1 | Compilado, binario estático para ARM64 |
| **Interfaz** | [Bubble Tea](https://github.com/charmbracelet/bubbletea) v1.3.5 | Framework TUI con patrón Model-View-Update |
| **Estilos** | [Lip Gloss](https://github.com/charmbracelet/lipgloss) v1.1.0 | Estilos de terminal (colores, bordes, alineación) |
| **Componentes** | [Bubbles](https://github.com/charmbracelet/bubbles) v0.21.0 | Inputs de texto, tablas, listas |
| **Base de datos** | [SQLite](https://modernc.org/sqlite) v1.37.1 | Embebida, pure Go (sin CGO), archivo único |
| **Búsqueda** | Levenshtein + normalización Unicode | Búsqueda fuzzy tolerante a acentos y errores tipográficos |

### Estructura del proyecto

```
pos/
├── cmd/
│   └── pos/
│       └── main.go                  # Punto de entrada
├── internal/
│   ├── database/
│   │   └── database.go              # Conexión SQLite, esquema y migraciones
│   ├── models/
│   │   └── models.go                # Estructuras de datos y helpers de formato
│   ├── search/
│   │   ├── fuzzy.go                 # Algoritmo de búsqueda fuzzy (Levenshtein)
│   │   └── fuzzy_test.go            # Tests de búsqueda
│   ├── store/
│   │   ├── store.go                 # Wrapper de base de datos
│   │   ├── products.go              # CRUD de productos y búsqueda
│   │   ├── sales.go                 # Transacciones de venta
│   │   ├── shrinkage.go             # Registro de merma
│   │   └── reports.go               # Cierre del día, reorden, inventario
│   └── tui/
│       ├── app.go                   # Controlador principal de la app
│       ├── menu.go                  # Menú principal
│       ├── sale.go                  # Pantalla de venta (búsqueda, carrito, cobro)
│       ├── products.go              # Lista de productos
│       ├── product_form.go          # Formulario de alta/edición de productos
│       ├── search.go                # Búsqueda de productos
│       ├── shrinkage.go             # Registro de merma
│       ├── reports.go               # Reportes de cierre y reorden
│       ├── inventory.go             # Reporte de inventario
│       ├── styles.go                # Estilos de terminal y helpers de color
│       └── format.go                # Formato de números con comas
├── scripts/
│   ├── seed.go                      # Script para poblar datos de ejemplo
│   └── seed.sql                     # Datos de ejemplo en SQL
├── go.mod
└── go.sum
```

---

## Base de datos

### Esquema

La base de datos se crea automáticamente en `~/.pos/pos.db` al ejecutar la aplicación por primera vez.

#### Tabla `products`

| Columna | Tipo | Descripción |
|---|---|---|
| `id` | INTEGER PK | Identificador único |
| `code` | TEXT UNIQUE | Código del producto (hasta 100 caracteres) |
| `name` | TEXT | Nombre del producto |
| `description` | TEXT | Descripción |
| `category` | TEXT | Categoría |
| `purchase_price` | REAL | Precio de compra (por unidad de compra) |
| `sale_price` | REAL | Precio de venta (por pieza o por unidad de medida) |
| `stock` | REAL | Existencias actuales |
| `min_stock` | REAL | Stock mínimo para reorden |
| `unit_type` | TEXT | `"unit"` (pieza) o `"measure"` (por medida) |
| `measurement_unit` | TEXT | Unidad de medida: `"metros"`, `"kilos"`, `"litros"`, etc. |
| `units_per_purchase` | REAL | Unidades de medida por compra (ej: 100 metros por rollo) |
| `created_at` | DATETIME | Fecha de creación |
| `updated_at` | DATETIME | Última modificación |

#### Tabla `sales`

| Columna | Tipo | Descripción |
|---|---|---|
| `id` | INTEGER PK | Identificador de la venta |
| `total` | REAL | Total de la venta |
| `payment` | REAL | Monto pagado por el cliente |
| `change_amount` | REAL | Cambio devuelto |
| `created_at` | DATETIME | Fecha y hora de la venta |

#### Tabla `sale_items`

| Columna | Tipo | Descripción |
|---|---|---|
| `id` | INTEGER PK | Identificador |
| `sale_id` | INTEGER FK | Referencia a `sales` |
| `product_id` | INTEGER FK | Referencia a `products` |
| `product_name` | TEXT | Nombre del producto (snapshot) |
| `quantity` | REAL | Cantidad vendida |
| `unit_price` | REAL | Precio unitario al momento de la venta |
| `subtotal` | REAL | Subtotal de la línea |
| `cost_per_unit` | REAL | Costo por unidad al momento de la venta (para cálculo de ganancia) |

#### Tabla `shrinkage`

| Columna | Tipo | Descripción |
|---|---|---|
| `id` | INTEGER PK | Identificador |
| `product_id` | INTEGER FK | Referencia a `products` |
| `quantity` | REAL | Cantidad perdida |
| `reason` | TEXT | Razón de la merma |
| `created_at` | DATETIME | Fecha de registro |

### Sistema de migraciones

La base de datos usa un sistema de migraciones versionadas con la tabla `schema_version`. Al iniciar, la aplicación verifica la versión actual y aplica las migraciones pendientes automáticamente.

### Configuración de SQLite

- **Modo WAL** (Write-Ahead Logging) para mejor rendimiento
- **Busy timeout** de 5 segundos
- **Una sola conexión** (suficiente para un solo usuario)
- **Índices** en `code`, `name`, `created_at` para consultas rápidas

---

## Productos por medida

El sistema soporta dos tipos de productos:

### Producto por unidad (pieza)

Se compra y se vende por pieza. Ejemplo: un sartén.

| Campo | Valor |
|---|---|
| Precio de compra | $85.00 (costo por sartén) |
| Precio de venta | $159.00 (precio al público por sartén) |
| Stock | 15 (piezas disponibles) |

### Producto por medida

Se compra en una presentación (rollo, carrete, bolsa) y se vende por unidad de medida. Ejemplo: soga de nylon.

| Campo | Valor |
|---|---|
| Precio de compra | $350.00 (costo por rollo) |
| Cantidad por compra | 100 metros (metros por rollo) |
| Precio de venta | $5.50 (precio por metro) |
| Stock | 100 metros (metros disponibles) |
| Costo por metro | $3.50 (calculado: $350 / 100) |

Al vender 2.5 metros:
- Se cobra: 2.5 × $5.50 = $13.75
- Se descuenta del stock: 100 - 2.5 = 97.5 metros
- Se registra costo: 2.5 × $3.50 = $8.75 (para cálculo de ganancia)

---

## Búsqueda fuzzy

El sistema usa un algoritmo de búsqueda inteligente que:

1. **Normaliza acentos**: `"sarten"` encuentra `"Sartén antiadherente"`
2. **Busca fragmentos**: `"licua"` encuentra `"Licuadora 3 velocidades"`
3. **Tolera errores**: usa distancia de Levenshtein para aproximar coincidencias
4. **Búsqueda multi-palabra**: `"foco led"` encuentra `"Foco LED 9W luz blanca"`
5. **Código exacto**: si escribes un código exacto, va directo al producto

El algoritmo prioriza: coincidencia exacta > subcadena > prefijo > Levenshtein.

---

## Pantallas de la aplicación

### 1. Menú principal
Acceso a todas las funciones con teclas numéricas o flechas.

### 2. Nueva venta
- Búsqueda de producto arriba, carrito abajo
- Resultados con fondo azul resaltado en la selección
- Carrito con total en tiempo real
- Cobro con cálculo automático de cambio
- Soporte para cantidades decimales en productos por medida

### 3. Productos
Lista completa con precio, stock y tipo. Crear, editar y eliminar productos.

### 4. Formulario de producto
11 campos con visibilidad condicional. Toggle entre "Unidad" y "Medida" con barra espaciadora.

### 5. Registrar merma
Flujo guiado: código → cantidad → razón. Descuenta del inventario automáticamente.

### 6. Buscar producto
Búsqueda fuzzy con tabla de resultados. Resalta productos con stock bajo.

### 7. Cierre del día
Resumen: ventas, ingresos, costos, ganancia, productos más vendidos, merma del día.

### 8. Reporte de reorden
Productos con stock por debajo del mínimo, ordenados por déficit.

### 9. Reporte de inventario
Valorización completa del inventario a costo y a venta, desglosado por categoría.

### Código de colores en controles

| Color | Significado | Ejemplo |
|---|---|---|
| **Verde** | Confirmar / acción positiva | `enter: seleccionar`, `F2: cobrar` |
| **Amarillo** | Acción general | `enter: buscar`, `e: editar` |
| **Azul** | Navegación | `↑↓: navegar`, `esc: menú` |
| **Rojo** | Eliminar / cancelar | `d: eliminar`, `esc: cancelar venta` |

---

## Requerimientos del sistema

### Hardware mínimo

- **CPU**: ARM64 o x86_64
- **RAM**: 512 MB (la aplicación usa ~20-30 MB)
- **Almacenamiento**: 50 MB (binario + base de datos)
- **Terminal**: cualquier emulador de terminal con soporte de colores (256 colores)

### Hardware recomendado

- Raspberry Pi 4 (4 GB RAM) o superior
- MicroSD de 16 GB o más
- Conexión SSH para acceso remoto

### Software requerido

- **Go 1.26.1** o superior (solo para compilar)
- **Sistema operativo**: Linux (ARM64 o x86_64), macOS, o Windows con WSL
- No requiere Node.js, Docker, ni servidor web
- No requiere interfaz gráfica ni navegador

---

## Instalación

### 1. Instalar Go

Descargar desde [go.dev](https://go.dev/dl/) la versión para tu arquitectura.

**Raspberry Pi 4 (ARM64):**

```bash
curl -LO https://go.dev/dl/go1.26.1.linux-arm64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.1.linux-arm64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
go version
```

**Linux x86_64:**

```bash
curl -LO https://go.dev/dl/go1.26.1.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.1.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
```

### 2. Clonar el repositorio

```bash
git clone <url-del-repositorio> pos
cd pos
```

### 3. Compilar

```bash
go build -o pos ./cmd/pos/
```

El binario resultante es `./pos` (~11 MB).

Para compilar una versión más pequeña sin símbolos de debug:

```bash
go build -ldflags="-s -w" -o pos ./cmd/pos/
```

### 4. Poblar datos de ejemplo (opcional)

```bash
go run scripts/seed.go
```

Esto inserta 58 productos de ejemplo en 10 categorías (48 por unidad, 10 por medida).

### 5. Ejecutar

```bash
./pos
```

La base de datos se crea automáticamente en `~/.pos/pos.db`.

---

## Compilación cruzada

Para compilar desde otra máquina hacia Raspberry Pi:

```bash
GOOS=linux GOARCH=arm64 go build -o pos ./cmd/pos/
```

Luego copiar el binario al Pi:

```bash
scp pos usuario@ip-del-pi:~/pos
```

---

## Ejecución como servicio systemd

Para que la aplicación arranque automáticamente (útil si se conecta a una terminal dedicada):

```bash
sudo tee /etc/systemd/system/pos.service > /dev/null << 'EOF'
[Unit]
Description=Punto de Venta
After=multi-user.target

[Service]
Type=simple
User=oscar
WorkingDirectory=/home/oscar
ExecStart=/home/oscar/pos
StandardInput=tty
TTYPath=/dev/tty1
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable pos
sudo systemctl start pos
```

Para acceder por SSH, en lugar de servicio, simplemente ejecutar `./pos` en la sesión SSH.

---

## Respaldo de la base de datos

La base de datos es un solo archivo en `~/.pos/pos.db`. Para respaldarlo:

```bash
# Respaldo simple (con la app detenida)
cp ~/.pos/pos.db ~/.pos/pos.db.backup

# Respaldo seguro con SQLite (con la app corriendo)
sqlite3 ~/.pos/pos.db ".backup ~/.pos/pos-$(date +%Y%m%d).db"
```

---

## Estructura de la base de datos (consultas útiles)

```bash
# Ver todos los productos
sqlite3 ~/.pos/pos.db "SELECT code, name, sale_price, stock, unit_type FROM products ORDER BY name;"

# Ver ventas del día
sqlite3 ~/.pos/pos.db "SELECT * FROM sales WHERE DATE(created_at) = DATE('now', 'localtime');"

# Ver productos con stock bajo
sqlite3 ~/.pos/pos.db "SELECT code, name, stock, min_stock FROM products WHERE stock <= min_stock;"

# Total de ventas del día
sqlite3 ~/.pos/pos.db "SELECT COUNT(*), SUM(total) FROM sales WHERE DATE(created_at) = DATE('now', 'localtime');"
```

---

## Licencia

Uso privado.
