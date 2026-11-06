package binding

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
)

// UUID is a validated, canonical UUID string, including the all-zero UUID.
type UUID string

func (u *UUID) UnmarshalGQL(value any) error {
	text, ok := value.(string)
	if !ok {
		return fmt.Errorf("UUID must be a string")
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return fmt.Errorf("invalid UUID: %w", err)
	}
	*u = UUID(parsed.String())
	return nil
}
func (u UUID) MarshalGQL(w io.Writer) { data, _ := json.Marshal(string(u)); _, _ = w.Write(data) }
func (u *UUID) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	return u.UnmarshalGQL(value)
}

// JSON preserves arbitrary JSON values: objects, arrays, scalars, and JSON null.
// RawMessage avoids converting JSON numbers to float64 on the Go side.
type JSON json.RawMessage

func (j *JSON) UnmarshalGQL(value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	*j = append((*j)[:0], data...)
	return nil
}
func (j JSON) MarshalGQL(w io.Writer) {
	data, err := j.MarshalJSON()
	if err != nil {
		_, _ = io.WriteString(w, "null")
		return
	}
	_, _ = w.Write(data)
}
func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	if !json.Valid(j) {
		return nil, fmt.Errorf("invalid JSON value")
	}
	return []byte(j), nil
}
func (j *JSON) UnmarshalJSON(data []byte) error {
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON value")
	}
	*j = append((*j)[:0], data...)
	return nil
}
