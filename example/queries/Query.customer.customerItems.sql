-- @many
-- @batch parent.id=customerId
WITH ranked AS (
  SELECT customer_id AS "customerId", id, label, quantity,
         row_number() OVER (PARTITION BY customer_id ORDER BY id) AS position
  FROM customer_items
  WHERE customer_id = ANY({{parent.id}}::int[])
    AND tenant_id = {{env.auth.tenantId}}::text
)
SELECT "customerId", id, label, quantity
FROM ranked
WHERE position <= {{param.limit}}::int
ORDER BY "customerId", id
