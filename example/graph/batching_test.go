package graph_test

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/henneberger/sqlgql/binding"
	db "github.com/henneberger/sqlgql/example/dbgen"
	"github.com/henneberger/sqlgql/example/graph"
	"github.com/henneberger/sqlgql/example/graph/model"
)

type authorizer func(context.Context, binding.Decision) error

func (f authorizer) Authorize(ctx context.Context, d binding.Decision) error { return f(ctx, d) }

type queryFixture struct {
	db.Querier
	fetch func(context.Context, db.CustomerCustomerItemsParams) ([]db.CustomerCustomerItemsRow, error)
}

func (f *queryFixture) CustomerCustomerItems(ctx context.Context, p db.CustomerCustomerItemsParams) ([]db.CustomerCustomerItemsRow, error) {
	return f.fetch(ctx, p)
}
func resolvers(f *queryFixture) *graph.SQLResolvers {
	return &graph.SQLResolvers{Queries: f, BatchWait: 20 * time.Millisecond, Authorizer: authorizer(func(context.Context, binding.Decision) error { return nil }), Environment: func(context.Context) map[string]any {
		return map[string]any{"auth": map[string]any{"tenantId": "acme"}}
	}}
}
func parent(id int) *model.Customer { return &model.Customer{ID: &id} }

func TestGeneratedBatchDeduplicatesAndRoutes(t *testing.T) {
	var calls atomic.Int32
	var seen []int32
	f := &queryFixture{fetch: func(ctx context.Context, p db.CustomerCustomerItemsParams) ([]db.CustomerCustomerItemsRow, error) {
		calls.Add(1)
		for _, id := range p.ParentID {
			seen = append(seen, *id)
		}
		var rows []db.CustomerCustomerItemsRow
		// Return parents in reverse order to prove routing uses the returned key.
		for i := len(p.ParentID) - 1; i >= 0; i-- {
			id := *p.ParentID[i]
			if id == 3 {
				continue
			}
			for j := int32(0); j < *p.ParamLimit; j++ {
				rows = append(rows, db.CustomerCustomerItemsRow{CustomerId: ptr(id), ID: ptr(id*100 + j), Label: ptr(fmt.Sprint(id)), Quantity: ptr(int32(1))})
			}
		}
		return rows, nil
	}}
	r := resolvers(f)
	ctx := r.WithLoaders(context.Background())
	ids := []int{1, 1, 2, 3}
	out := make([][]*model.CustomerItems, len(ids))
	errs := make([]error, len(ids))
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		go func(i, id int) { defer wg.Done(); out[i], errs[i] = r.Customer().CustomerItems(ctx, parent(id), 2) }(i, id)
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("wanted one unchanged SQL invocation, got %d", calls.Load())
	}
	sort.Slice(seen, func(i, j int) bool { return seen[i] < seen[j] })
	if fmt.Sprint(seen) != "[1 2 3]" {
		t.Fatal(seen)
	}
	for i, id := range ids {
		if errs[i] != nil {
			t.Fatal(errs[i])
		}
		if id == 3 {
			if out[i] == nil || len(out[i]) != 0 {
				t.Fatal("missing parent must return empty list")
			}
			continue
		}
		if len(out[i]) != 2 || *out[i][0].CustomerID != id || *out[i][0].ID != id*100 {
			t.Fatalf("incorrect route for %d: %+v", id, out[i])
		}
	}
}

func TestGeneratedBatchesSeparateArgumentsAndErrors(t *testing.T) {
	var calls atomic.Int32
	f := &queryFixture{fetch: func(ctx context.Context, p db.CustomerCustomerItemsParams) ([]db.CustomerCustomerItemsRow, error) {
		calls.Add(1)
		if *p.ParamLimit == 9 {
			return nil, errors.New("query failed")
		}
		return []db.CustomerCustomerItemsRow{{CustomerId: p.ParentID[0], ID: ptr(int32(101)), Label: ptr("ok")}}, nil
	}}
	r := resolvers(f)
	ctx := r.WithLoaders(context.Background())
	var wg sync.WaitGroup
	results := make([]error, 2)
	for i, limit := range []int{1, 9} {
		wg.Add(1)
		go func(i, limit int) { defer wg.Done(); _, results[i] = r.Customer().CustomerItems(ctx, parent(1), limit) }(i, limit)
	}
	wg.Wait()
	if calls.Load() != 2 || results[0] != nil || results[1] == nil {
		t.Fatalf("groups: calls=%d errors=%v", calls.Load(), results)
	}
}

func TestGeneratedCacheIsOperationLocalAndAuthorized(t *testing.T) {
	var calls, decisions atomic.Int32
	var deny atomic.Bool
	f := &queryFixture{fetch: func(ctx context.Context, p db.CustomerCustomerItemsParams) ([]db.CustomerCustomerItemsRow, error) {
		calls.Add(1)
		return nil, nil
	}}
	r := resolvers(f)
	r.Authorizer = authorizer(func(context.Context, binding.Decision) error {
		decisions.Add(1)
		if deny.Load() {
			return binding.ErrForbidden
		}
		return nil
	})
	first := r.WithLoaders(context.Background())
	for i := 0; i < 2; i++ {
		if _, err := r.Customer().CustomerItems(first, parent(1), 1); err != nil {
			t.Fatal(err)
		}
	}
	if calls.Load() != 1 || decisions.Load() != 2 {
		t.Fatal("cache must deduplicate SQL without skipping authorization")
	}
	deny.Store(true)
	if _, err := r.Customer().CustomerItems(first, parent(1), 1); !errors.Is(err, binding.ErrForbidden) {
		t.Fatal(err)
	}
	deny.Store(false)
	second := r.WithLoaders(context.Background())
	if _, err := r.Customer().CustomerItems(second, parent(1), 1); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Fatal("cache leaked across operations")
	}
}

func TestGeneratedBatchRejectsUnrequestedRows(t *testing.T) {
	f := &queryFixture{fetch: func(context.Context, db.CustomerCustomerItemsParams) ([]db.CustomerCustomerItemsRow, error) {
		return []db.CustomerCustomerItemsRow{{CustomerId: ptr(int32(999))}}, nil
	}}
	r := resolvers(f)
	if _, err := r.Customer().CustomerItems(r.WithLoaders(context.Background()), parent(1), 1); err == nil {
		t.Fatal("accepted row for an unrequested parent")
	}
}

func ptr[T any](value T) *T { return &value }
