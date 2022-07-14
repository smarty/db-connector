package mysql

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/smarty/db-connector/internal"
)

func New(options ...option) (*sql.DB, error) {
	var config configuration
	Options.apply(options...)(&config)

	config.Logger.Printf("[INFO] Establishing MySQL database handle [%s] with data source settings: [%s]", config.Name, config.render(true))
	if handle, err := sql.Open("mysql", config.render(false)); err != nil {
		return nil, err
	} else {
		handle.SetConnMaxIdleTime(config.MaxConnectionIdleTimeout)
		handle.SetConnMaxLifetime(config.MaxConnectionLifetime)
		handle.SetMaxOpenConns(config.MaxOpenConnections)
		handle.SetMaxIdleConns(config.MaxIdleConnections)
		return handle, nil
	}
}

type configuration struct {
	TLSConfig                *tls.Config
	TLSRegistration          func(string, *tls.Config) error
	tlsName                  string
	Name                     string
	Username                 string
	Password                 string
	Network                  string
	Address                  string
	Schema                   string
	Collation                string
	ParseTime                bool
	InterpolateParameters    bool
	MultipleStatements       bool
	AllowReadOnly            bool
	ClientFoundRows          bool
	DialTimeout              time.Duration
	ReadTimeout              time.Duration
	WriteTimeout             time.Duration
	MaxConnectionIdleTimeout time.Duration
	MaxConnectionLifetime    time.Duration
	MaxIdleConnections       int
	MaxOpenConnections       int
	IsolationLevel           sql.IsolationLevel
	Logger                   logger
}

func (this *configuration) uniqueTLSName() string {
	if len(this.tlsName) > 0 {
		return this.tlsName
	}

	if this.TLSConfig == nil {
		return ""
	}

	this.tlsName = fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	return this.tlsName
}

func (this *configuration) render(redact bool) string {
	builder := &strings.Builder{}

	var (
		username = internal.TryReadValue(this.Username)
		password = internal.TryReadValue(this.Password)
		tlsName  = this.uniqueTLSName()
	)

	if redact && len(password) > 0 {
		password = "REDACTED"
	}

	if len(username) > 0 && len(password) > 0 && redact {
		_, _ = fmt.Fprintf(builder, "%s:%s@", username, password)
	} else if len(username) > 0 {
		_, _ = fmt.Fprintf(builder, "%s@", username)
	}

	_, _ = fmt.Fprintf(builder, "%s(%s)", this.Network, internal.TryReadValue(this.Address))
	_, _ = fmt.Fprintf(builder, "/%s", this.Schema)

	settings := url.Values{
		"collation":             []string{this.Collation},
		"parseTime":             []string{fmt.Sprintf("%v", this.ParseTime)},
		"interpolateParams":     []string{fmt.Sprintf("%v", this.InterpolateParameters)},
		"multiStatements":       []string{fmt.Sprintf("%v", this.MultipleStatements)},
		"rejectReadOnly":        []string{fmt.Sprintf("%v", !this.AllowReadOnly)},
		"clientFoundRows":       []string{fmt.Sprintf("%v", this.ClientFoundRows)},
		"timeout":               []string{fmt.Sprintf("%v", this.DialTimeout)},
		"readTimeout":           []string{fmt.Sprintf("%v", this.ReadTimeout)},
		"writeTimeout":          []string{fmt.Sprintf("%v", this.WriteTimeout)},
		"transaction_isolation": []string{fmt.Sprintf("'%s'", isolationLevels[this.IsolationLevel])},
	}

	if len(tlsName) > 0 {
		settings["tls"] = []string{tlsName}
		if this.TLSRegistration != nil {
			_ = this.TLSRegistration(tlsName, this.TLSConfig)
		}
	}

	_, _ = fmt.Fprintf(builder, "?%s", settings.Encode())
	return builder.String()
}
func (singleton) TLS(value *tls.Config, registration func(string, *tls.Config) error) option {
	return func(this *configuration) { this.TLSConfig = value; this.TLSRegistration = registration }
}
func (singleton) Name(value string) option {
	return func(this *configuration) { this.Name = value }
}
func (singleton) Username(value string) option {
	return func(this *configuration) { this.Username = value }
}
func (singleton) Password(value string) option {
	return func(this *configuration) { this.Password = value }
}
func (singleton) Network(value string) option {
	return func(this *configuration) { this.Network = value }
}
func (singleton) Address(value string) option {
	return func(this *configuration) { this.Address = value }
}
func (singleton) Schema(value string) option {
	return func(this *configuration) { this.Schema = value }
}
func (singleton) Collation(value string) option {
	return func(this *configuration) { this.Collation = value }
}
func (singleton) ParseTime(value bool) option {
	return func(this *configuration) { this.ParseTime = value }
}
func (singleton) InterpolateParameters(value bool) option {
	return func(this *configuration) { this.InterpolateParameters = value }
}
func (singleton) MultipleStatements(value bool) option {
	return func(this *configuration) { this.MultipleStatements = value }
}
func (singleton) AllowReadOnly(value bool) option {
	return func(this *configuration) { this.AllowReadOnly = value }
}
func (singleton) ClientFoundRows(value bool) option {
	return func(this *configuration) { this.ClientFoundRows = value }
}
func (singleton) DialTimeout(value time.Duration) option {
	return func(this *configuration) { this.DialTimeout = value }
}
func (singleton) ReadTimeout(value time.Duration) option {
	return func(this *configuration) { this.ReadTimeout = value }
}
func (singleton) WriteTimeout(value time.Duration) option {
	return func(this *configuration) { this.WriteTimeout = value }
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
	return func(this *configuration) { this.IsolationLevel = value }
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
		Options.TLS(nil, nil),
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
type logger interface{ Printf(string, ...any) }

var Options singleton

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type nop struct{}

func (*nop) Printf(string, ...any) {}
