package null

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/smarty/db-connector/mysql"
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
		{Name: "actual uint64     ", Source: uint64(42) /****/, Want: Uint64{Valid: true, Uint64: 42}, Err: nil},
		{Name: "large uint64      ", Source: largeUint64 /***/, Want: Uint64{Valid: true, Uint64: largeUint64}, Err: nil},
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
func TestUint64Value_Invalid(t *testing.T) {
	var v Uint64
	v.Valid = false
	v.Uint64 = 42

	value, err := v.Value()

	if value != nil {
		t.Error("value should have been nil, but was:", value)
	}
	if err != nil {
		t.Error("err should have been nil, but was:", err)
	}
}
func TestUint64Value_ValidInt64Value(t *testing.T) {
	var v Uint64

	v.Uint64 = 42
	v.Valid = true

	value, err := v.Value()

	if value != uint64(42) {
		t.Errorf("database/sql requires uint64(42), but was: %s(%d)", reflect.TypeOf(value), value)
	}
	if err != nil {
		t.Error("err should have been nil, but was:", err)
	}
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	db, err := mysql.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}

	exec(t, db, "DROP SCHEMA IF EXISTS dbConnector;")
	defer exec(t, db, "DROP SCHEMA IF EXISTS dbConnector;")
	exec(t, db, "CREATE SCHEMA IF NOT EXISTS dbConnector;")
	exec(t, db, "DROP TABLE IF EXISTS dbConnector.null_uint64_integration_test;")
	exec(t, db, `CREATE TABLE IF NOT EXISTS dbConnector.null_uint64_integration_test (
		  id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
		  special BIGINT unsigned NULL
		);`,
	)

	var (
		null  Uint64
		small = Uint64{Uint64: 42, Valid: true}
		large = Uint64{Uint64: largeUint64, Valid: true}
	)
	numbers := []Uint64{null, small, large}
	exec(t, db,
		"INSERT INTO dbConnector.null_uint64_integration_test (special) VALUES (?), (?), (?);",
		null, small, large,
	)

	rows, err := db.Query("SELECT special FROM dbConnector.null_uint64_integration_test;")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	for x := 0; rows.Next(); x++ {
		var toScan Uint64
		err = rows.Scan(&toScan)
		if err != nil {
			t.Fatal(err)
		}
		if numbers[x].Uint64 != toScan.Uint64 {
			t.Errorf("\n"+
				"got:  %v\n"+
				"want: %v",
				toScan.Uint64,
				numbers[x].Uint64,
			)
		}
	}

	row := db.QueryRow("SELECT COUNT(*) FROM dbConnector.null_uint64_integration_test WHERE special IS NULL;")
	var nullCount int
	err = row.Scan(&nullCount)
	if err != nil {
		t.Fatal(err)
	}
	if nullCount != 1 {
		t.Errorf("\n"+
			"got:  %v\n"+
			"want: %v",
			nullCount,
			1,
		)
	}
}
func exec(t *testing.T, db *sql.DB, statement string, args ...any) {
	t.Helper()
	_, err := db.Exec(statement, args...)
	if err != nil {
		t.Fatal(err)
	}
}

const (
	maxUint64   = ^uint64(0)
	maxInt64    = maxUint64 >> 1
	largeUint64 = maxInt64 + 1
)
