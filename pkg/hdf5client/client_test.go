package hdf5client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientListAndHealth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": "hdf5-agent", "version": "1.1.0", "api": "v1"})
	})
	mux.HandleFunc("GET /api/v1/files", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"files": []map[string]any{{"name": "a.h5", "size_bytes": 3}}})
	})
	mux.HandleFunc("GET /api/v1/files/{name}/datasets", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("path") != "/g/ds" {
			w.WriteHeader(400)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"name": "ds", "path": "/g/ds", "shape": []int{2}, "dtype": "float64", "npoints": 2, "truncated": false, "data": []float64{1, 2}})
	})
	mux.HandleFunc("PUT /api/v1/files/{name}/datasets", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"status":"updated"}`))
	})
	mux.HandleFunc("POST /api/v1/files", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status":"created"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := New(srv.URL)
	ctx := context.Background()
	h, err := c.Health(ctx)
	if err != nil || h.Status != "ok" {
		t.Fatalf("health %v %v", h, err)
	}
	files, err := c.ListFiles(ctx)
	if err != nil || len(files) != 1 || files[0].Name != "a.h5" {
		t.Fatalf("files %#v %v", files, err)
	}
	ds, err := c.ReadDataset(ctx, "a.h5", "/g/ds")
	if err != nil || ds.NPoints != 2 {
		t.Fatalf("dataset %#v %v", ds, err)
	}
	if err := c.CreateFile(ctx, "b.h5"); err != nil {
		t.Fatal(err)
	}
	if err := c.UpdateDataset(ctx, "a.h5", "/g/ds", []int{0}, []any{9}); err != nil {
		t.Fatal(err)
	}
}

func TestClientAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error":{"code":"not_found","message":"missing"}}`))
	}))
	t.Cleanup(srv.Close)
	c := New(srv.URL)
	_, err := c.Structure(context.Background(), "nope.h5")
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.StatusCode != 404 {
		t.Fatalf("err = %#v", err)
	}
	if apiErr.Body.Error.Code != "not_found" {
		t.Fatalf("code %s", apiErr.Body.Error.Code)
	}
}
