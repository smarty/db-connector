package dbtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/smarty/db-connector/internal"
)

func New(options ...option) (*tls.Config, error) {
	var config configuration
	Options.apply(options...)(&config)

	if !config.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		MinVersion:             config.MinTLSVersion,
		ServerName:             config.ServerName,
		RootCAs:                config.newPool(),
		SessionTicketsDisabled: true,
	}

	sanitizedFilename := sanitize(config.TrustedCAsPEMFile)
	if trustedCAs, err := resolvePEM(config.TrustedCAsPEM, sanitizedFilename); err != nil {
		return nil, err
	} else if ok := tlsConfig.RootCAs.AppendCertsFromPEM(trustedCAs); !ok {
		return nil, fmt.Errorf("unable to parse trusted CA PEM: %w", ErrMalformedPEM)
	}

	// FUTURE: support server certificate(s), RSA/EC private key, and password-protected RSA/EC private key
	return tlsConfig, nil
}
func sanitize(value string) string {
	switch value {
	case "public-ca", "true":
		return ""
	default:
		return value
	}
}
func resolvePEM(source, filename string) ([]byte, error) {
	if len(filename) > 0 {
		if raw, err := ioutil.ReadFile(filename); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrReadPEMFile, err)
		} else {
			return raw, nil
		}
	}

	if len(source) == 0 {
		return nil, nil
	}

	if resolved := internal.TryReadValue(source); len(resolved) > 0 {
		return []byte(resolved), nil
	}

	return nil, ErrReadPEMFile
}

func (this configuration) newPool() *x509.CertPool {
	if !this.TrustSystemRootCAs {
		return x509.NewCertPool()
	}

	pool, _ := x509.SystemCertPool()
	return pool
}

type configuration struct {
	Enabled            bool
	TrustSystemRootCAs bool
	ServerName         string
	TrustedCAsPEM      string
	TrustedCAsPEMFile  string
	MinTLSVersion      uint16
}

func (singleton) Enabled(value bool) option {
	return func(this *configuration) { this.Enabled = value }
}
func (singleton) TrustSystemRootCAs(value bool) option {
	return func(this *configuration) { this.TrustSystemRootCAs = value }
}
func (singleton) ServerName(value string) option {
	return func(this *configuration) { this.ServerName = value }
}
func (singleton) TrustedCAsPEM(value string) option {
	return func(this *configuration) { this.TrustedCAsPEM = value }
}
func (singleton) TrustedCAsPEMFile(value string) option {
	return func(this *configuration) { this.TrustedCAsPEMFile = value }
}
func (singleton) MinTLSVersion(value uint16) option {
	return func(this *configuration) { this.MinTLSVersion = value }
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
		Options.Enabled(true),
		Options.TrustSystemRootCAs(true),
		Options.ServerName(""),
		Options.TrustedCAsPEM(""),
		Options.TrustedCAsPEMFile(""),
		Options.MinTLSVersion(tls.VersionTLS12),
	}, options...)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type option func(*configuration)
type singleton struct{}

var Options singleton
