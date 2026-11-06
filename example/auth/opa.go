package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/henneberger/sqlgql/binding"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OPAClient is application wiring, not part of the SQL compiler.
type OPAClient struct {
	URL    string
	Client *http.Client
}

func (o *OPAClient) Authorize(ctx context.Context, input Decision) error {
	body, err := json.Marshal(map[string]any{"input": input})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.URL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := o.Client
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("OPA decision unavailable: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("OPA decision failed: HTTP %d", res.StatusCode)
	}
	var decision struct {
		Result *bool `json:"result"`
	}
	if err = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&decision); err != nil {
		return fmt.Errorf("invalid OPA decision: %w", err)
	}
	if decision.Result == nil || !*decision.Result {
		return ErrForbidden
	}
	return nil
}

type contextKey struct{}

func Environment(ctx context.Context) map[string]any {
	env, _ := ctx.Value(contextKey{}).(map[string]any)
	return env
}

// Demo identity is fixed server-side, never read from untrusted tenant/role headers.
// Replace this middleware with your session/JWT validation in a real application.
func DemoIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := map[string]any{"auth": map[string]any{"tenantId": "acme", "role": "reader"}}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, env)))
	})
}

// Policy types are shared with the generated adapters; policy decisions remain external.
type Decision = binding.Decision
type Authorizer = binding.Authorizer

var ErrForbidden = binding.ErrForbidden
