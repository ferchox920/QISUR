-- Update products table to include price and remove legacy columns.
ALTER TABLE products
    DROP COLUMN IF EXISTS price_cents,
    DROP COLUMN IF EXISTS category_id,
    ADD COLUMN IF NOT EXISTS price NUMERIC(18,2) NOT NULL DEFAULT 0;

-- Product to Category many-to-many
CREATE TABLE IF NOT EXISTS product_category (
    product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, category_id)
);

-- Product history for price/stock changes
CREATE TABLE IF NOT EXISTS product_history (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price      NUMERIC(18,2) NOT NULL,
    stock      BIGINT NOT NULL DEFAULT 0,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_product_history_product ON product_history(product_id);