-- Productos de ejemplo: Tienda de Enseres Domésticos
-- Categoría: Cocina
INSERT OR IGNORE INTO products (code, name, description, category, purchase_price, sale_price, stock, min_stock) VALUES
('COC001', 'Sartén antiadherente 26cm', 'Sartén con recubrimiento antiadherente', 'Cocina', 85.00, 159.00, 15, 5),
('COC002', 'Olla de presión 6L', 'Olla de presión de aluminio 6 litros', 'Cocina', 320.00, 589.00, 8, 3),
('COC003', 'Juego de cuchillos x5', 'Set de 5 cuchillos de acero inoxidable', 'Cocina', 120.00, 229.00, 12, 4),
('COC004', 'Licuadora 3 velocidades', 'Licuadora de vaso de vidrio 1.5L', 'Cocina', 280.00, 499.00, 6, 3),
('COC005', 'Batería de cocina 9 pzas', 'Juego de ollas y sartenes aluminio', 'Cocina', 450.00, 849.00, 5, 2),
('COC006', 'Tostador de pan 2 rebanadas', 'Tostador eléctrico con control', 'Cocina', 150.00, 279.00, 10, 3),
('COC007', 'Tabla para picar bambú', 'Tabla de cortar de bambú 30x40cm', 'Cocina', 45.00, 89.00, 20, 8),
('COC008', 'Colador de acero inoxidable', 'Colador de malla fina 20cm', 'Cocina', 35.00, 69.00, 18, 6),
('COC009', 'Molde para pastel redondo', 'Molde antiadherente 24cm', 'Cocina', 55.00, 99.00, 14, 5),
('COC010', 'Rallador 4 caras', 'Rallador de acero inoxidable', 'Cocina', 28.00, 55.00, 22, 8);

-- Categoría: Limpieza
INSERT OR IGNORE INTO products (code, name, description, category, purchase_price, sale_price, stock, min_stock) VALUES
('LIM001', 'Escoba de plástico', 'Escoba resistente con mango largo', 'Limpieza', 35.00, 65.00, 25, 10),
('LIM002', 'Trapeador de microfibra', 'Trapeador con mango telescópico', 'Limpieza', 55.00, 99.00, 18, 8),
('LIM003', 'Cubeta 12 litros', 'Cubeta plástica resistente', 'Limpieza', 22.00, 45.00, 30, 12),
('LIM004', 'Recogedor con mango', 'Recogedor de basura con mango largo', 'Limpieza', 28.00, 55.00, 20, 8),
('LIM005', 'Cepillo para baño', 'Cepillo con base de plástico', 'Limpieza', 25.00, 49.00, 15, 6),
('LIM006', 'Guantes de hule par', 'Guantes para limpieza talla mediana', 'Limpieza', 12.00, 25.00, 40, 15),
('LIM007', 'Jalador para vidrios', 'Jalador de hule 30cm', 'Limpieza', 30.00, 59.00, 12, 5),
('LIM008', 'Bote de basura 20L', 'Bote con tapa de balancín', 'Limpieza', 65.00, 119.00, 10, 4);

-- Categoría: Baño
INSERT OR IGNORE INTO products (code, name, description, category, purchase_price, sale_price, stock, min_stock) VALUES
('BAN001', 'Cortina para baño', 'Cortina de poliéster 180x180cm', 'Baño', 75.00, 139.00, 12, 4),
('BAN002', 'Tapete antiderrapante', 'Tapete para baño con ventosas', 'Baño', 45.00, 85.00, 15, 5),
('BAN003', 'Organizador de ducha', 'Organizador esquinero 3 niveles', 'Baño', 90.00, 169.00, 8, 3),
('BAN004', 'Porta cepillos de dientes', 'Porta cepillos de cerámica', 'Baño', 35.00, 69.00, 20, 6),
('BAN005', 'Jabonera de cerámica', 'Jabonera decorativa', 'Baño', 30.00, 59.00, 18, 6),
('BAN006', 'Toallero de barra 60cm', 'Toallero de acero cromado', 'Baño', 55.00, 99.00, 10, 4);

-- Categoría: Organización
INSERT OR IGNORE INTO products (code, name, description, category, purchase_price, sale_price, stock, min_stock) VALUES
('ORG001', 'Caja organizadora grande', 'Caja plástica con tapa 50L', 'Organización', 85.00, 159.00, 12, 4),
('ORG002', 'Caja organizadora mediana', 'Caja plástica con tapa 30L', 'Organización', 55.00, 99.00, 15, 5),
('ORG003', 'Ganchos para ropa x10', 'Ganchos de plástico resistente', 'Organización', 18.00, 35.00, 50, 20),
('ORG004', 'Zapatera de 10 niveles', 'Zapatera metálica desmontable', 'Organización', 180.00, 339.00, 6, 2),
('ORG005', 'Canasta de mimbre mediana', 'Canasta decorativa multiusos', 'Organización', 65.00, 119.00, 10, 4),
('ORG006', 'Perchero de pie', 'Perchero metálico 8 ganchos', 'Organización', 150.00, 279.00, 5, 2);

-- Categoría: Herramientas del hogar
INSERT OR IGNORE INTO products (code, name, description, category, purchase_price, sale_price, stock, min_stock) VALUES
('HER001', 'Martillo de uña', 'Martillo de acero con mango de fibra', 'Herramientas', 45.00, 85.00, 12, 5),
('HER002', 'Desarmador juego x6', 'Set de desarmadores plano y cruz', 'Herramientas', 55.00, 99.00, 10, 4),
('HER003', 'Pinzas de presión', 'Pinzas de presión 10 pulgadas', 'Herramientas', 40.00, 75.00, 15, 5),
('HER004', 'Cinta métrica 5m', 'Flexómetro de 5 metros', 'Herramientas', 25.00, 49.00, 20, 8),
('HER005', 'Llave ajustable 10"', 'Llave perica de acero', 'Herramientas', 65.00, 119.00, 8, 3),
('HER006', 'Cinta de aislar negra', 'Cinta aislante 18m', 'Herramientas', 8.00, 18.00, 35, 15);

-- Categoría: Iluminación
INSERT OR IGNORE INTO products (code, name, description, category, purchase_price, sale_price, stock, min_stock) VALUES
('ILU001', 'Foco LED 9W luz blanca', 'Foco ahorrador equivalente 60W', 'Iluminación', 18.00, 35.00, 40, 15),
('ILU002', 'Foco LED 9W luz cálida', 'Foco ahorrador luz cálida', 'Iluminación', 18.00, 35.00, 35, 15),
('ILU003', 'Lámpara de escritorio', 'Lámpara LED flexible con base', 'Iluminación', 120.00, 229.00, 7, 3),
('ILU004', 'Extensión eléctrica 3m', 'Extensión con 3 contactos', 'Iluminación', 35.00, 65.00, 20, 8),
('ILU005', 'Multicontacto 6 entradas', 'Multicontacto con supresor de picos', 'Iluminación', 85.00, 159.00, 12, 5),
('ILU006', 'Linterna LED recargable', 'Linterna de mano 500 lumens', 'Iluminación', 65.00, 119.00, 10, 4);

-- Categoría: Textiles del hogar
INSERT OR IGNORE INTO products (code, name, description, category, purchase_price, sale_price, stock, min_stock) VALUES
('TEX001', 'Juego de sábanas matrimonial', 'Sábanas de microfibra 3 piezas', 'Textiles', 180.00, 339.00, 8, 3),
('TEX002', 'Almohada estándar', 'Almohada de fibra hipoalergénica', 'Textiles', 55.00, 99.00, 15, 5),
('TEX003', 'Toalla de baño grande', 'Toalla de algodón 70x140cm', 'Textiles', 65.00, 119.00, 20, 8),
('TEX004', 'Mantel rectangular', 'Mantel de poliéster 150x250cm', 'Textiles', 75.00, 139.00, 10, 4),
('TEX005', 'Cobertor matrimonial', 'Cobertor de fleece suave', 'Textiles', 220.00, 399.00, 6, 2),
('TEX006', 'Cojín decorativo 45x45', 'Cojín con funda lavable', 'Textiles', 45.00, 85.00, 18, 6);
