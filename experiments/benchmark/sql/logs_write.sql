\set order_id random(1, 100000)

BEGIN;
INSERT INTO order_event_logs (order_id, event_type, raw_payload)
VALUES (:order_id, 'BENCHMARK_EVENT', '{"source":"pgbench"}'::jsonb);
SELECT id
  FROM order_event_logs
 WHERE order_id = :order_id
 ORDER BY id DESC
 LIMIT 1;
COMMIT;
