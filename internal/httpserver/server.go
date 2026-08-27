package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tacurran/hdf5-agent/api"
	"github.com/tacurran/hdf5-agent/internal/apierror"
	"github.com/tacurran/hdf5-agent/internal/config"
	"github.com/tacurran/hdf5-agent/internal/hdf5store"
	"github.com/tacurran/hdf5-agent/internal/version"
)

// Store is the HDF5 persistence surface used by HTTP handlers.
type Store interface {
	Dir() string
	Ready(ctx context.Context) error
	ListFiles(ctx context.Context) ([]hdf5store.FileInfo, error)
	Structure(ctx context.Context, name string) (*hdf5store.Node, error)
	ReadDataset(ctx context.Context, name, datasetPath string) (*hdf5store.Dataset, error)
	CreateFile(ctx context.Context, name string) error
	DeleteFile(ctx context.Context, name string) error
	UpdateDataset(ctx context.Context, name, datasetPath string, indices []int, values []any) error
}

// Server is the HDF5 Agent HTTP service.
type Server struct {
	cfg   config.Config
	store Store
	log   *slog.Logger
	http  *http.Server
}

// New constructs a Server with production timeouts and middleware.
func New(cfg config.Config, store Store, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	s := &Server{cfg: cfg, store: store, log: log}
	s.http = &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           s.routes(),
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    1 << 20,
	}
	return s
}

// Handler exposes the HTTP handler for tests.
func (s *Server) Handler() http.Handler {
	return s.http.Handler
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.HandleFunc("GET /readyz", s.handleReady)
	mux.HandleFunc("GET /api/v1/health", s.handleHealth)
	mux.HandleFunc("GET /api/v1/ready", s.handleReady)
	mux.HandleFunc("GET /api/v1/openapi.yaml", s.handleOpenAPI)
	mux.HandleFunc("GET /api/v1/files", s.handleListFiles)
	mux.HandleFunc("POST /api/v1/files", s.handleCreateFile)
	mux.HandleFunc("GET /api/v1/files/{name}", s.handleFileStructure)
	mux.HandleFunc("DELETE /api/v1/files/{name}", s.handleDeleteFile)
	mux.HandleFunc("GET /api/v1/files/{name}/datasets", s.handleReadDataset)
	mux.HandleFunc("PUT /api/v1/files/{name}/datasets", s.handleUpdateDataset)

	var handler http.Handler = mux
	if s.cfg.StaticDir != "" {
		handler = s.withStatic(mux)
	}
	handler = http.MaxBytesHandler(handler, s.cfg.MaxRequestBytes)
	handler = withSecurityHeaders(handler)
	handler = withCORS(s.cfg.CORSOrigins, handler)
	handler = withLogging(s.log, handler)
	handler = withRecover(s.log, handler)
	handler = withRequestID(handler)
	return handler
}

func (s *Server) withStatic(api http.Handler) http.Handler {
	dir := http.Dir(s.cfg.StaticDir)
	files := http.FileServer(dir)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			api.ServeHTTP(w, r)
			return
		}
		path := r.URL.Path
		if path == "/" {
			files.ServeHTTP(w, r)
			return
		}
		f, err := dir.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			if r.Method == http.MethodGet {
				r = r.Clone(r.Context())
				r.URL.Path = "/"
				files.ServeHTTP(w, r)
				return
			}
			http.NotFound(w, r)
			return
		}
		_ = f.Close()
		files.ServeHTTP(w, r)
	})
}

// Run listens until ctx is cancelled, then shuts down gracefully.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.log.Info("listen", "addr", s.cfg.HTTPAddr, "data_dir", s.cfg.DataDir, "version", version.Version)
		err := s.http.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()
		s.log.Info("shutdown", "timeout", s.cfg.ShutdownTimeout.String())
		if err := s.http.Shutdown(shutCtx); err != nil {
			return err
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}

// Serve binds to ln. Used by tests that need a real listener.
func (s *Server) Serve(ln net.Listener) error {
	s.http.Addr = ln.Addr().String()
	return s.http.Serve(ln)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, r, http.StatusOK, hdf5store.HealthResponse{
		Status:  "ok",
		Service: version.Name,
		Version: version.Version,
		API:     version.API,
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ready(r.Context()); err != nil {
		s.writeError(w, r, apierror.CodeNotReady, "data directory is not ready: "+err.Error())
		return
	}
	s.writeJSON(w, r, http.StatusOK, hdf5store.ReadyResponse{
		Status:  "ready",
		DataDir: s.store.Dir(),
	})
}

func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(api.OpenAPIYAML)
}

func (s *Server) handleListFiles(w http.ResponseWriter, r *http.Request) {
	files, err := s.store.ListFiles(r.Context())
	if err != nil {
		s.writeStoreError(w, r, err)
		return
	}
	s.writeJSON(w, r, http.StatusOK, hdf5store.FilesResponse{Files: files})
}

func (s *Server) handleCreateFile(w http.ResponseWriter, r *http.Request) {
	var req hdf5store.CreateFileRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, r, apierror.CodeInvalidRequest, err.Error())
		return
	}
	if req.Name == "" {
		s.writeError(w, r, apierror.CodeInvalidRequest, "name is required")
		return
	}
	if err := s.store.CreateFile(r.Context(), req.Name); err != nil {
		s.writeStoreError(w, r, err)
		return
	}
	s.writeJSON(w, r, http.StatusCreated, map[string]string{
		"status": "created",
		"name":   req.Name,
	})
}

func (s *Server) handleFileStructure(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	node, err := s.store.Structure(r.Context(), name)
	if err != nil {
		s.writeStoreError(w, r, err)
		return
	}
	s.writeJSON(w, r, http.StatusOK, node)
}

func (s *Server) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := s.store.DeleteFile(r.Context(), name); err != nil {
		s.writeStoreError(w, r, err)
		return
	}
	s.writeJSON(w, r, http.StatusOK, map[string]string{"status": "deleted", "name": name})
}

func (s *Server) handleReadDataset(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	path := r.URL.Query().Get("path")
	if path == "" {
		s.writeError(w, r, apierror.CodeInvalidRequest, "query parameter path is required")
		return
	}
	ds, err := s.store.ReadDataset(r.Context(), name, path)
	if err != nil {
		s.writeStoreError(w, r, err)
		return
	}
	s.writeJSON(w, r, http.StatusOK, ds)
}

func (s *Server) handleUpdateDataset(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var req hdf5store.UpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		s.writeError(w, r, apierror.CodeInvalidRequest, err.Error())
		return
	}
	if req.Path == "" {
		s.writeError(w, r, apierror.CodeInvalidRequest, "path is required")
		return
	}
	values := req.ValuesOrData()
	if err := s.store.UpdateDataset(r.Context(), name, req.Path, req.Indices, values); err != nil {
		s.writeStoreError(w, r, err)
		return
	}
	s.writeJSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}

func decodeJSON(r *http.Request, dest any) error {
	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	if err := dec.Decode(dest); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("request body is required")
		}
		return err
	}
	return nil
}

func (s *Server) writeJSON(w http.ResponseWriter, r *http.Request, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		s.log.Error("encode json", "err", err, "request_id", RequestID(r.Context()))
	}
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, code apierror.Code, message string) {
	s.writeJSON(w, r, apierror.HTTPStatus(code), apierror.New(code, message, RequestID(r.Context())))
}

func (s *Server) writeStoreError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, hdf5store.ErrInvalidName):
		s.writeError(w, r, apierror.CodeInvalidName, err.Error())
	case errors.Is(err, hdf5store.ErrInvalidPath):
		s.writeError(w, r, apierror.CodeInvalidPath, err.Error())
	case errors.Is(err, hdf5store.ErrNotFound):
		s.writeError(w, r, apierror.CodeNotFound, err.Error())
	case errors.Is(err, hdf5store.ErrConflict):
		s.writeError(w, r, apierror.CodeConflict, err.Error())
	case errors.Is(err, hdf5store.ErrUnsupportedType):
		s.writeError(w, r, apierror.CodeUnsupportedType, err.Error())
	case errors.Is(err, hdf5store.ErrTooLarge):
		s.writeError(w, r, apierror.CodeTooLarge, err.Error())
	case errors.Is(err, os.ErrNotExist), errors.Is(err, fs.ErrNotExist):
		s.writeError(w, r, apierror.CodeNotFound, err.Error())
	default:
		s.log.Error("store error", "err", err, "request_id", RequestID(r.Context()))
		s.writeError(w, r, apierror.CodeInternal, "internal error")
	}
}
