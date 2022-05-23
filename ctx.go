package mercury

import (
	"fmt"
)

type Ctx struct {
	request   *request
	response  *response
	callstack []*handler
	// stackPointer points to the current handler in use from callstack
	stackPointer int
}

func newCtx(callStack []*handler, req *request) *Ctx {
	resp := &response{
		status: StatusSuccess,
		meta:   []byte("text/plain"),
	}

	return &Ctx{
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
		if doesHandlerMatchPath(ctx.request.pathComponents, h) {
			// Some context functions use the stackPointer variable to get a
			// reference to the current *handler. Incrementing stackPointer
			// after calling the handler function ensures the stackPointer
			// always contains the correct value to point to the current
			// handler when the handler function is running.
			e := h.f(ctx)
			ctx.stackPointer += 1
			return e
		}
		ctx.stackPointer += 1
	}
}
