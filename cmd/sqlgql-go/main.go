package main

import (
	"github.com/henneberger/sqlgql/internal/goadapter"
	"flag"
	"log"
)

func main() {
	manifest := flag.String("manifest", "generated/manifest.json", "SQL metadata manifest")
	cfg := flag.String("config", "gqlgen.yml", "application gqlgen config")
	db := flag.String("sqlc-package", "", "sqlc generated Go package import path")
	flag.Parse()
	if *db == "" {
		log.Fatal("-sqlc-package is required")
	}
	if err := goadapter.Generate(*manifest, *cfg, *db); err != nil {
		log.Fatal(err)
	}
}
