package binding

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
	"testing"
)

func TestBigIntBoundaries(t *testing.T) {
	for _, text := range []string{"-9223372036854775808", "9223372036854775807", "9007199254740991", "9007199254740992", "9007199254740993", "-9007199254740993", "0"} {
		var scalar BigInt
		if err := scalar.UnmarshalGQL(text); err != nil {
			t.Fatal(err)
		}
		value, err := ConvertBigInt[int64](scalar)
		if err != nil || strconv.FormatInt(value, 10) != text {
			t.Fatalf("input %s => %d (%v)", text, value, err)
		}
		output, err := Copy[BigInt](value)
		if err != nil || string(output) != text {
			t.Fatalf("output %d => %s (%v)", value, output, err)
		}
		var wire bytes.Buffer
		output.MarshalGQL(&wire)
		if wire.String() != strconv.Quote(text) {
			t.Fatal(wire.String())
		}
	}
	for _, input := range []any{"-9223372036854775809", "9223372036854775808", "1.5", "1e3", "NaN", "", float64(9007199254740992), int64(1)} {
		var scalar BigInt
		if err := scalar.UnmarshalGQL(input); err == nil {
			t.Fatalf("invalid BigInt accepted: %v", input)
		}
	}
	values, err := ConvertBigInt[[]*int64]([]any{"-9223372036854775808", nil, "9223372036854775807"})
	if err != nil || len(values) != 3 || values[1] != nil || *values[0] != math.MinInt64 || *values[2] != math.MaxInt64 {
		t.Fatal(values, err)
	}
	for _, v := range []string{"9223372036854775808", "-9223372036854775809"} {
		if _, err := ConvertBigInt[int64](v); err == nil {
			t.Fatal("overflow accepted")
		}
	}
}

func TestFloatingPointBoundaries(t *testing.T) {
	for _, v := range []float64{0, math.MaxFloat32, -math.MaxFloat32, math.SmallestNonzeroFloat32, -math.SmallestNonzeroFloat32, 0.1, 16777217} {
		got, err := ConvertFloat32[float32](v)
		if err != nil || got != float32(v) {
			t.Fatalf("float4 %g => %g (%v)", v, got, err)
		}
	}
	for _, v := range []float64{3.5e38, -3.5e38, 1e-46, -1e-46, math.Inf(1), math.Inf(-1), math.NaN()} {
		if _, err := ConvertFloat32[float32](v); err == nil {
			t.Fatalf("invalid float4 accepted: %g", v)
		}
	}
	if _, err := ConvertFloat32[[]*float32]([]any{1.0, nil, 1e-46}); err == nil {
		t.Fatal("array underflow accepted")
	}
	for _, v := range []float64{0, math.MaxFloat64, -math.MaxFloat64, math.SmallestNonzeroFloat64, -math.SmallestNonzeroFloat64, 0.1} {
		got, err := Convert[float64](v)
		if err != nil || got != v {
			t.Fatalf("float8 %g => %g (%v)", v, got, err)
		}
	}
	for _, v := range []any{json.Number("1.8e308"), json.Number("-1.8e308"), math.Inf(1), math.Inf(-1), math.NaN()} {
		if _, err := Convert[float64](v); err == nil {
			t.Fatalf("invalid float8 accepted: %v", v)
		}
	}
}
