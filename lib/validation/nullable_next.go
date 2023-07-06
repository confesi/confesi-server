package validation

import (
	"encoding/json"
	"time"
)

// NullableNext is a nullable int64 holding a unix epoch timestamp in microseconds.
type NullableNext struct {
	Next *int64 `json:"next" binding:"omitempty"`
}

// Returns either nothing if cursor was null, or [function arg] + [cursor] concatenated
// syntactically correct for a SQL query.
func (n *NullableNext) Cursor(sql string) string {
	if n.IsNull() {
		return ""
	}
	return sql + "'" + time.Unix(0, n.Value()*int64(time.Microsecond)).Format("2006-01-02 15:04:05.000000") + "'"
}

func (n *NullableNext) IsNull() bool {
	return n.Next == nil
}

func (n *NullableNext) IsNotNull() bool {
	return n.Next != nil
}

func (n *NullableNext) Value() int64 {
	return *n.Next
}

func (h *NullableNext) Set(value int64) {
	h.Next = &value
}

func (n *NullableNext) Time() time.Time {
	return time.Unix(0, n.Value()*int64(time.Microsecond))
}

func (n *NullableNext) SQLTime() string {
	return "'" + time.Unix(0, n.Value()*int64(time.Microsecond)).Format("2006-01-02 15:04:05.000000") + "'"
}

func (n *NullableNext) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Next = nil
		return nil
	}
	var value int64
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	n.Next = &value
	return nil
}
