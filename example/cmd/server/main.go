package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/henneberger/sqlgql/example/auth"
	db "github.com/henneberger/sqlgql/example/dbgen"
	"github.com/henneberger/sqlgql/example/graph"
	"github.com/henneberger/sqlgql/example/persisted"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := pgxpool.New(ctx, env("DATABASE_URL", "postgres://demo:demo@127.0.0.1:55432/demo?sslmode=disable"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err = pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	sqlResolvers := &graph.SQLResolvers{Queries: db.New(pool), Environment: auth.Environment, Authorizer: &auth.OPAClient{URL: env("OPA_URL", "http://127.0.0.1:8181/v1/data/demo/authz/allow")}}
	gql := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{SQLResolvers: sqlResolvers}}))
	gql.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		return next(sqlResolvers.WithLoaders(ctx))
	})
	data, err := os.ReadFile(env("PERSISTED_MANIFEST", "example/generated/persisted.json"))
	if err != nil {
		log.Fatal(err)
	}
	var manifest map[string]string
	if err = json.Unmarshal(data, &manifest); err != nil {
		log.Fatal(err)
	}
	api, err := persisted.Handler(manifest, gql)
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/graphql", auth.DemoIdentity(api))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	srv := &http.Server{Addr: env("LISTEN_ADDR", "127.0.0.1:8080"), Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdown)
	}()
	log.Printf("GraphQL API: http://%s/graphql", srv.Addr)
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
