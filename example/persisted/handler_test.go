package persisted

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAllowlist(t *testing.T) {
	query := "query Customer { customer { id } }"
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(query)))
	calls := 0
	h, err := Handler(map[string]string{hash: query}, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var req request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Query != query {
			t.Fatal(req.Query)
		}
		w.WriteHeader(200)
	}))
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		body   string
		status int
	}{{fmt.Sprintf(`{"extensions":{"persistedQuery":{"version":1,"sha256Hash":%q}}}`, hash), 200}, {`{"query":"{ arbitrary }"}`, 400}, {`{"extensions":{"persistedQuery":{"version":1,"sha256Hash":"unknown"}}}`, 400}, {fmt.Sprintf(`{"query":"{ other }","extensions":{"persistedQuery":{"version":1,"sha256Hash":%q}}}`, hash), 400}, {`{} {}`, 400}} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", "/graphql", strings.NewReader(tt.body)))
		if w.Code != tt.status {
			t.Fatalf("%s: %d %s", tt.body, w.Code, w.Body)
		}
	}
	if calls != 1 {
		t.Fatalf("unexpected executor calls: %d", calls)
	}
	if _, err := Handler(map[string]string{"bad": query}, http.NotFoundHandler()); err == nil {
		t.Fatal("accepted invalid manifest")
	}
}
