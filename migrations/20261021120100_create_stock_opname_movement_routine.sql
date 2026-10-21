-- +goose Up

-- Function & Trigger Stock Opname Item Movement
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_auto_create_opname_stock_movement()
RETURNS TRIGGER AS $$
DECLARE
  current_stock INT;
  new_quantity INT;
BEGIN
  SELECT stock INTO current_stock
  FROM spareparts
  WHERE id = NEW.sparepart_id;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Sparepart with ID % not found', NEW.sparepart_id;
  END IF;

  new_quantity := NEW.quantity;
  
  IF new_quantity > current_stock THEN
    INSERT INTO stock_movements (sparepart_id, direction, reason, qty, reference_id)
    VALUES (NEW.sparepart_id, 'IN', 'ADJUSTMENT', (new_quantity - current_stock), NEW.stock_opname_id);
  ELSIF new_quantity < current_stock THEN
    INSERT INTO stock_movements (sparepart_id, direction, reason, qty, reference_id)
    VALUES (NEW.sparepart_id, 'OUT', 'ADJUSTMENT', (current_stock - new_quantity), NEW.stock_opname_id);
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_stock_opname_item_movement
  BEFORE INSERT ON stock_opname_items
  FOR EACH ROW
  EXECUTE FUNCTION fn_auto_create_opname_stock_movement();
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_stock_opname_item_movement ON stock_opname_items;
DROP FUNCTION IF EXISTS fn_auto_create_opname_stock_movement();
