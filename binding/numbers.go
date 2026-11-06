package binding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

// BigInt uses decimal strings at the GraphQL boundary and exact int64 values in SQL.
type BigInt string

func parseBigInt(text string) (BigInt, error) {
	v, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid BigInt: %w", err)
	}
	return BigInt(strconv.FormatInt(v, 10)), nil
}
func (b *BigInt) UnmarshalGQL(value any) error {
	v, ok := value.(string)
	if !ok {
		return fmt.Errorf("BigInt must be a decimal string, not a JSON number")
	}
	parsed, err := parseBigInt(v)
	if err == nil {
		*b = parsed
	}
	return err
}
func (b BigInt) MarshalGQL(w io.Writer) { data, _ := json.Marshal(string(b)); _, _ = w.Write(data) }
func (b *BigInt) UnmarshalJSON(data []byte) error {
	var text string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
	} else {
		text = string(data)
	}
	parsed, err := parseBigInt(text)
	if err == nil {
		*b = parsed
	}
	return err
}

// ConvertBigInt converts scalar/array decimal strings to sqlc's integer types
// without ever passing through float64. JSON null elements remain null.
func ConvertBigInt[T any](value any) (T, error) {
	var result T
	data, err := json.Marshal(value)
	if err != nil {
		return result, err
	}
	if string(data) == "null" {
		return result, fmt.Errorf("required binding is null")
	}
	var decoded any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err = decoder.Decode(&decoded); err != nil {
		return result, err
	}
	normalized, err := integerValues(decoded)
	if err != nil {
		return result, err
	}
	data, err = json.Marshal(normalized)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}
func integerValues(value any) (any, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case string:
		parsed, err := parseBigInt(v)
		return json.Number(parsed), err
	case json.Number:
		parsed, err := parseBigInt(string(v))
		return json.Number(parsed), err
	case []any:
		for i, item := range v {
			next, err := integerValues(item)
			if err != nil {
				return nil, err
			}
			v[i] = next
		}
		return v, nil
	default:
		return nil, fmt.Errorf("BigInt requires an integer or decimal string")
	}
}

// ConvertFloat32 rejects underflow to zero as well as overflow. Normal float4
// rounding remains PostgreSQL's/IEEE 754's 32-bit floating-point behavior.
func ConvertFloat32[T any](value any) (T, error) {
	var result T
	data, err := json.Marshal(value)
	if err != nil {
		return result, err
	}
	var decoded any
	if err = json.Unmarshal(data, &decoded); err != nil {
		return result, err
	}
	if err = checkFloat32(decoded); err != nil {
		return result, err
	}
	return Convert[T](value)
}
func checkFloat32(value any) error {
	switch v := value.(type) {
	case nil:
		return nil
	case float64:
		if v != 0 && float32(v) == 0 {
			return fmt.Errorf("float4 underflow: nonzero value rounds to zero")
		}
	case []any:
		for _, item := range v {
			if err := checkFloat32(item); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("float4 requires a number")
	}
	return nil
}
