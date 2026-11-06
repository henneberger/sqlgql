package example

// The application orchestrates tools. sqlgql knows nothing about gqlgen or Apollo.
//go:generate go run ../cmd/sqlgql -queries queries -out generated
//go:generate go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate
//go:generate go run ../cmd/sqlgql-go -sqlc-package github.com/henneberger/sqlgql/example/dbgen
//go:generate npm --prefix web ci --no-audit --no-fund
//go:generate npm --prefix web run generate
