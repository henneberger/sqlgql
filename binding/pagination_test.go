package binding

import (
	"encoding/base64"
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

func pInt(v int) *int          { return &v }
func pString(v string) *string { return &v }
func TestPaginationBounds(t *testing.T) {
	for _, tc := range []struct {
		first, last *int
		valid       bool
	}{{nil, nil, true}, {pInt(0), nil, true}, {pInt(100), nil, true}, {nil, pInt(1), true}, {pInt(-1), nil, false}, {pInt(101), nil, false}, {pInt(1), pInt(1), false}} {
		_, err := PreparePage("q", []string{"id"}, nil, nil, nil, tc.first, nil, tc.last, nil)
		if (err == nil) != tc.valid {
			t.Fatalf("%+v: %v", tc, err)
		}
	}
}
func TestCursorScopeAndPrecision(t *testing.T) {
	env := map[string]any{"tenant": "acme"}
	parent := map[string]any{"id": 1}
	params := map[string]any{"search": "x", "first": 1}
	p, err := PreparePage("q", []string{"key", "id"}, env, parent, params, pInt(1), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	text, err := p.Cursor("same", BigInt("9223372036854775807"))
	if err != nil {
		t.Fatal(err)
	}
	next, err := PreparePage("q", []string{"key", "id"}, env, parent, map[string]any{"search": "x", "first": 2}, pInt(2), &text, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	id, err := OptionalBigInt[*int64](next.Value("after.id"))
	if err != nil || *id != math.MaxInt64 {
		t.Fatal(id, err)
	}
	for _, tc := range []struct {
		path                string
		env, parent, params map[string]any
	}{{"other", env, parent, params}, {"q", map[string]any{"tenant": "other"}, parent, params}, {"q", env, map[string]any{"id": 2}, params}, {"q", env, parent, map[string]any{"search": "y"}}} {
		if _, err := PreparePage(tc.path, []string{"key", "id"}, tc.env, tc.parent, tc.params, nil, &text, nil, nil); err == nil {
			t.Fatal("cross-scope cursor accepted")
		}
	}
	if _, err := p.Cursor("same", nil); err == nil {
		t.Fatal("null cursor key accepted")
	}
	raw, _ := base64.RawURLEncoding.DecodeString(text)
	var c cursor
	json.Unmarshal(raw, &c)
	c.Version = 99
	changed, _ := json.Marshal(c)
	bad := base64.RawURLEncoding.EncodeToString(changed)
	for _, value := range []string{"", "not valid base64!", bad} {
		if _, err := PreparePage("q", []string{"key", "id"}, env, parent, params, nil, &value, nil, nil); err == nil {
			t.Fatal("malformed cursor accepted")
		}
	}
}
func TestPageRowsDoesNotMutateCache(t *testing.T) {
	original := []int{4, 3, 2}
	got, more := PageRows(original, &Page{Size: 2, Backward: true})
	if !more || !reflect.DeepEqual(got, []int{3, 4}) || !reflect.DeepEqual(original, []int{4, 3, 2}) {
		t.Fatal(got, original, more)
	}
	empty, more := PageRows(original, &Page{Size: 0})
	if len(empty) != 0 || !more {
		t.Fatal(empty, more)
	}
	empty, more = PageRows([]int{}, &Page{Size: 2})
	if empty == nil || len(empty) != 0 || more {
		t.Fatal(empty, more)
	}
}
