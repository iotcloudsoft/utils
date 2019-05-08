package xerror

import (
	"context"
	"encoding/json"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
)

const contentType = "application/json; charset=utf-8"

// ErrorEncoder writes the error to the ResponseWriter, by default a content
// type of application/json, a body of json with key "error" and the value
// error.Error(), and a status code of 500. If the error implements Headerer,
// the provided headers will be applied to the response. If the error
// implements json.Marshaler, and the marshaling succeeds, the JSON encoded
// form of the error will be used. If the error implements StatusCoder, the
// provided StatusCode will be used instead of 500.
func HttpEncoder(_ context.Context, err error, w http.ResponseWriter) {
	var (
		ec   = -1
		code = http.StatusInternalServerError
	)

	// extract error code
	if ce, ok := err.(CodeError); ok {
		ec = ce.Code()
	} else if sc, ok := err.(httptransport.StatusCoder); ok {
		code = sc.StatusCode()
		ec = code
	}

	// write header
	w.Header().Set("Content-Type", contentType)
	if headerer, ok := err.(httptransport.Headerer); ok {
		for k := range headerer.Headers() {
			w.Header().Set(k, headerer.Headers().Get(k))
		}
	}
	w.WriteHeader(code)

	// write body
	ew := errorWrapper{}
	ew.Error.Code = ec
	ew.Error.Message = err.Error()
	body, _ := json.Marshal(ew)
	if marshaler, ok := err.(json.Marshaler); ok {
		if jsonBody, marshalErr := marshaler.MarshalJSON(); marshalErr == nil {
			body = jsonBody
		}
	}
	w.Write(body)
}

type errorWrapper struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
