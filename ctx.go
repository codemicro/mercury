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
		ctx.stackPointer += 1
		if doesHandlerMatchPath(ctx.request.pathComponents, h) {
			return h.f(ctx)
		}
	}
}
