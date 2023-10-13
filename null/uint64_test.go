package null

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestUint64Scan(t *testing.T) {
	cases := []struct {
		Name   string
		Source any
		Want   Uint64
		Err    error
	}{
		{Name: "NULL              ", Source: nil /***********/, Want: Uint64{Valid: false, Uint64: 0}, Err: nil},
		{Name: "valid byte slice  ", Source: []byte("42") /**/, Want: Uint64{Valid: true, Uint64: 42}, Err: nil},
		{Name: "invalid byte slice", Source: []byte("invalid"), Want: Uint64{Valid: true, Uint64: 0}, Err: ErrScan},
		{Name: "positive int64    ", Source: int64(42) /*****/, Want: Uint64{Valid: true, Uint64: 42}, Err: nil},
		{Name: "zero int64        ", Source: int64(0) /******/, Want: Uint64{Valid: true, Uint64: 0}, Err: nil},
		{Name: "negative int64    ", Source: int64(-1) /*****/, Want: Uint64{Valid: true, Uint64: 0}, Err: ErrScan},
		{Name: "integer float64   ", Source: 42.0 /**********/, Want: Uint64{Valid: true, Uint64: 42}, Err: nil},
		{Name: "non-int float64   ", Source: 42.1 /**********/, Want: Uint64{Valid: true, Uint64: 0}, Err: ErrScan},
		{Name: "negative float64  ", Source: -42.0 /*********/, Want: Uint64{Valid: true, Uint64: 0}, Err: ErrScan},
		{Name: "bool              ", Source: true /**********/, Want: Uint64{Valid: true, Uint64: 0}, Err: ErrScan},
		{Name: "string            ", Source: "42" /**********/, Want: Uint64{Valid: true, Uint64: 42}, Err: nil},
		{Name: "time              ", Source: time.Now() /****/, Want: Uint64{Valid: true, Uint64: 0}, Err: ErrScan},
	}
	for _, test := range cases {
		t.Run(strings.TrimSpace(test.Name), func(t *testing.T) {
			v := Uint64{
				Valid:  !test.Want.Valid, // these values should both be re-assigned by Scan in every case.
				Uint64: test.Want.Uint64 + 1,
			}
			err := v.Scan(test.Source)
			if !errors.Is(err, test.Err) {
				t.Fatal("unexpected err value:", err)
			}
			if v != test.Want {
				t.Errorf("\n"+
					"got:  %v\n"+
					"want: %v", v, test.Want,
				)
			}

			// Visually compare outcome with sql.NullInt64:
			var v2 sql.NullInt64
			err2 := v2.Scan(test.Source)
			t.Logf("   \n"+
				"%+v %v\n"+
				"%+v %v",
				v, err,
				v2, err2,
			)
		})
	}
}
func TestUint64Value(t *testing.T) {
	var v Uint64

	v.Uint64 = 42
	value, err := v.Value()
	if value != nil {
		t.Error("value should have been nil, but was:", value)
	}
	if err != nil {
		t.Error("err should have been nil, but was:", err)
	}

	v.Valid = true
	value, err = v.Value()
	if value != uint64(42) {
		t.Error("valid value should have been 42, but was:", value)
	}
	if err != nil {
		t.Error("err should have been nil, but was:", err)
	}
}
