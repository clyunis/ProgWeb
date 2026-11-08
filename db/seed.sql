-- Datos iniciales: ciudades disponibles para publicar.
-- Se cargan una sola vez, cuando Docker crea la base (después del schema).
INSERT INTO CIUDAD (nombre, provincia) VALUES
    ('Tandil', 'Buenos Aires'),
    ('Mar del Plata', 'Buenos Aires'),
    ('La Plata', 'Buenos Aires'),
    ('Bahía Blanca', 'Buenos Aires'),
    ('Ciudad Autónoma de Buenos Aires', 'CABA'),
    ('Córdoba', 'Córdoba'),
    ('Rosario', 'Santa Fe'),
    ('Mendoza', 'Mendoza'),
    ('San Carlos de Bariloche', 'Río Negro'),
    ('Salta', 'Salta');
