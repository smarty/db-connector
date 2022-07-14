package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Connector interface {
	Open() (*sql.DB, error)
}
type connector struct {
	dsn      string
	redacted string
	config   configuration
	logger   logger
}

func New(options ...option) Connector {
	var config configuration
	Options.apply(options...)(&config)
	return &connector{
		dsn:      renderDSN(config, false),
		redacted: renderDSN(config, true),
		logger:   config.Logger,
	}
}

func (this *connector) Open() (*sql.DB, error) {
	handle, err := sql.Open("mysql", this.dsn)
	if err != nil {
		return nil, err
	}

	this.logger.Printf("[INFO] Established MySQL database handle [%s] with data source settings: [%s]", this.config.Name, this.redacted)
	handle.SetConnMaxIdleTime(this.config.MaxConnectionIdleTimeout)
	handle.SetConnMaxLifetime(this.config.MaxConnectionLifetime)
	handle.SetMaxOpenConns(this.config.MaxOpenConnections)
	handle.SetMaxIdleConns(this.config.MaxIdleConnections)
	return handle, nil
}
func renderDSN(config configuration, redact bool) string {
	builder := &strings.Builder{}

	var (
		username = tryReadValue(config.Username)
		password = tryReadValue(config.Password)
		tlsName  = uniqueTLSName(config)
	)

	if redact {
		password = "REDACTED"
	}

	if len(username) > 0 && len(password) > 0 && redact {
		_, _ = fmt.Fprintf(builder, "%s:%s@", username, password)
	} else if len(username) > 0 {
		_, _ = fmt.Fprintf(builder, "%s@", username)
	}

	_, _ = fmt.Fprintf(builder, "%s(%s)", config.Protocol, tryReadValue(config.Address))
	_, _ = fmt.Fprintf(builder, "/%s", config.Schema)

	_, _ = fmt.Fprintf(builder, "?collation=%s", config.Collation)
	_, _ = fmt.Fprintf(builder, "&parseTime=%v", config.ParseTime)
	_, _ = fmt.Fprintf(builder, "&interpolateParams=%v", config.InterpolateParameters)
	_, _ = fmt.Fprintf(builder, "&rejectReadOnly=%v", !config.AllowReadOnly)
	_, _ = fmt.Fprintf(builder, "&clientFoundRows=%v", config.ClientFoundRows)
	_, _ = fmt.Fprintf(builder, "&timeout=%s", config.DialTimeout)
	_, _ = fmt.Fprintf(builder, "&readTimeout=%s", config.ReadTimeout)
	_, _ = fmt.Fprintf(builder, "&writeTimeout=%s", config.WriteTimeout)
	_, _ = fmt.Fprintf(builder, "&transaction_isolation='%s'", isolationLevels[config.IsolationLevel])

	if len(tlsName) > 0 {
		_, _ = fmt.Fprintf(builder, "&tls=%s", tlsName)

		if config.TLSRegistration != nil {
			_ = config.TLSRegistration(tlsName, config.TLSConfig)
		}
	}

	return builder.String()
}
func uniqueTLSName(tlsConfig any) string {
	if tlsConfig == nil {
		return ""
	}

	return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
}
