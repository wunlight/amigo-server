-- +goose Up
CREATE TABLE IF NOT EXISTS sales (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    total_amount INT NOT NULL DEFAULT 0,
    notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sales_user ON sales(user_id);
CREATE INDEX IF NOT EXISTS idx_sales_created ON sales(created_at);

CREATE TABLE IF NOT EXISTS sale_sparepart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id UUID NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    sparepart_id UUID NOT NULL REFERENCES spareparts(id),
    quantity INT2 NOT NULL CHECK (quantity > 0),
    unit_price INT NOT NULL CHECK (unit_price >= 0),
    subtotal INT NOT NULL CHECK (subtotal >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sale_sparepart_sale ON sale_sparepart_items(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_sparepart_sparepart ON sale_sparepart_items(sparepart_id);

CREATE TABLE IF NOT EXISTS sale_operation_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sale_id UUID NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    operation_id UUID NOT NULL REFERENCES operations(id),
    quantity INT2 NOT NULL CHECK (quantity > 0),
    unit_price INT NOT NULL CHECK (unit_price >= 0),
    subtotal INT NOT NULL CHECK (subtotal >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_sale_operation_sale ON sale_operation_items(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_operation_operation ON sale_operation_items(operation_id);

-- +goose Down
DROP TABLE IF EXISTS sale_operation_items;
DROP TABLE IF EXISTS sale_sparepart_items;
DROP TABLE IF EXISTS sales;