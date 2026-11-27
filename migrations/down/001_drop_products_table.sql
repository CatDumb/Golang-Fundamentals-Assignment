-- Reverts 001_create_products_table.sql.
--   psql -U postgres -d product_management -f migrations/down/001_drop_products_table.sql
-- Kept in a subfolder so docker-compose's init step (which runs top-level .sql files only) skips it.

DROP TABLE IF EXISTS products;
