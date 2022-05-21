package mercury

type Ctx struct {
	request  *request
	response *response
}

func (ctx *Ctx) SetStatus(status Status) {
	ctx.response.status = status
}

func (ctx *Ctx) SetMeta(meta string) {
	ctx.response.meta = []byte(meta)
}

func (ctx *Ctx) SetBody(body string) {
	ctx.response.content = []byte(body)
}
