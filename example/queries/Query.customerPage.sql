-- @many
-- @connection id
SELECT id, name, email
FROM customers
WHERE tenant_id = {{env.auth.tenantId}}::text
  AND ({{page.after.id}}::int IS NULL OR id > {{page.after.id}}::int)
  AND ({{page.before.id}}::int IS NULL OR id < {{page.before.id}}::int)
ORDER BY CASE WHEN {{page.backward}}::boolean THEN id END DESC,
         CASE WHEN NOT {{page.backward}}::boolean THEN id END ASC
LIMIT {{page.limit}}::int
