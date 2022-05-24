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

// SetStatus sets the status code of the response.
func (ctx *Ctx) SetStatus(status Status) {
	ctx.response.status = status
}

// SetMeta sets the meta field of the response. The meaning of this field
// depends on the status code used in the response. By default, this is set to
// "text/plain", which is suitable for use with a 20 status code.
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

// GetBody returns a pointer to the bytearray containing the response content.
func (ctx *Ctx) GetBody() *[]byte {
	return &ctx.response.content
}

// GetMeta returns a pointer to the bytearray containing the response meta
// field.
func (ctx *Ctx) GetMeta() *[]byte {
	return &ctx.response.meta
}

// ClearBody empties the request body.
func (ctx *Ctx) ClearBody() {
	ctx.response.content = nil
}

// Next executes the next handler in the callstack, or returns an error if one
// doesn't exist.
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

// GetURLParamWithDefault behaves identically to GetURLParam, except it will
// return a default value instead of an empty string if the key cannot be
// found.
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

// GetURLParam retrieves the contents of the named URL parameter in the request
// URL. The parameter key is case-insensitive.
//
// For example, if you registered a handler with the path /hello/:name and
// someone requested /hello/Abi, the result of calling GetURLParam("name")
// would be "Abi".
//
// If the key is not recognised, an empty string is returned.
func (ctx *Ctx) GetURLParam(name string) string {
	return ctx.GetURLParamWithDefault(name, "")
}

// GetRawQueryWithDefault behaves identically to GetRawQuery, except it returns
// the specified default value instead of an empty string if no query string
// was provided.
func (ctx *Ctx) GetRawQueryWithDefault(defaultValue string) string {
	if x := ctx.request.URL.RawQuery; x == "" {
		return defaultValue
	} else {
		return x
	}
}

// GetRawQuery returns the raw query string from the request URL, returning an
// empty string if there isn't one provided.
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
