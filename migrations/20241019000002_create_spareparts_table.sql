-- +goose Up
CREATE TABLE IF NOT EXISTS spareparts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    selling_price INT4 NOT NULL DEFAULT 0 CHECK (selling_price >= 0),
    stock INT2 NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_spareparts_name ON spareparts(name);

-- +goose Down
DROP TABLE IF EXISTS spareparts;
