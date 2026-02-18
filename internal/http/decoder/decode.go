package decoder

import (
	"encoding/json"
	stderrs "errors"
	"io"
	"mime"
	"net/http"
	"strings"

	httpErrors "admin.com/admin-api/internal/http/errors"
	"admin.com/admin-api/internal/http/response"
)

const DefaultMaxBodyBytes int64 = 1 << 20 // 1 MiB

type DecodeError struct {
	Code    string
	Message string
}

func (e DecodeError) Error() string {
	return e.Message
}

func DecodeBody(w http.ResponseWriter, r *http.Request, target any) error {
	if err := validateJSONContentType(r); err != nil {
		return err
	}

	r.Body = http.MaxBytesReader(w, r.Body, DefaultMaxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return mapDecodeError(err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return DecodeError{
			Code:    httpErrors.MultipleJSON.Code,
			Message: httpErrors.MultipleJSON.Message,
		}
	}

	return nil
}

func WriteDecodeError(w http.ResponseWriter, err error) {
	var decodeErr DecodeError
	if stderrs.As(err, &decodeErr) {
		response.WriteErrorWithCode(w, http.StatusBadRequest, decodeErr.Code, decodeErr.Message)
		return
	}

	response.WriteErrorWithCode(w, http.StatusBadRequest, httpErrors.InvalidPayload.Code, httpErrors.InvalidPayload.Message)
}

func validateJSONContentType(r *http.Request) error {
	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	if contentType == "" {
		return DecodeError{
			Code:    httpErrors.InvalidContentType.Code,
			Message: httpErrors.InvalidContentType.Message,
		}
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "application/json" {
		return DecodeError{
			Code:    httpErrors.InvalidContentType.Code,
			Message: httpErrors.InvalidContentType.Message,
		}
	}

	return nil
}

func mapDecodeError(err error) error {
	var syntaxError *json.SyntaxError
	var unmarshalTypeError *json.UnmarshalTypeError
	var maxBytesError *http.MaxBytesError

	switch {
	case stderrs.Is(err, io.EOF):
		return DecodeError{
			Code:    httpErrors.InvalidBody.Code,
			Message: httpErrors.InvalidBody.Message,
		}
	case stderrs.Is(err, io.ErrUnexpectedEOF), stderrs.As(err, &syntaxError):
		return DecodeError{
			Code:    httpErrors.MalformedJSON.Code,
			Message: httpErrors.MalformedJSON.Message,
		}
	case stderrs.As(err, &unmarshalTypeError):
		return DecodeError{
			Code:    httpErrors.InvalidFieldType.Code,
			Message: httpErrors.InvalidFieldType.Message,
		}
	case stderrs.As(err, &maxBytesError):
		return DecodeError{
			Code:    httpErrors.BodyTooLarge.Code,
			Message: httpErrors.BodyTooLarge.Message,
		}
	default:
		return DecodeError{
			Code:    httpErrors.InvalidPayload.Code,
			Message: httpErrors.InvalidPayload.Message,
		}
	}
}
