package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// NoResponse tells the Respond function to not respond to the request. In these
// cases the app layer code has already done so.
type NoResponse struct{}

// NewNoResponse constructs a no reponse value.
func NewNoResponse() NoResponse {
	return NoResponse{}
}

// Encode implements the Encoder interface
func (NoResponse) Encode() ([]byte, string, error) {
	return []byte("No Response from omer again and again"), "", nil
}

type HttpStatus interface {
	HTTPStatus() int
}

func Respond(cxt context.Context, w http.ResponseWriter, dataModel Encoder) error {
	// if i return no response here we handle it; by returning nil
	_, ok := dataModel.(NoResponse)
	if ok {
		data, _, _ := dataModel.Encode()
		w.Write(data)
		return nil
	}

	// if context has been canceled, it mean tne client is no longer waiting for response
	if err := cxt.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return errors.New("client disconnected, do not send response")
		}
	}
	var statusCode = http.StatusOK

	switch v := dataModel.(type) {
	case HttpStatus:
		statusCode = v.HTTPStatus()
	case error:
		statusCode = http.StatusInternalServerError
	default:
		if dataModel == nil {
			statusCode = http.StatusNoContent
		}
	}
	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	data, contentType, err := dataModel.Encode()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("respond: encode: %w", err)
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(statusCode)
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("respond: write: %w", err)
	}
	return nil
}
