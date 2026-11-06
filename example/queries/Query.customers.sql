-- @many
SELECT id, name, email
FROM customers
WHERE tenant_id = {{env.auth.tenantId}}::text
ORDER BY id
LIMIT {{param.limit}}::int
