-- @many
-- @connection id
-- @batch parent.id=customerId
WITH ranked AS (
  SELECT customer_id AS "customerId", id, label, quantity,
         row_number() OVER (
           PARTITION BY customer_id
           ORDER BY CASE WHEN {{page.backward}}::boolean THEN id END DESC,
                    CASE WHEN NOT {{page.backward}}::boolean THEN id END ASC
         ) AS position
  FROM customer_items
  WHERE customer_id = ANY({{parent.id}}::int[])
    AND tenant_id = {{env.auth.tenantId}}::text
    AND ({{page.after.id}}::int IS NULL OR id > {{page.after.id}}::int)
    AND ({{page.before.id}}::int IS NULL OR id < {{page.before.id}}::int)
)
SELECT "customerId", id, label, quantity
FROM ranked
WHERE position <= {{page.limit}}::int
ORDER BY "customerId", position
