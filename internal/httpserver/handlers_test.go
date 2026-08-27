package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tacurran/hdf5-agent/internal/config"
	"github.com/tacurran/hdf5-agent/internal/hdf5store"
)

type fakeStore struct {
	dir     string
	ready   error
	files   []hdf5store.FileInfo
	tree    *hdf5store.Node
	dataset *hdf5store.Dataset
	err     error
	created string
	deleted string
	updated bool
}

func (f *fakeStore) Dir() string { return f.dir }
func (f *fakeStore) Ready(ctx context.Context) error {
	return f.ready
}
func (f *fakeStore) ListFiles(ctx context.Context) ([]hdf5store.FileInfo, error) {
	return f.files, f.err
}
func (f *fakeStore) Structure(ctx context.Context, name string) (*hdf5store.Node, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.tree, nil
}
func (f *fakeStore) ReadDataset(ctx context.Context, name, datasetPath string) (*hdf5store.Dataset, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.dataset, nil
}
func (f *fakeStore) CreateFile(ctx context.Context, name string) error {
	f.created = name
	return f.err
}
func (f *fakeStore) DeleteFile(ctx context.Context, name string) error {
	f.deleted = name
	return f.err
}
func (f *fakeStore) UpdateDataset(ctx context.Context, name, datasetPath string, indices []int, values []any) error {
	f.updated = true
	return f.err
}

func testServer(store Store) *Server {
	cfg := config.Config{
		HTTPAddr:         ":0",
		DataDir:          "/data",
		CORSOrigins:      []string{"*"},
		ReadTimeout:      time.Second,
		WriteTimeout:     time.Second,
		IdleTimeout:      time.Second,
		ShutdownTimeout:  time.Second,
		MaxDatasetPoints: 100,
		MaxRequestBytes:  1 << 20,
	}
	return New(cfg, store, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestHealthAndReady(t *testing.T) {
	s := testServer(&fakeStore{dir: "/data"})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("health status %d body %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request id")
	}
	var health hdf5store.HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatal(err)
	}
	if health.Status != "ok" || health.API != "v1" {
		t.Fatalf("health = %#v", health)
	}

	req = httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("ready status %d", rec.Code)
	}
}

func TestReadyFails(t *testing.T) {
	s := testServer(&fakeStore{ready: context.DeadlineExceeded})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestListAndStructure(t *testing.T) {
	s := testServer(&fakeStore{
		files: []hdf5store.FileInfo{{Name: "a.h5", SizeBytes: 12}},
		tree:  &hdf5store.Node{Name: "a.h5", Type: "file"},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"a.h5"`) {
		t.Fatalf("body %s", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/files/a.h5", nil)
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
}

func TestReadDatasetRequiresPath(t *testing.T) {
	s := testServer(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/a.h5/datasets", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestReadDataset(t *testing.T) {
	s := testServer(&fakeStore{dataset: &hdf5store.Dataset{Name: "waveform", Path: "/measurements/waveform", Data: []float64{1, 2}}})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/a.h5/datasets?path=/measurements/waveform", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
}

func TestCreateDeleteUpdate(t *testing.T) {
	fs := &fakeStore{}
	s := testServer(fs)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/files", strings.NewReader(`{"name":"new.h5"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	if fs.created != "new.h5" {
		t.Fatalf("created %q", fs.created)
	}

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/files/new.h5", nil)
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 || fs.deleted != "new.h5" {
		t.Fatalf("delete %d %q", rec.Code, fs.deleted)
	}

	body := `{"path":"/measurements/waveform","indices":[0],"values":[1.25]}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/files/a.h5/datasets", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 || !fs.updated {
		t.Fatalf("update %d %s", rec.Code, rec.Body.String())
	}
}

func TestNotFoundMapping(t *testing.T) {
	s := testServer(&fakeStore{err: hdf5store.ErrNotFound})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/files/missing.h5", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"not_found"`) {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestOpenAPI(t *testing.T) {
	s := testServer(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "openapi:") {
		t.Fatalf("not openapi: %s", rec.Body.String()[:80])
	}
}

func TestCORSPreflight(t *testing.T) {
	s := testServer(&fakeStore{})
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/files", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("cors %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestIncomingRequestIDPreserved(t *testing.T) {
	s := testServer(&fakeStore{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", "client-id-1")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Header().Get("X-Request-ID") != "client-id-1" {
		t.Fatalf("id %q", rec.Header().Get("X-Request-ID"))
	}
}
