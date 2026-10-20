-- +goose Up
CREATE TABLE restocks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  supplier_name VARCHAR(255) NULL,
  total_amount INT NOT NULL CHECK (total_amount >= 0),
  notes TEXT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_restocks_user ON restocks(user_id);
CREATE INDEX IF NOT EXISTS idx_restocks_created ON restocks(created_at);

CREATE TABLE restock_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  restock_id UUID NOT NULL REFERENCES restocks(id) ON DELETE CASCADE,
  sparepart_id UUID NOT NULL REFERENCES spareparts(id) ON DELETE RESTRICT,
  quantity INT NOT NULL CHECK (quantity > 0),
  unit_cost INT NOT NULL CHECK (unit_cost >= 0),
  subtotal INT NOT NULL CHECK (subtotal >= 0),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_restock_items_restock_id ON restock_items(restock_id);
CREATE INDEX idx_restock_items_sparepart_id ON restock_items(sparepart_id);

-- +goose Down
DROP TABLE IF EXISTS restock_items;
DROP TABLE IF EXISTS restocks;
