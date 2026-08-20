// Command catalog is a sibling data-platform service that inventories HDF5 Agent.
//
// It never links libhdf5. All communication is the versioned HTTP API.
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tacurran/hdf5-agent/internal/hdf5store"
	"github.com/tacurran/hdf5-agent/pkg/hdf5client"
)

func main() {
	base := getenv("HDF5_AGENT_URL", "http://127.0.0.1:8080")
	addr := getenv("HTTP_ADDR", ":8090")
	client := hdf5client.New(base)
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"data-catalog"}`))
	})
	mux.HandleFunc("GET /inventory", func(w http.ResponseWriter, r *http.Request) {
		inv, err := buildInventory(r.Context(), client)
		if err != nil {
			log.Error("inventory", "err", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{"code": "upstream", "message": err.Error()},
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(inv)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		log.Info("catalog listen", "addr", addr, "hdf5_agent", base)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen", "err", err)
			os.Exit(1)
		}
	}()
	<-ctx.Done()
	shut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shut)
}

type inventory struct {
	Source      string          `json:"source"`
	GeneratedAt string          `json:"generated_at"`
	Files       []inventoryFile `json:"files"`
}

type inventoryFile struct {
	Name      string             `json:"name"`
	SizeBytes int64              `json:"size_bytes"`
	Datasets  []inventoryDataset `json:"datasets"`
}

type inventoryDataset struct {
	Path    string `json:"path"`
	Dtype   string `json:"dtype"`
	Shape   []uint `json:"shape"`
	NPoints int    `json:"npoints"`
}

func buildInventory(ctx context.Context, client *hdf5client.Client) (*inventory, error) {
	files, err := client.ListFiles(ctx)
	if err != nil {
		return nil, err
	}
	inv := &inventory{
		Source:      "hdf5-agent",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Files:       []inventoryFile{},
	}
	for _, f := range files {
		entry := inventoryFile{Name: f.Name, SizeBytes: f.SizeBytes, Datasets: []inventoryDataset{}}
		tree, err := client.Structure(ctx, f.Name)
		if err != nil {
			return nil, err
		}
		collectDatasets(tree, &entry.Datasets)
		inv.Files = append(inv.Files, entry)
	}
	return inv, nil
}

func collectDatasets(node *hdf5store.Node, out *[]inventoryDataset) {
	if node == nil {
		return
	}
	if node.Type == "dataset" {
		*out = append(*out, inventoryDataset{
			Path:    node.Path,
			Dtype:   node.Dtype,
			Shape:   node.Shape,
			NPoints: node.NPoints,
		})
	}
	for _, child := range node.Children {
		collectDatasets(child, out)
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
