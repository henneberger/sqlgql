// Package persisted supplies application-owned persisted-operation transport.
package persisted

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type request struct {
	Query         string          `json:"query"`
	OperationName string          `json:"operationName,omitempty"`
	Variables     json.RawMessage `json:"variables,omitempty"`
	Extensions    struct {
		PersistedQuery struct {
			Version int    `json:"version"`
			Hash    string `json:"sha256Hash"`
		} `json:"persistedQuery"`
	} `json:"extensions"`
}

// Handler is a prebuilt allowlist, not a dynamically populated APQ cache.
// Unknown operations never reach gqlgen, including requests supplying query text.
func Handler(manifest map[string]string, next http.Handler) (http.Handler, error) {
	for hash, query := range manifest {
		sum := sha256.Sum256([]byte(query))
		if hex.EncodeToString(sum[:]) != hash {
			return nil, fmt.Errorf("invalid persisted operation hash %s", hash)
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", "POST")
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		var req request
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "invalid request", 400)
			return
		}
		var extra any
		if decoder.Decode(&extra) != io.EOF {
			http.Error(w, "invalid request", 400)
			return
		}
		extension := req.Extensions.PersistedQuery
		query, ok := manifest[extension.Hash]
		if !ok || extension.Version != 1 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			io.WriteString(w, `{"errors":[{"message":"Persisted operation not allowed","extensions":{"code":"PERSISTED_QUERY_NOT_ALLOWED"}}]}`)
			return
		}
		if req.Query != "" && req.Query != query {
			http.Error(w, "persisted query text mismatch", 400)
			return
		}
		req.Query = query
		body, err := json.Marshal(req)
		if err != nil {
			http.Error(w, "invalid request", 400)
			return
		}
		r = r.Clone(r.Context())
		r.Body = io.NopCloser(bytes.NewReader(body))
		r.ContentLength = int64(len(body))
		r.Header.Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	}), nil
}
