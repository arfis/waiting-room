package errors

import (
	"encoding/json"
	"errors"
	"net/http"
)

type BodyProvider interface {
	Body() []byte
}

func ParseClientError[T BodyProvider](err error, response *http.Response) (*ApplicationError, error) {
	if response == nil {
		return nil, err
	}

	httpCode := response.StatusCode
	var apiErr T
	if errors.As(err, &apiErr) { // Ensure we use a pointer to apiErr here
		var appError ApplicationError
		body := apiErr.Body()
		if body != nil {
			if jsonErr := json.Unmarshal(body, &appError); jsonErr != nil {
				// If JSON unmarshaling fails, treat the body as a plain string - can be a server error
				return nil, New("", string(body), httpCode, nil)
			}
			appError.HttpCode = httpCode
		}
		return &appError, nil
	}
	return nil, err
}
