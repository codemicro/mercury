package mercury

import (
	"bytes"
	"errors"
	"net/url"
	"strconv"
)

var (
	errorRequestURLTooLong     = errors.New("mercury: request URL length is greater than 1024 characters")
	errorRequestURLHasNoScheme = errors.New("mercury: request URL has no scheme")
	errorMalformedRequest      = errors.New("mercury: malformed request")
)

type request struct {
	*url.URL
}

func parseRequest(x []byte) (*request, error) {
	if bytes.HasPrefix(x, []byte("\uFFFF")) {
		return nil, errorMalformedRequest
	}

	crlfLocation := bytes.Index(x, []byte{'\r', '\n'})
	if crlfLocation == -1 {
		return nil, errorMalformedRequest
	}

	rawURL := string(x[:crlfLocation])

	if len(rawURL) > 1024 {
		return nil, errorRequestURLTooLong
	}

	parsed, err := url.Parse(string(x[:crlfLocation]))
	if err != nil {
		return nil, err
	}

	if parsed.Scheme == "" {
		return nil, errorRequestURLHasNoScheme
	}

	return &request{
		URL: parsed,
	}, nil
}

var (
	errorResponseMetaTooLong = errors.New("mercury: meta too long")
	errorImpossibleResponse  = errors.New("mercury: impossible response")
)

type response struct {
	status  Status
	meta    []byte
	content []byte
}

func (r *response) Encode() ([]byte, error) {
	if len(r.meta) > 1024 {
		return nil, errorResponseMetaTooLong
	}

	if bytes.HasPrefix(r.meta, []byte("\uFFFF")) {
		return nil, errorImpossibleResponse
	}

	if r.status/10 != 2 { // 2 denotes the success range of codes
		if len(r.content) != 0 {
			return nil, errorImpossibleResponse
		}
	}

	var b []byte
	b = strconv.AppendInt(b, int64(r.status), 10)
	b = append(b, ' ')
	b = append(b, r.meta...)
	b = append(b, '\r', '\n')
	b = append(b, r.content...)

	return b, nil
}
