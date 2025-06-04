-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
  id SERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  status BIGINT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
   order_id BIGINT NOT NULL,
   sku_id BIGINT NOT NULL,
   count INT NOT NULL,
   PRIMARY KEY (order_id, sku_id)
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'order_item_type') THEN
CREATE TYPE order_item_type AS (
  order_id bigint,
  sku_id bigint,
  count integer
  );
END IF;
END$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
  DROP TABLE orders;
  DROP TABLE order_items;
-- +goose StatementEnd
