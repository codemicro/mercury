package mercury

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"
)

type Ctx struct {
	tlsConn   *tls.Conn
	request   *request
	response  *response
	callstack []*handler
	// stackPointer points to the current handler in use from callstack
	stackPointer int
}

func newCtx(tlsConn *tls.Conn, callStack []*handler, req *request) *Ctx {
	resp := &response{
		status: StatusSuccess,
		meta:   []byte("text/plain"),
	}

	return &Ctx{
		tlsConn:   tlsConn,
		request:   req,
		response:  resp,
		callstack: callStack,
	}
}

func (ctx *Ctx) SetStatus(status Status) {
	ctx.response.status = status
}

func (ctx *Ctx) SetMeta(meta string) error {
	if len(meta) > 1024 {
		return fmt.Errorf("mercury: meta too long (%d > 1024)", len(meta))
	}
	ctx.response.meta = []byte(meta)
	return nil
}

func (ctx *Ctx) SetBody(body string) {
	ctx.response.content = []byte(body)
}

func (ctx *Ctx) ClearBody() {
	ctx.response.content = nil
}

func (ctx *Ctx) Next() error {
	for {
		if ctx.stackPointer >= len(ctx.callstack) {
			return NewError("Not found", StatusNotFound)
		}
		h := ctx.callstack[ctx.stackPointer]
		ctx.stackPointer += 1
		if doesHandlerMatchPath(ctx.request.pathComponents, h) {
			return h.f(ctx)
		}
	}
}

func (ctx *Ctx) getHandler() *handler {
	return ctx.callstack[ctx.stackPointer-1]
}

func (ctx *Ctx) GetURLParamWithDefault(name, defaultValue string) string {
	h := ctx.getHandler()
	for i, part := range h.pathComponents {
		if !strings.HasPrefix(part, ":") {
			continue
		}
		if part[1:] == name {
			return ctx.request.pathComponents[i]
		}
	}
	return defaultValue
}

func (ctx *Ctx) GetURLParam(name string) string {
	return ctx.GetURLParamWithDefault(name, "")
}

func (ctx *Ctx) GetRawQueryWithDefault(defaultValue string) string {
	if x := ctx.request.URL.RawQuery; x == "" {
		return defaultValue
	} else {
		return x
	}
}

func (ctx *Ctx) GetRawQuery() string {
	return ctx.GetRawQueryWithDefault("")
}

func (ctx *Ctx) GetClientCertificates() []*x509.Certificate {
	return ctx.tlsConn.ConnectionState().PeerCertificates
}
