-- @one
SELECT id, name, email
FROM customers
WHERE id = {{param.id}}::int
  AND tenant_id = {{env.auth.tenantId}}::text
