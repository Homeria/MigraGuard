-- [DANGER] Rename the main orders table.
-- While metadata-only, it instantly breaks all application queries until code is updated.
ALTER TABLE orders RENAME TO customer_orders;
