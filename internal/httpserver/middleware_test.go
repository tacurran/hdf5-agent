package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRequestIDLength(t *testing.T) {
	id := newRequestID()
	if len(id) != 32 {
		t.Fatalf("len = %d id=%s", len(id), id)
	}
}

func TestRequestIDEmptyWithoutMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if RequestID(req.Context()) != "" {
		t.Fatal("expected empty")
	}
}
