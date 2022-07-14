package mysql

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func New(options ...option) (*sql.DB, error) {
	var config configuration
	Options.apply(options...)(&config)

	if handle, err := sql.Open("mysql", config.String(false)); err != nil {
		return nil, err
	} else {
		config.Logger.Printf("[INFO] Creating MySQL database handle [%s] with data source settings: [%s]", config.Name, config.String(true))
		return handle, nil
	}
}

type configuration struct {
	TLSConfig             *tls.Config
	TLSRegistration       func(string, *tls.Config) error
	Name                  string
	Username              string
	Password              string
	Protocol              string
	Address               string
	Schema                string
	Collation             string
	ParseTime             bool
	InterpolateParameters bool
	AllowReadOnly         bool
	ClientFoundRows       bool
	DialTimeout           time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	IsolationLevel        sql.IsolationLevel
	Logger                logger
}

func (this *configuration) UniqueTLSName() string {
	if this.TLSConfig == nil {
		return ""
	}

	return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
}

func (this *configuration) String(redact bool) string {
	builder := &strings.Builder{}

	var (
		username = tryReadValue(this.Username)
		password = tryReadValue(this.Password)
		tlsName  = this.UniqueTLSName()
	)

	if redact {
		password = "REDACTED"
	}

	if len(username) > 0 && len(password) > 0 && redact {
		_, _ = fmt.Fprintf(builder, "%s:%s@", username, password)
	} else if len(username) > 0 {
		_, _ = fmt.Fprintf(builder, "%s@", username)
	}

	_, _ = fmt.Fprintf(builder, "%s(%s)", this.Protocol, tryReadValue(this.Address))
	_, _ = fmt.Fprintf(builder, "/%s", this.Schema)

	_, _ = fmt.Fprintf(builder, "?collation=%s", this.Collation)
	_, _ = fmt.Fprintf(builder, "&parseTime=%v", this.ParseTime)
	_, _ = fmt.Fprintf(builder, "&interpolateParams=%v", this.InterpolateParameters)
	_, _ = fmt.Fprintf(builder, "&rejectReadOnly=%v", !this.AllowReadOnly)
	_, _ = fmt.Fprintf(builder, "&clientFoundRows=%v", this.ClientFoundRows)
	_, _ = fmt.Fprintf(builder, "&timeout=%s", this.DialTimeout)
	_, _ = fmt.Fprintf(builder, "&readTimeout=%s", this.ReadTimeout)
	_, _ = fmt.Fprintf(builder, "&writeTimeout=%s", this.WriteTimeout)
	_, _ = fmt.Fprintf(builder, "&transaction_isolation='%s'", isolationLevels[this.IsolationLevel])

	if len(tlsName) > 0 {
		_, _ = fmt.Fprintf(builder, "&tls=%s", tlsName)

		if this.TLSRegistration != nil {
			_ = this.TLSRegistration(tlsName, this.TLSConfig)
		}
	}

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
func (singleton) Protocol(value string) option {
	return func(this *configuration) { this.Protocol = value }
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
		Options.Protocol("tcp"),
		Options.Address("127.0.0.1"),
		Options.Collation("utf8_unicode_520_ci"),
		Options.ParseTime(true),
		Options.InterpolateParameters(true),
		Options.AllowReadOnly(false),
		Options.ClientFoundRows(true),
		Options.DialTimeout(time.Second * 15),
		Options.ReadTimeout(time.Second * 15),
		Options.WriteTimeout(time.Second * 30),
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
