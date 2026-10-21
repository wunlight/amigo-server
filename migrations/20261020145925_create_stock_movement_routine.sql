-- +goose Up

-- 1. Function & Trigger Stock Movement
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_process_stock_movement()
RETURNS TRIGGER AS $$
DECLARE
  current_stock INT;
  new_stock INT;
BEGIN
  SELECT stock INTO current_stock
  FROM spareparts
  WHERE id = NEW.sparepart_id
  FOR UPDATE;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Sparepart with ID % not found', NEW.sparepart_id;
  ELSIF NEW.direction = 'OUT' THEN
    new_stock := current_stock - NEW.qty;
    IF new_stock < 0 THEN
      RAISE EXCEPTION 'Insufficient stock for sparepart ID %. Current stock: %, Requested: %', 
        NEW.sparepart_id, current_stock, NEW.qty USING ERRCODE = '22003';
    END IF;

    UPDATE spareparts
    SET stock = new_stock, updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.sparepart_id;
  ELSIF NEW.direction = 'IN' THEN
    new_stock := current_stock + NEW.qty;

    UPDATE spareparts
    SET stock = new_stock, updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.sparepart_id;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_stock_movement
  BEFORE INSERT ON stock_movements
  FOR EACH ROW
  EXECUTE FUNCTION fn_process_stock_movement();
-- +goose StatementEnd

-- 2. Function & Trigger Sale Item Stock Deduction
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION fn_auto_create_sale_stock_movement()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO stock_movements (sparepart_id, direction, reason, qty, reference_id)
  VALUES (NEW.sparepart_id, 'OUT', 'SALE', NEW.quantity, NEW.sale_id);

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trg_sale_item_stock_deduction
  BEFORE INSERT ON sale_sparepart_items
  FOR EACH ROW
  EXECUTE FUNCTION fn_auto_create_sale_stock_movement();
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_sale_item_stock_deduction ON sale_sparepart_items;
DROP FUNCTION IF EXISTS fn_auto_create_sale_stock_movement();

DROP TRIGGER IF EXISTS trg_stock_movement ON stock_movements;
DROP FUNCTION IF EXISTS fn_process_stock_movement();