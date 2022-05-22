package mercury

import (
	"fmt"
)

type Ctx struct {
	request  *request
	response *response
}

func newCtx(req *request) *Ctx {
	resp := &response{
		status: StatusSuccess,
		meta:   []byte("text/plain"),
	}

	return &Ctx{
		request:  req,
		response: resp,
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
