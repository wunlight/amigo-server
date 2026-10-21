-- +goose Up
CREATE TABLE stock_opnames (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  idempotency_key VARCHAR(255) NOT NULL UNIQUE,
  notes TEXT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_stock_opnames_user ON stock_opnames(user_id);
CREATE INDEX IF NOT EXISTS idx_stock_opnames_created ON stock_opnames(created_at);

CREATE TABLE stock_opname_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  stock_opname_id UUID NOT NULL REFERENCES stock_opnames(id) ON DELETE CASCADE,
  sparepart_id UUID NOT NULL REFERENCES spareparts(id) ON DELETE RESTRICT,
  quantity INT NOT NULL CHECK (quantity > 0),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_stock_opname_items_opname_id ON stock_opname_items(stock_opname_id);
CREATE INDEX idx_stock_opname_items_sparepart_id ON stock_opname_items(sparepart_id);

-- +goose Down
DROP TABLE IF EXISTS stock_opname_items;
DROP TABLE IF EXISTS stock_opnames;
