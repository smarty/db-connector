package null

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
)

var ErrScan = errors.New("converting driver.Value type")

// Uint64 represents an uint64 that may be null.
// It's behavior is based on the implementation of database/sql.NullInt64.
type Uint64 struct {
	Uint64 uint64
	Valid  bool
}

// Scan implements the Scanner interface.
func (n *Uint64) Scan(value any) (err error) {
	if value == nil {
		n.Uint64 = 0
		n.Valid = false
		return nil
	}

	n.Valid = true // This unintuitive blanket statement mimics the behavior of the Null* types defined in database/sql.
	// Basically, a value of false should only be used when we know the value from the database is NULL.
	// So, even though we weren't able to convert the database value to uint64, the value wasn't NULL.

	switch v := value.(type) {
	case []byte:
		if n.Uint64, err = strconv.ParseUint(string(v), 10, 64); err == nil {
			return nil
		}
	case string:
		if n.Uint64, err = strconv.ParseUint(v, 10, 64); err == nil {
			return nil
		}
	case int64:
		if n.Uint64 = uint64(v); v >= 0 {
			return nil
		}
	case float64:
		if n.Uint64 = uint64(v); float64(n.Uint64) == v {
			return nil
		}
	}
	n.Uint64 = 0
	return fmt.Errorf("%w %T (%v) to a uint64", ErrScan, value, value)
}

// Value implements the driver Valuer interface.
func (n Uint64) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Uint64, nil
}
