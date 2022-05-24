package mercury

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
)

type Ctx struct {
	tlsConn *tls.Conn

	request  *request
	response *response

	bodyBuilder *strings.Builder

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
		return fmt.Errorf("mercury: meta too long (len %d > 1024)", len(meta))
	}
	ctx.response.meta = []byte(meta)
	return nil
}

// SetBody sets the response body to a single string. This will be overridden
// if (*ctx).SetBodyBuilder is used.
func (ctx *Ctx) SetBody(body string) {
	ctx.response.content = []byte(body)
}

// SetBodyFromFile reads a file with the specified name and uses its contents
// as the response body. This will be overridden if (*ctx).SetBodyBuilder is
// used.
func (ctx *Ctx) SetBodyFromFile(filename string) error {
	cont, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	ctx.response.content = cont
	return nil
}

// SetBodyBuilder allows a strings.Builder to be used to create the response
// body. This will overwrite any other calls made to set the response body.
//
// To undo this, call SetBodyBuilder with a nil *strings.Builder then call
// other body-setting functions as normal.
func (ctx *Ctx) SetBodyBuilder(sb *strings.Builder) {
	ctx.bodyBuilder = sb
}

func (ctx *Ctx) GetBody() *[]byte {
	return &ctx.response.content
}

func (ctx *Ctx) GetMeta() *[]byte {
	return &ctx.response.meta
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
			e := h.f(ctx)
			if ctx.bodyBuilder != nil {
				ctx.SetBody(ctx.bodyBuilder.String())
			}
			return e
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
			return splitPath(ctx.request.URL.Path)[i]
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

func (ctx *Ctx) GetRemoteAddress() net.Addr {
	return ctx.tlsConn.RemoteAddr()
}

func (ctx *Ctx) GetRequestURL() *url.URL {
	return ctx.request.URL
}
