package mercury

import (
	"crypto/tls"
	"log"
	"strings"
)

type HandlerFunction func(ctx *Ctx) error

type handler struct {
	f    HandlerFunction
	path string
}

type App struct {
	certificate  tls.Certificate
	logger       *log.Logger
	callstack    []*handler
	errorHandler ErrorHandlerFunction
}

func New(conf ...AppConfigFunction) (*App, error) {
	app := &App{
		logger:       log.Default(),
		errorHandler: DefaultErrorHandler,
	}

	for _, f := range conf {
		if err := f(app); err != nil {
			return nil, err
		}
	}

	app.logger.SetPrefix("mercury: ")

	return app, nil
}

func (app *App) log(format string, args ...any) {
	if app.logger != nil {
		app.logger.Printf(format, args...)
	}
}

// Add registers a handler function to be used to serve requests to a specific
// URL.
func (app *App) Add(path string, handlerFunction HandlerFunction) {
	app.callstack = append(app.callstack, &handler{
		f:    handlerFunction,
		path: strings.ToLower(path),
	})
}

func (app *App) Listen(addr string) error {
	listener, err := tls.Listen("tcp", addr, &tls.Config{
		Certificates: []tls.Certificate{app.certificate},
		ServerName:   "", // TODO(codemicro)
		ClientAuth:   tls.RequestClientCert,
		MinVersion:   tls.VersionTLS12,
	})
	if err != nil {
		return err
	}
	// TODO: You can call this and it'll stop any blocked calls to listener.Accept
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			app.log("error when accepting connection: %v", err)
			continue
		}

		tlsConn, ok := conn.(*tls.Conn)
		if !ok {
			app.log("got non-TLS connection")
			_ = conn.Close()
			continue
		}

		requestBytes := make([]byte, 1026) // Maximum length request URL + CRLF = 1026 bytes
		if _, err = tlsConn.Read(requestBytes); err != nil {
			app.log("could not read request: %v", err)
			_ = tlsConn.Close()
			continue
		}

		parsedRequest, err := parseRequest(requestBytes)
		if err != nil {
			_ = app.callErrorHandler(tlsConn, nil, err)
			continue // when ctx == nil in callErrorHandler, the connection is always closed.
		}

		handlerForPath := app.getHandlerForPath(parsedRequest.URL.Path)
		if handlerForPath == nil {
			_ = app.callErrorHandler(tlsConn, nil, NewError("Not found", StatusNotFound))
			continue // when ctx == nil in callErrorHandler, the connection is always closed.
		}

		ctx := newCtx(parsedRequest)

		if err := handlerForPath.f(ctx); err != nil {
			if requestClosed := app.callErrorHandler(tlsConn, ctx, err); requestClosed {
				continue
			}
		}

		respBytes, err := ctx.response.Encode()
		if err != nil {
			if requestClosed := app.callErrorHandler(tlsConn, ctx, err); requestClosed {
				continue
			}
		}
		_, _ = tlsConn.Write(respBytes)
		_ = tlsConn.Close()
	}

	return nil
}

// callErrorHandler will always close the request if no ctx is provided, else
// the connection may or may not be closed.
func (app *App) callErrorHandler(conn *tls.Conn, ctx *Ctx, err error) (connClosed bool) {
	ctxWasProvided := ctx != nil
	if !ctxWasProvided {
		ctx = newCtx(nil)
	}

	if err2 := app.errorHandler(ctx, err); err2 != nil {
		app.log("error handler returned error '%v' when handling error '%v'", err2, err)
		_ = conn.Close()
		return true
	}

	if !ctxWasProvided {
		respBytes, err := ctx.response.Encode()
		if err != nil {
			_ = conn.Close()
			return true
		}
		_, _ = conn.Write(respBytes)
		_ = conn.Close()
		return true
	}
	return false
}
