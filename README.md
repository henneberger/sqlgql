# sqlgql

Write SQL in flat, dotted filenames. Execute that SQL as written, substituting bound variables. Generate the GraphQL and Go glue around it.

```text
queries/
  Query.customer.sql
  Query.customer.customerItems.sql
```

The types are `Customer` and `CustomerItems`: uppercase the first letter of the last filename segment.

## SQL is the source of truth

`Query.customer.sql`:

```sql
-- @one
SELECT id, name, email
FROM customers
WHERE id = {{param.id}}::int
  AND tenant_id = {{env.auth.tenantId}}::text
```

A minimal batched `Query.customer.customerItems.sql`:

```sql
-- @many
-- @batch parent.id=customerId
SELECT customer_id AS "customerId", id, label, quantity
FROM customer_items
WHERE customer_id = ANY({{parent.id}}::int[])
  AND tenant_id = {{env.auth.tenantId}}::text
ORDER BY customer_id, id
```

- `param.id` becomes a GraphQL argument.
- `env.auth.tenantId` comes from the application's request environment.
- `parent.id` collects parent IDs into the SQL array parameter.
- `@batch parent.id=customerId` declares which returned column routes rows to each parent.
- `@one` returns an object or null; `@many` (the default) returns a list.

**The generator never adds joins, predicates, ordering, limits, or batching SQL.** You write `ANY()` yourself. A normal SQL `LIMIT` applies to the entire result. If you want a per-parent limit, write it explicitly in your SQL; the demo files contain a window-function example.

The compiler only binds placeholders and adds sqlc query-name/cardinality comments. Quoted strings and comments are left intact. PostgreSQL describes the statements to infer types without executing them.

## What gets generated

`go generate ./example` orchestrates these steps:

1. `sqlgql`: schema, complete GraphQL operations, bound SQL for sqlc, and binding metadata.
2. `sqlc`: pgx query methods and database row/argument types.
3. `sqlgql-go`: invokes gqlgen and generates resolver adapters, argument binding, row mapping, authorization calls, and request-local loaders. Nested resolver configuration is automatic.
4. GraphQL Code Generator: TypeScript types and typed Apollo documents.
5. The example's manifest script: persisted-operation hashes.

**No handwritten GraphQL selections, resolver bodies, or loaders are needed.** Generated operations select the SQL columns and nested fields. New SQL files participate in the same pipeline.

The SQL compiler stays separate from the optional gqlgen/sqlc adapter generator. sqlc owns SQL execution; gqlgen owns GraphQL types and execution.

## Optional Relay pagination

Pagination is opt-in. **`-- @many` still returns a plain list.** Add `-- @connection id` (or comma-separated result columns for a composite cursor) to opt into a Relay connection:

```sql
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
```

This SQL is authored, not composed by the generator. It defines the pagination predicates and ordering. The generator supplies:

- Optional `first`, `after`, `last`, and `before` arguments.
- `<Type>Connection`, `<Type>Edge`, and `PageInfo` types.
- Every GraphQL selection, typed Apollo document, resolver adapter, and cursor codec.
- `page.limit`: requested size plus one lookahead row.
- `page.backward`: true for `last` requests.
- `page.after.<column>` and `page.before.<column>`: decoded cursor values, or SQL NULL when absent.

Use either `first` or `last`; the default is `first: 20`. Sizes from 0 through 100 are accepted. SQL must return its rows in canonical order for forward requests and reverse order for backward requests. The generated adapter trims the extra row and reverses backward results back to canonical order.

`pageInfo` contains `hasNextPage`, `hasPreviousPage`, `startCursor`, and `endCursor`. Lookahead determines `hasNextPage` going forward and `hasPreviousPage` going backward. The opposite-direction flag is conservatively false, as permitted by the [Relay Connections specification](https://relay.dev/graphql/connections.htm). Empty pages have null start/end cursors.

Cursor keys must be non-null scalar columns and provide a stable unique ordering (include an ID tie-breaker when needed). Cursors are versioned base64url values scoped to the connection, declared keys, environment, parent, and non-pagination arguments. They preserve bigint precision. They are opaque client tokens, not encrypted or signed authorization credentials; OPA still authorizes every load.

Nested connections work with the same explicit `ANY()` SQL and `@batch` mapping. Your SQL applies the limit per parent; the generated loaders still batch compatible requests. See `Query.customerPage.pageItems.sql` for the authored window-function example. A nested cursor belongs to its specific parent.

`Query.customerPage.sql` opts into pagination. `Query.customers.sql` deliberately remains an ordinary list. The generated `CustomerPageDocument` works with the existing Apollo client; adopting the Relay client is not required.

## What stays in your application

HTTP handlers, routes, authentication middleware, server startup, the database connection, and the OPA client/policy are application code. The generator does not write them.

The example's handwritten resolver root just embeds the generated implementation:

```go
type Resolver struct { *SQLResolvers }
```

Supply `Queries`, `Authorizer`, and `Environment` to `SQLResolvers`. In your gqlgen operation middleware, attach operation-local loaders:

```go
gql.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
    return next(sqlResolvers.WithLoaders(ctx))
})
```

See `example/cmd/server/main.go` for the complete application setup. Typed optional override hooks allow custom behavior without editing generated files.

OPA is called before each field's SQL/load, including cache hits. It receives `path`, `env`, `parent`, and `params`. Undefined/false decisions and client errors deny access. Policy decisions remain external.

## Batching behavior

Generated loaders collect compatible parent requests, bind their arrays, and invoke the original sqlc query once. Rows are matched by the declared result key, regardless of return order. Missing rows become an empty list or null.

Different arguments or bound environment values form separate groups. Duplicate keys share work within an operation; caches never carry across operations. Parentless nested queries with identical inputs are deduplicated. The default batching window is 2 ms with at most 128 keys per batch.

Every parent placeholder in a batched query must be an array with a corresponding `@batch` mapping. SQL controls all filtering and pairing of keys. Returned keys that were not requested produce an error. Batched `@one` fields reject multiple rows for the same parent.

## Run the demo

Requires Go 1.23.3+, Node 22+, npm, and Docker Compose. Run from this directory:

```sh
make up          # PostgreSQL on 55432; OPA on 8181
make generate    # runs go generate
make server      # application-owned API on 8080
```

In another terminal:

```sh
make web         # Vite frontend, normally http://127.0.0.1:5173
make smoke       # typed Apollo client against the running API
```

The demo fixes the identity to tenant `acme`, role `reader`. Customer 2 belongs to another tenant and returns null. The generated `Customers` operation exercises batching over customers 1, 3, and 4, including a customer with no items. Limits above 100 are denied by OPA.

Set `DATABASE_URL` and `OPA_URL` to use existing services. `.env.example` documents their defaults; environment files are not loaded automatically. The database schema must match the migrations configured in `example/sqlc.yaml` before generation.

Stop the API/frontend with Ctrl-C; `make down` stops the demo containers.

## Check it

```sh
make test
TEST_DATABASE_URL='postgres://...' go test -race ./...
opa test example/policy
```

The optional integration test uses temporary PostgreSQL tables and automatically closed HTTP test servers. It runs the actual Apollo client through gqlgen and sqlc, and asserts one list SQL call plus one nested SQL call for three parents. Its OPA endpoint is a contract fixture; the Rego tests exercise policy separately.

Tests also cover SQL preservation, binding validation, overflow/null handling, result routing, duplicate keys, argument groups, authorization before cache access, and operation-local caches.

## UUID, JSON, and arrays

| PostgreSQL | GraphQL | TypeScript |
| --- | --- | --- |
| `uuid` | `UUID` | `string` |
| `json`, `jsonb` | `JSON` | Recursive `JSONValue` |
| Supported one-dimensional arrays | Lists, e.g. `[Int]`, `[UUID]`, `[JSON]` | Typed arrays with nullable elements |

These work as query arguments and result columns. UUIDs are validated and normalized, including the all-zero UUID. JSON supports objects, arrays, strings, numbers, booleans, and null. SQL null arrays stay null, empty arrays stay empty, and null elements are preserved.

Array element types cover bool, int2/int4/int8, float4/float8, text/varchar/bpchar, UUID, and JSON/JSONB. Multidimensional arrays are outside the supported API contract.

The supplied sqlc configuration maps JSON to `json.RawMessage` and uses pointer element types to retain null array elements. Those element overrides also make the corresponding scalar sqlc fields pointers. gqlgen scalar mappings and the Go adapters are generated automatically; the TypeScript codegen configuration maps the scalars to client types.

`Query.typeDemo.sql` and the generated Apollo `TypeDemoDocument` exercise these types without casting results to text. The integration test checks them against PostgreSQL, including invalid UUID input.

## Numeric types and boundaries

| PostgreSQL | Go database type | GraphQL | TypeScript |
| --- | --- | --- | --- |
| `real` / `float4` | `float32` | `Float` | `number` |
| `double precision` / `float8` | `float64` | `Float` | `number` |
| `bigint` / `int8` | `int64` | `BigInt` | Decimal `string` |

Both floating types use GraphQL's standard double-precision `Float`, while sqlc retains the PostgreSQL width. Float4 results widen directly to float64, preserving the stored float4 value. Float4 input rounds to float32 precision; overflow and nonzero values that underflow to zero are rejected. Non-finite values cannot be returned as GraphQL floats and produce a field error.

BigInt arguments must be decimal strings. Results are decimal strings too, so values above JavaScript's safe-integer range are never rounded. Conversion to/from SQL uses exact signed 64-bit integers. The supported range is **-9223372036854775808 through 9223372036854775807**, inclusive. Bigint arrays preserve the same precision and null elements.

`Query.numericDemo.sql` exercises scalar and array values through the generated Apollo client. Tests cover bigint min/max and out-of-range values, the 2^53 precision boundary, float32/64 finite maximums and smallest nonzero values, float32 rounding/underflow, float64 overflow, null elements, and PostgreSQL NaN/infinity output.

## Current limits

- Query fields only. No mutation/subscription generation.
- Other PostgreSQL types (such as numeric and timestamps) still require explicit casts to text. No obscure/custom type mappings are added.
- Output scalar fields are conservatively nullable; client arguments are required. General optional/defaulted SQL arguments are not implemented; Relay pagination arguments are optional.
- Object names must be unique across SQL files. Each nested SQL file requires its parent SQL file.
- Generated operations select all descendants. Query shape and SQL cardinality remain explicit; no SQL rewriting or inferred ordering is added.
