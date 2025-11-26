-- Convierte precios a BIGINT (centavos) para evitar truncamiento de decimales.
-- Requiere migracion previa que define price como NUMERIC(18,2).

ALTER TABLE products
    ALTER COLUMN price TYPE BIGINT USING (price * 100)::BIGINT,
    ALTER COLUMN price SET NOT NULL,
    ALTER COLUMN price SET DEFAULT 0;

ALTER TABLE product_history
    ALTER COLUMN price TYPE BIGINT USING (price * 100)::BIGINT,
    ALTER COLUMN price SET NOT NULL;
