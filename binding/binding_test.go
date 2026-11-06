package binding

import "testing"

func TestBindingConversions(t *testing.T) {
	if _, err := Convert[int32](int64(1) << 40); err == nil {
		t.Fatal("integer overflow accepted")
	}
	if _, err := Convert[int32](1.5); err == nil {
		t.Fatal("fractional integer accepted")
	}
	var absent *int
	if _, err := Convert[int32](absent); err == nil {
		t.Fatal("null binding accepted")
	}
	if v, err := Copy[*string](nil); err != nil || v != nil {
		t.Fatal("null output lost")
	}
	if value, err := Lookup(map[string]any{"auth": map[string]any{"tenantId": "acme"}}, "auth.tenantId"); err != nil || value != "acme" {
		t.Fatal(value, err)
	}
	if _, err := Lookup(nil, "auth.tenantId"); err == nil {
		t.Fatal("missing environment accepted")
	}
}
