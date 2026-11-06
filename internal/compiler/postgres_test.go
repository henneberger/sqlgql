package compiler

import (
	"context"
	"github.com/jackc/pgx/v5"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostgresValidation(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"missing parent", map[string]string{"Query.customer.items.sql": "SELECT 1::int AS id"}, "missing parent"},
		{"duplicate column", map[string]string{"Query.customer.sql": "SELECT 1::int AS id, 2::int AS id"}, "duplicate column"},
		{"unsupported precision", map[string]string{"Query.customer.sql": "SELECT 123::numeric AS id"}, "unsupported PostgreSQL type"},
		{"parent missing column", map[string]string{"Query.customer.sql": "SELECT 1::int AS id", "Query.customer.items.sql": "SELECT {{parent.missing}}::int AS id"}, "parent column missing not found"},
		{"root parent", map[string]string{"Query.customer.sql": "SELECT {{parent.id}}::int AS id"}, "root has no parent"},
		{"type collision", map[string]string{"Query.customer.sql": "SELECT 1::int AS id", "Query.customer.customer.sql": "SELECT 1::int AS id"}, "type Customer collides"},
		{"field collision", map[string]string{"Query.customer.sql": "SELECT 1::int AS items", "Query.customer.items.sql": "SELECT 1::int AS id"}, "child conflicts"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, sql := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(sql), 0600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := Inspect(ctx, conn, dir)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected %q, got %v", tt.want, err)
			}
		})
	}
}
