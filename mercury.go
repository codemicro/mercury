package mercury

import (
	"crypto/tls"
	"fmt"
	"log"
)

type HandlerFunction func(ctx *Ctx) error

type handler struct {
	f    HandlerFunction
	path string
}

type App struct {
	certificate tls.Certificate
	logger      *log.Logger
	callstack   []*handler
}

func New(conf ...AppConfigFunction) (*App, error) {
	app := &App{
		logger: log.Default(),
	}

	for _, f := range conf {
		if err := f(app); err != nil {
			return nil, err
		}
	}

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
		path: path,
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
			app.log("mercury: error when accepting connection: %v", err)
			continue
		}

		tlsConn, ok := conn.(*tls.Conn)
		if !ok {
			app.log("mercury: got non-TLS connection")
			_ = conn.Close()
			continue
		}

		requestBytes := make([]byte, 1026) // Maximum length request URL + CRLF = 1026 bytes
		if _, err = tlsConn.Read(requestBytes); err != nil {
			app.log("mercury: could not read request: %v", err)
			_ = tlsConn.Close()
			continue
		}

		parsedRequest, err := parseRequest(requestBytes)
		if err != nil {
			// TODO: Pass this to an proper error handler instead of dropping the connection
			app.log("mercury: could not parse request: %v", err)
			_ = tlsConn.Close()
			continue
		}

		content := fmt.Sprintf(
			"Hello world! This page was served from Mercury.\n\nYou requested %s.",
			parsedRequest.URL.String(),
		)

		respBytes, _ := (&response{status: StatusSuccess, meta: []byte("text/plain"), content: []byte(content)}).Encode()

		_, _ = tlsConn.Write(respBytes)
		_ = tlsConn.Close()
	}

	return nil
}
