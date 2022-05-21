package mercury

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
