package binding

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Page describes an opt-in Relay request. SQL owns filtering and ordering.
type Page struct {
	Size          int
	Backward      bool
	After, Before map[string]json.RawMessage
	scope         string
	fields        []string
}
type cursor struct {
	Version int               `json:"v"`
	Scope   string            `json:"s"`
	Values  []json.RawMessage `json:"k"`
}

func PreparePage(path string, fields []string, env, parent, params map[string]any, first *int, after *string, last *int, before *string) (*Page, error) {
	if first != nil && last != nil {
		return nil, fmt.Errorf("use either first or last, not both")
	}
	size := 20
	if first != nil {
		size = *first
	}
	if last != nil {
		size = *last
	}
	if size < 0 || size > 100 {
		return nil, fmt.Errorf("page size must be between 0 and 100")
	}
	filters := map[string]any{}
	for key, value := range params {
		switch key {
		case "first", "last", "after", "before":
		default:
			filters[key] = value
		}
	}
	data, err := json.Marshal([]any{path, fields, env, parent, filters})
	if err != nil {
		return nil, err
	}
	scope := fmt.Sprintf("%x", sha256.Sum256(data))
	p := &Page{Size: size, Backward: last != nil, scope: scope, fields: fields}
	if p.After, err = p.decode(after); err != nil {
		return nil, err
	}
	if p.Before, err = p.decode(before); err != nil {
		return nil, err
	}
	return p, nil
}
func (p *Page) decode(text *string) (map[string]json.RawMessage, error) {
	values := map[string]json.RawMessage{}
	if text == nil {
		return values, nil
	}
	if len(*text) > 8192 {
		return nil, fmt.Errorf("invalid pagination cursor")
	}
	raw, err := base64.RawURLEncoding.DecodeString(*text)
	if err != nil {
		return nil, fmt.Errorf("invalid pagination cursor")
	}
	var c cursor
	if err = json.Unmarshal(raw, &c); err != nil || c.Version != 1 || c.Scope != p.scope || len(c.Values) != len(p.fields) {
		return nil, fmt.Errorf("invalid pagination cursor for this connection")
	}
	for i, key := range p.fields {
		if string(c.Values[i]) == "null" {
			return nil, fmt.Errorf("cursor keys cannot be null")
		}
		values[key] = c.Values[i]
	}
	return values, nil
}
func (p *Page) Value(path string) any {
	switch path {
	case "limit":
		return p.Size + 1
	case "backward":
		return p.Backward
	}
	for _, direction := range []string{"after", "before"} {
		prefix := direction + "."
		if len(path) > len(prefix) && path[:len(prefix)] == prefix {
			values := p.After
			if direction == "before" {
				values = p.Before
			}
			if value, ok := values[path[len(prefix):]]; ok {
				return value
			}
			return nil
		}
	}
	return nil
}
func (p *Page) Cursor(values ...any) (string, error) {
	if len(values) != len(p.fields) {
		return "", fmt.Errorf("cursor key count mismatch")
	}
	c := cursor{Version: 1, Scope: p.scope}
	for _, value := range values {
		raw, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		if string(raw) == "null" {
			return "", fmt.Errorf("cursor keys cannot be null")
		}
		c.Values = append(c.Values, raw)
	}
	raw, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	if len(encoded) > 8192 {
		return "", fmt.Errorf("pagination cursor exceeds 8192 bytes")
	}
	return encoded, nil
}

// PageRows trims the lookahead row and restores canonical ordering for last/before.
// It copies the slice so request-local loader caches cannot be mutated by reversal.
func PageRows[T any](rows []T, page *Page) ([]T, bool) {
	more := len(rows) > page.Size
	end := len(rows)
	if more {
		end = page.Size
	}
	out := append([]T{}, rows[:end]...)
	if page.Backward {
		for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
			out[i], out[j] = out[j], out[i]
		}
	}
	return out, more
}
func OptionalBigInt[T any](value any) (T, error) {
	if value == nil {
		return Copy[T](nil)
	}
	return ConvertBigInt[T](value)
}
