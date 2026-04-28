-- [WARNING] Create a standard index on orders.
-- Takes a Share Lock, blocking writes (INSERT/UPDATE/DELETE) during creation.
CREATE INDEX idx_orders_status ON orders(status);
