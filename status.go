package mercury

type Status int

const (
	StatusInput          Status = 10
	StatusSensitiveInput Status = 11

	StatusSuccess Status = 20

	StatusTemporaryRedirect Status = 30
	StatusPermanentRedirect Status = 31

	StatusTemporaryFailure  Status = 40
	StatusServerUnavailable Status = 41
	StatusCGIError          Status = 42
	StatusProxyError        Status = 43
	StatusSlowDown          Status = 44

	StatusPermanentFailure    Status = 50
	StatusNotFound            Status = 51
	StatusGone                Status = 52
	StatusProxyRequestRefused Status = 53
	StatusBadRequest          Status = 59

	StatusClientCertificateRequired Status = 60
	StatusCertificateNotAuthorised  Status = 61
	StatusCertificateNotValid       Status = 62
)
