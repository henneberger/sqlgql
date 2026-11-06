package binding

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestUUID(t *testing.T) {
	for _, text := range []string{"00000000-0000-0000-0000-000000000000", "123E4567-E89B-12D3-A456-426614174000"} {
		var value UUID
		if err := value.UnmarshalGQL(text); err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		value.MarshalGQL(&buf)
		if !json.Valid(buf.Bytes()) || buf.String() == "null" {
			t.Fatal(buf.String())
		}
	}
	for _, value := range []any{"not-a-uuid", 42, nil} {
		var u UUID
		if err := u.UnmarshalGQL(value); err == nil {
			t.Fatalf("accepted %v", value)
		}
	}
}
func TestJSON(t *testing.T) {
	for _, text := range []string{`{"big":9007199254740993,"nested":[true,null]}`, `[1,"x",false]`, `"text"`, `42`, `true`, `null`} {
		var value JSON
		if err := json.Unmarshal([]byte(text), &value); err != nil {
			t.Fatal(err)
		}
		out, err := json.Marshal(value)
		if err != nil || string(out) != text {
			t.Fatalf("%s => %s (%v)", text, out, err)
		}
		var buf bytes.Buffer
		value.MarshalGQL(&buf)
		if buf.String() != text {
			t.Fatal(buf.String())
		}
	}
	var bad JSON
	if err := bad.UnmarshalJSON([]byte("{bad")); err == nil {
		t.Fatal("invalid JSON accepted")
	}
}
