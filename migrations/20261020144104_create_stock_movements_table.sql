-- +goose Up
CREATE TYPE stock_movement_direction AS ENUM ('IN', 'OUT');
CREATE TYPE stock_movement_reason AS ENUM ('INITIAL', 'SALE', 'RESTOCK', 'ADJUSTMENT');

CREATE TABLE stock_movements (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  sparepart_id UUID NOT NULL REFERENCES spareparts(id) ON DELETE CASCADE,
  direction stock_movement_direction NOT NULL,
  reason stock_movement_reason NOT NULL,
  qty INT2 NOT NULL CHECK (qty > 0),
  reference_id UUID NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS stock_movements;
