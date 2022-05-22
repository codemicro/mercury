package mercury

type ErrorHandlerFunction func(ctx *Ctx, err error) error

type Error struct {
	Message string
	Status  Status
}

func NewError(message string, status Status) error {
	return &Error{
		Message: message,
		Status:  status,
	}
}

func (e Error) Error() string {
	return e.Message
}

func DefaultErrorHandler(ctx *Ctx, err error) error {
	ctx.ClearBody()
	if e, ok := err.(*Error); ok {
		ctx.SetStatus(e.Status)
		return ctx.SetMeta(e.Message)
	} else {
		ctx.SetStatus(StatusTemporaryFailure)
		return ctx.SetMeta("Internal server error")
	}
}
