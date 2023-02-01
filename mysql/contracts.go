package mysql

import (
	"context"
	"errors"
	"net"

	"github.com/go-sql-driver/mysql"
)

type Dialer interface {
	DialContext(context.Context, string, string) (net.Conn, error)
}

var (
	ErrOptimisticConcurrency = errors.New("another writer has modified the underlying rows")
)

func NormalizeError(err error) error {
	if err == nil {
		return nil
	} else if mysqlErr, ok := err.(*mysql.MySQLError); !ok {
		return err
	} else if mysqlErr.Number != optimisticConcurrencyErrorID {
		return err
	} else {
		return ErrOptimisticConcurrency
	}
}

const optimisticConcurrencyErrorID = 1062
