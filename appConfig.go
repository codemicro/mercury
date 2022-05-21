package mercury

import (
	"crypto/tls"
	"log"
)

type AppConfigFunction func(*App) error

// WithX509KeyPair loads an X509 certificate file and key file from disk.
func WithX509KeyPair(certFile, keyFile string) AppConfigFunction {
	return func(app *App) error {
		x, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
		app.certificate = x
		return nil
	}
}

// WithX509KeyData loads an X509 certificate and key from the provided bytes.
func WithX509KeyData(certPEMBlock, keyPEMBlock []byte) AppConfigFunction {
	return func(app *App) error {
		x, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
		if err != nil {
			return err
		}
		app.certificate = x
		return nil
	}
}

// WithLogger sets the error logger to the one provided.
//
// To disable logging, use this function with a nil logger.
func WithLogger(x *log.Logger) AppConfigFunction {
	return func(app *App) error {
		app.logger = x
		return nil
	}
}