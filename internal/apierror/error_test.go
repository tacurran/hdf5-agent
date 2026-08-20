package apierror

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHTTPStatus(t *testing.T) {
	cases := map[Code]int{
		CodeInvalidRequest:   http.StatusBadRequest,
		CodeInvalidName:      http.StatusBadRequest,
		CodeNotFound:         http.StatusNotFound,
		CodeConflict:         http.StatusConflict,
		CodeUnsupportedType:  http.StatusUnprocessableEntity,
		CodeTooLarge:         http.StatusRequestEntityTooLarge,
		CodeNotReady:         http.StatusServiceUnavailable,
		CodeInternal:         http.StatusInternalServerError,
		Code("unknown"):      http.StatusInternalServerError,
		CodeMethodNotAllowed: http.StatusMethodNotAllowed,
	}
	for code, want := range cases {
		if got := HTTPStatus(code); got != want {
			t.Errorf("HTTPStatus(%q) = %d, want %d", code, got, want)
		}
	}
}

func TestNewJSONShape(t *testing.T) {
	body := New(CodeNotFound, "file not found: a.h5", "req-1")
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"error":{"code":"not_found","message":"file not found: a.h5","request_id":"req-1"}}`
	if string(raw) != want {
		t.Errorf("json = %s", raw)
	}
}
