package mercury

import (
	"crypto/tls"
	"errors"
	"log"
	"time"
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

// WithErrorHandler sets the error handler that's used by the app.
//
// The error handler must not make use of request variables. The
// error handler must not call (*ctx).Next()
func WithErrorHandler(eh ErrorHandlerFunction) AppConfigFunction {
	return func(app *App) error {
		if eh == nil {
			return errors.New("mercury: no error handler provided")
		}
		app.errorHandler = eh
		return nil
	}
}

// WithReadTimeout sets the read timeout for any incoming connections.
//
// Setting this value to zero disables read timeouts.
func WithReadTimeout(x time.Duration) AppConfigFunction {
	return func(app *App) error {
		if x < 0 {
			return errors.New("mercury: cannot have negative read timeout")
		}
		app.readTimeout = x
		return nil
	}
}

// WithWriteTimeout sets the write timeout for any incoming connections.
//
// Setting this value to zero disables write timeouts.
func WithWriteTimeout(x time.Duration) AppConfigFunction {
	return func(app *App) error {
		if x < 0 {
			return errors.New("mercury: cannot have negative write timeout")
		}
		app.writeTimeout = x
		return nil
	}
}

func WithDebugModeEnabled() AppConfigFunction {
	return func(app *App) error {
		app.debug = true
		return nil
	}
}

// WithDisableStartupMessage will disable the startup message printed to
// os.Stderr on server start.
func WithDisableStartupMessage() AppConfigFunction {
	return func(app *App) error {
		app.disableStartupMessage = true
		return nil
	}
}

// WithServerName sets the server name used as part of the TLS configuration.
// This can be left blank.
func WithServerName(name string) AppConfigFunction {
	return func(app *App) error {
		app.serverName = name
		return nil
	}
}
