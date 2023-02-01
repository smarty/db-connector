package mysql

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
)

func New(options ...option) (*sql.DB, error) {
	driverConfig := &mysql.Config{
		Params:                  map[string]string{},
		Loc:                     time.UTC,
		MaxAllowedPacket:        4194304,
		AllowAllFiles:           false,
		AllowCleartextPasswords: false,
		AllowNativePasswords:    false,
		AllowOldPasswords:       false,
		CheckConnLiveness:       true,
		ColumnsWithAlias:        false,
	}
	config := configuration{DriverConfig: driverConfig}
	Options.apply(options...)(&config)

	_ = mysql.SetLogger(config.Logger)

	var encryption string = "plaintext"
	if config.TLSConfig != nil {
		encryption = "TLS"
		driverConfig.TLSConfig = strconv.FormatInt(config.TLSIdentifier, 10)
		_ = mysql.RegisterTLSConfig(driverConfig.TLSConfig, config.TLSConfig)
	}

	network := driverConfig.Net
	if config.Dialer != nil {
		driverConfig.Net = strconv.FormatInt(config.TLSIdentifier, 10)
		mysql.RegisterDialContext(driverConfig.Net, func(ctx context.Context, address string) (net.Conn, error) {
			if connection, err := config.Dialer.DialContext(ctx, network, address); err != nil {
				return nil, err
			} else {
				config.Logger.Printf("[INFO] Established [%s] MySQL database connection [%s] with user [%s] to [%s://%s] using schema [%s].", encryption, config.Name, driverConfig.User, network, driverConfig.Addr, driverConfig.DBName)
				return connection, nil
			}
		})
	}

	config.Logger.Printf("[INFO] Establishing [%s] MySQL database connection [%s] with user [%s] to [%s://%s] using schema [%s]...", encryption, config.Name, driverConfig.User, network, driverConfig.Addr, driverConfig.DBName)
	connector, err := mysql.NewConnector(driverConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to establish MySQL database handle: %w", err)
	}

	handle := sql.OpenDB(connector)
	handle.SetConnMaxIdleTime(config.MaxConnectionIdleTimeout)
	handle.SetConnMaxLifetime(config.MaxConnectionLifetime)
	handle.SetMaxOpenConns(config.MaxOpenConnections)
	handle.SetMaxIdleConns(config.MaxIdleConnections)
	return handle, nil
}

type configuration struct {
	TLSConfig                *tls.Config
	TLSIdentifier            int64
	Dialer                   Dialer
	DialerIdentifier         int64
	Name                     string
	DriverConfig             *mysql.Config
	MaxConnectionIdleTimeout time.Duration
	MaxConnectionLifetime    time.Duration
	MaxIdleConnections       int
	MaxOpenConnections       int
	Logger                   logger
}

func (singleton) TLS(value *tls.Config) option {
	return func(this *configuration) {
		this.TLSConfig = value
		if value != nil {
			this.TLSIdentifier = time.Now().UTC().UnixNano()
		} else {
			this.TLSIdentifier = 0
		}
	}
}
func (singleton) Dialer(value Dialer) option {
	return func(this *configuration) {
		this.Dialer = value
		if value != nil {
			this.DialerIdentifier = time.Now().UTC().UnixNano()
		} else {
			this.DialerIdentifier = 0
		}
	}
}
func (singleton) Name(value string) option {
	return func(this *configuration) { this.Name = value }
}
func (singleton) Username(value string) option {
	return func(this *configuration) { this.DriverConfig.User = value }
}
func (singleton) Password(value string) option {
	return func(this *configuration) { this.DriverConfig.Passwd = value }
}
func (singleton) Network(value string) option {
	return func(this *configuration) { this.DriverConfig.Net = value }
}
func (singleton) Address(value string) option {
	return func(this *configuration) { this.DriverConfig.Addr = value }
}
func (singleton) Schema(value string) option {
	return func(this *configuration) { this.DriverConfig.DBName = value }
}
func (singleton) Collation(value string) option {
	return func(this *configuration) { this.DriverConfig.Collation = value }
}
func (singleton) ParseTime(value bool) option {
	return func(this *configuration) { this.DriverConfig.ParseTime = value }
}
func (singleton) InterpolateParameters(value bool) option {
	return func(this *configuration) { this.DriverConfig.InterpolateParams = value }
}
func (singleton) MultipleStatements(value bool) option {
	return func(this *configuration) { this.DriverConfig.MultiStatements = value }
}
func (singleton) AllowReadOnly(value bool) option {
	return func(this *configuration) { this.DriverConfig.RejectReadOnly = !value }
}
func (singleton) ClientFoundRows(value bool) option {
	return func(this *configuration) { this.DriverConfig.ClientFoundRows = value }
}
func (singleton) DialTimeout(value time.Duration) option {
	return func(this *configuration) { this.DriverConfig.Timeout = value }
}
func (singleton) ReadTimeout(value time.Duration) option {
	return func(this *configuration) { this.DriverConfig.ReadTimeout = value }
}
func (singleton) WriteTimeout(value time.Duration) option {
	return func(this *configuration) { this.DriverConfig.WriteTimeout = value }
}
func (singleton) MaxConnectionIdleTimeout(value time.Duration) option {
	return func(this *configuration) { this.MaxConnectionIdleTimeout = value }
}
func (singleton) MaxConnectionLifetime(value time.Duration) option {
	return func(this *configuration) { this.MaxConnectionLifetime = value }
}
func (singleton) MaxOpenConnections(value uint16) option {
	return func(this *configuration) { this.MaxOpenConnections = int(value) }
}
func (singleton) MaxIdleConnections(value uint16) option {
	return func(this *configuration) { this.MaxIdleConnections = int(value) }
}
func (singleton) IsolationLevel(value sql.IsolationLevel) option {
	return func(this *configuration) { this.DriverConfig.Params["transaction_isolation"] = isolationLevels[value] }
}
func (singleton) Logger(value logger) option {
	return func(this *configuration) {
		if value == nil {
			value = &nop{}
		}

		this.Logger = value
	}
}

func (singleton) apply(options ...option) option {
	return func(this *configuration) {
		for _, item := range Options.defaults(options...) {
			item(this)
		}
	}
}
func (singleton) defaults(options ...option) []option {
	return append([]option{
		Options.TLS(nil),
		Options.Dialer(nil),
		Options.Name("default-mysql-pool"),
		Options.Username("root"),
		Options.Password(""),
		Options.Network("tcp"),
		Options.Address("127.0.0.1:3306"),
		Options.Collation("utf8_unicode_520_ci"),
		Options.ParseTime(true),
		Options.InterpolateParameters(true),
		Options.MultipleStatements(false),
		Options.AllowReadOnly(false),
		Options.ClientFoundRows(true),
		Options.DialTimeout(time.Second * 15),
		Options.ReadTimeout(time.Second * 15),
		Options.WriteTimeout(time.Second * 30),
		Options.MaxConnectionIdleTimeout(time.Hour * 720),
		Options.MaxConnectionLifetime(time.Hour * 720),
		Options.MaxIdleConnections(1024),
		Options.MaxOpenConnections(1024),
		Options.IsolationLevel(sql.LevelReadCommitted),
		Options.Logger(&nop{}),
	}, options...)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type option func(*configuration)
type singleton struct{}
type logger interface {
	Print(...any)
	Printf(string, ...any)
}

var Options singleton

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type nop struct{}

func (*nop) Print(...any)          {}
func (*nop) Printf(string, ...any) {}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var isolationLevels = map[sql.IsolationLevel]string{
	sql.LevelDefault:         "'READ-COMMITTED'",
	sql.LevelReadUncommitted: "'READ-UNCOMMITTED'",
	sql.LevelReadCommitted:   "'READ-COMMITTED'",
	sql.LevelWriteCommitted:  "'WRITE-COMMITTED'",
	sql.LevelRepeatableRead:  "'REPEATABLE-READ'",
	sql.LevelSnapshot:        "'SNAPSHOT'",
	sql.LevelSerializable:    "'SERIALIZABLE'",
	sql.LevelLinearizable:    "'LINEARIZABLE'",
}
