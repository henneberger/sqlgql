// Package binding is the application-facing contract used by generated adapters.
// It does not execute SQL or provide HTTP handlers.
package binding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Decision struct {
	Path   string         `json:"path"`
	Env    map[string]any `json:"env"`
	Parent map[string]any `json:"parent"`
	Params map[string]any `json:"params"`
}
type Authorizer interface {
	Authorize(context.Context, Decision) error
}

var ErrForbidden = errors.New("forbidden")

// Convert preserves nulls and rejects numeric overflow instead of silently narrowing.
// It also honors PostgreSQL wrapper types' JSON representations (e.g. UUID).
func Convert[T any](value any) (T, error) {
	var result T
	data, err := json.Marshal(value)
	if err != nil {
		return result, err
	}
	if string(data) == "null" {
		return result, fmt.Errorf("required binding is null")
	}
	err = json.Unmarshal(data, &result)
	return result, err
}
func Lookup(env map[string]any, path string) (any, error) {
	var value any = env
	for _, key := range strings.Split(path, ".") {
		m, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("missing environment binding %s", path)
		}
		value, ok = m[key]
		if !ok {
			return nil, fmt.Errorf("missing environment binding %s", path)
		}
	}
	return value, nil
}

// Copy converts SQLC row values to nullable GraphQL scalar fields without losing nulls.
func Copy[T any](value any) (T, error) {
	var result T
	data, err := json.Marshal(value)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}
