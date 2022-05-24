package mercury

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const startupLogo = " _____                         \n|     |___ ___ ___ _ _ ___ _ _ \n| | | | -_|  _|  _| | |  _| | |\n|_|_|_|___|_| |___|___|_| |_  |\n                          |___|"

type HandlerFunction func(ctx *Ctx) error

type handler struct {
	f              HandlerFunction
	pathComponents []string
	isMiddleware   bool
}

type App struct {
	debug                 bool
	certificate           tls.Certificate
	logger                *log.Logger
	callstack             []*handler
	errorHandler          ErrorHandlerFunction
	readTimeout           time.Duration
	writeTimeout          time.Duration
	disableStartupMessage bool
	serverName            string

	mu               *sync.Mutex
	isListenerClosed bool
	listener         net.Listener
}

func New(conf ...AppConfigFunction) (*App, error) {
	app := &App{
		logger:       log.Default(),
		errorHandler: DefaultErrorHandler,
		mu:           new(sync.Mutex),
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
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	app.callstack = append(app.callstack, &handler{
		f:              handlerFunction,
		pathComponents: splitPath(strings.ToLower(path)),
	})
}

func (app *App) UseOnPath(path string, hf HandlerFunction) {
	app.callstack = append(app.callstack, &handler{
		f:              hf,
		pathComponents: splitPath(path),
		isMiddleware:   true,
	})
}

func (app *App) Use(hf HandlerFunction) {
	app.UseOnPath("/", hf)
}

func (app *App) Listen(addr string) error {
	if !app.disableStartupMessage {
		_, _ = fmt.Fprintln(os.Stderr, startupLogo)
		_, _ = fmt.Fprint(os.Stderr, "Listening on gemini://")
		if strings.HasPrefix(addr, ":") {
			_, _ = fmt.Fprintln(os.Stderr, "0.0.0.0"+addr)
		} else {
			_, _ = fmt.Fprintln(os.Stderr, addr)
		}
	}

	app.mu.Lock()
	listener, err := tls.Listen("tcp", addr, &tls.Config{
		Certificates: []tls.Certificate{app.certificate},
		ServerName:   app.serverName,
		ClientAuth:   tls.RequestClientCert,
		MinVersion:   tls.VersionTLS12,
	})
	app.listener = listener
	app.mu.Unlock()
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			app.mu.Lock()
			closed := app.isListenerClosed
			app.mu.Unlock()

			if closed {
				break
			}

			app.log("error when accepting connection: %v", err)
			continue
		}

		tlsConn, ok := conn.(*tls.Conn)
		if !ok {
			app.log("got non-TLS connection")
			_ = conn.Close()
			continue
		}

		if app.readTimeout != 0 {
			_ = tlsConn.SetReadDeadline(time.Now().Add(app.readTimeout))
		}

		if app.writeTimeout != 0 {
			_ = tlsConn.SetWriteDeadline(time.Now().Add(app.writeTimeout))
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

		ctx := newCtx(tlsConn, app.callstack, parsedRequest)

		if err := ctx.Next(); err != nil {
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

		app.writeToConn(tlsConn, respBytes)
		_ = tlsConn.Close()
	}

	return nil
}

// Shutdown shuts down the app if it's listening
func (app *App) Shutdown() error {
	app.mu.Lock()
	defer app.mu.Unlock()

	if app.listener == nil {
		return nil
	}

	if err := app.listener.Close(); err != nil {
		return err
	}
	app.isListenerClosed = true

	return nil
}

// callErrorHandler will always close the request if no ctx is provided, else
// the connection may or may not be closed.
func (app *App) callErrorHandler(conn *tls.Conn, ctx *Ctx, err error) (connClosed bool) {
	ctxWasProvided := ctx != nil
	if !ctxWasProvided {
		ctx = newCtx(conn, nil, nil)
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
		app.writeToConn(conn, respBytes)
		_ = conn.Close()
		return true
	}
	return false
}

func (app *App) writeToConn(tls *tls.Conn, content []byte) {
	if app.debug {
		app.logger.Printf("sending response with content %#v", string(content))
	}
	_, _ = tls.Write(content)
}
