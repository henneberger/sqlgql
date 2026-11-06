package main

import (
	"context"
	"github.com/henneberger/sqlgql/internal/compiler"
	"flag"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
	"time"
)

func main() {
	dir := flag.String("queries", "queries", "directory containing flat Query.*.sql files")
	out := flag.String("out", "generated", "generator-owned output directory")
	flag.Parse()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)
	if err = compiler.Generate(ctx, conn, *dir, *out); err != nil {
		log.Fatal(err)
	}
}
