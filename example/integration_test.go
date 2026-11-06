package example

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/henneberger/sqlgql/example/auth"
	db "github.com/henneberger/sqlgql/example/dbgen"
	"github.com/henneberger/sqlgql/example/graph"
	"github.com/henneberger/sqlgql/example/persisted"
	"github.com/henneberger/sqlgql/internal/compiler"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5"
)

// This test creates connection-local temporary tables, never permanent tables.
// HTTP test servers close automatically. No external service is started.
func TestPostgresApollo(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL + Apollo integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())
	fixture, err := os.ReadFile("db/init.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Exec(ctx, strings.ReplaceAll(string(fixture), "CREATE TABLE", "CREATE TEMP TABLE")); err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	if err = compiler.Generate(ctx, conn, "queries", out); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"schema.graphqls", "queries.sql", "operations/Customer.graphql", "operations/Customers.graphql", "manifest.json", "operations/TypeDemo.graphql", "operations/NumericDemo.graphql", "operations/CustomerPage.graphql"} {
		actual, err := os.ReadFile(filepath.Join(out, name))
		if err != nil {
			t.Fatal(err)
		}
		expected, err := os.ReadFile(filepath.Join("generated", name))
		if err != nil {
			t.Fatal(err)
		}
		if string(actual) != string(expected) {
			t.Fatalf("stale generated file %s; regenerate", name)
		}
	}
	// An HTTP contract fixture exercises the actual OPA client. The Rego policy
	// itself is tested separately with `opa test example/policy`.
	var decisions atomic.Int32
	var numericDecisions atomic.Int32
	var connectionDecisions atomic.Int32
	opa := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Input auth.Decision `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		d := payload.Input
		decisions.Add(1)
		identity, _ := d.Env["auth"].(map[string]any)
		allowed := identity["role"] == "reader" && identity["tenantId"] == "acme"
		switch d.Path {
		case "Query.customerPage", "Query.customerPage.pageItems":
			connectionDecisions.Add(1)
		case "Query.numericDemo":
			numericDecisions.Add(1)
		case "Query.typeDemo":
		case "Query.customer":
			id, _ := d.Params["id"].(float64)
			allowed = allowed && id > 0
		case "Query.customers":
			limit, _ := d.Params["limit"].(float64)
			allowed = allowed && limit > 0 && limit <= 100
		case "Query.customer.customerItems", "Query.customers.items":
			limit, _ := d.Params["limit"].(float64)
			id, _ := d.Parent["id"].(float64)
			allowed = allowed && id > 0 && limit > 0 && limit <= 100
		default:
			allowed = false
		}
		json.NewEncoder(w).Encode(map[string]bool{"result": allowed})
	}))
	defer opa.Close()
	tracked := &trackingDB{DBTX: conn}
	sqlResolvers := &graph.SQLResolvers{Queries: db.New(tracked), BatchWait: 20 * time.Millisecond, Environment: auth.Environment, Authorizer: &auth.OPAClient{URL: opa.URL}}
	gql := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{SQLResolvers: sqlResolvers}}))
	gql.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return next(sqlResolvers.WithLoaders(ctx))
	})
	bytes, err := os.ReadFile("generated/persisted.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]string
	if err = json.Unmarshal(bytes, &manifest); err != nil {
		t.Fatal(err)
	}
	api, err := persisted.Handler(manifest, gql)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(auth.DemoIdentity(api))
	defer server.Close()
	cmd := exec.CommandContext(ctx, "npm", "run", "smoke")
	cmd.Dir = "web"
	cmd.Env = append(os.Environ(), "API_URL="+server.URL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Apollo smoke: %v\n%s", err, output)
	}
	t.Log(string(output))
	if decisions.Load()-numericDecisions.Load()-connectionDecisions.Load() != 15 {
		t.Fatalf("expected 15 OPA decisions, got %d", decisions.Load()-numericDecisions.Load()-connectionDecisions.Load())
	}
	if numericDecisions.Load() != 11 {
		t.Fatalf("expected 11 numeric boundary calls, got %d", numericDecisions.Load())
	}
	if tracked.batches.Load() != 2 || tracked.roots.Load() != 2 {
		t.Fatalf("expected one root + one batch SQL call per list operation, got roots=%d batches=%d", tracked.roots.Load(), tracked.batches.Load())
	}

	if tracked.connectionRoots.Load() != 13 || tracked.connectionBatches.Load() != 9 || tracked.maxConnectionBatch.Load() != 3 {
		t.Fatalf("connection SQL calls: roots=%d children=%d largest batch=%d", tracked.connectionRoots.Load(), tracked.connectionBatches.Load(), tracked.maxConnectionBatch.Load())
	}
}

type trackingDB struct {
	db.DBTX
	batches            atomic.Int32
	roots              atomic.Int32
	connectionRoots    atomic.Int32
	connectionBatches  atomic.Int32
	maxConnectionBatch atomic.Int32
}

func (d *trackingDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	switch {
	case strings.HasPrefix(sql, "-- name: CustomerPage "):
		d.connectionRoots.Add(1)
	case strings.HasPrefix(sql, "-- name: CustomerPagePageItems "):
		d.connectionBatches.Add(1)
		for _, arg := range args {
			if ids, ok := arg.([]*int32); ok {
				size := int32(len(ids))
				for old := d.maxConnectionBatch.Load(); size > old; old = d.maxConnectionBatch.Load() {
					if d.maxConnectionBatch.CompareAndSwap(old, size) {
						break
					}
				}
			}
		}
	case strings.HasPrefix(sql, "-- name: CustomersItems "):
		d.batches.Add(1)

	case strings.HasPrefix(sql, "-- name: Customers "):
		d.roots.Add(1)
	}
	return d.DBTX.Query(ctx, sql, args...)
}
