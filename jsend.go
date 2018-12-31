package aclsrv

import (
	"encoding/json"
	"net/http"
)

const (
	JSendSuccess = "success"
	JSendFail    = "fail"
	JSendError   = "error"
)

type JSend struct {
	Status           string          `json:"status"`
	Data             json.RawMessage `json:"data,omitempty"`
	Message          string          `json:"message,omitempty"`
	HTTPCode         int             `json:"http_code,omitempty"`
	InternalHTTPCode int             `json:"internal_http_code,omitempty"`
	ErrorCode        int             `json:"error_code,omitempty"`
}

func (j *JSend) write(w http.ResponseWriter) {
	if j.Data == nil && j.Message != "" {
		j.Status = JSendError
	}

	body, err := json.Marshal(j)
	if err != nil {
		body = []byte(`{"status":"` + JSendError + `","message":"unable to correctly parse response on server"}`)
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}
