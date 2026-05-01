package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gonum.org/v1/hdf5"
)

type Server struct {
	router *mux.Router
	port   string
}

func NewServer(port string) *Server {
	return &Server{
		router: mux.NewRouter(),
		port:   port,
	}
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/api/health", s.health).Methods("GET")
	s.router.HandleFunc("/api/files", s.listFiles).Methods("GET")
	s.router.HandleFunc("/api/file/structure", s.getFileStructure).Methods("GET")
	s.router.HandleFunc("/api/dataset/read", s.readDataset).Methods("GET")
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("public")))
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"backend": "Go with gonum/hdf5",
	})
}

func (s *Server) listFiles(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("/app/data")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var hdf5Files []map[string]interface{}
	for _, f := range files {
		if !f.IsDir() && (strings.HasSuffix(f.Name(), ".h5") || strings.HasSuffix(f.Name(), ".hdf5")) {
			info, _ := f.Info()
			hdf5Files = append(hdf5Files, map[string]interface{}{
				"name": f.Name(),
				"size": info.Size(),
				"path": "/app/data/" + f.Name(),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hdf5Files)
}

func (s *Server) getFileStructure(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		http.Error(w, "Missing path parameter", http.StatusBadRequest)
		return
	}

	// If path doesn't start with /, assume it's relative
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/app/data/" + filePath
	}

	file, err := hdf5.OpenFile(filePath, hdf5.F_ACC_RDONLY)
	if err != nil {
		http.Error(w, "Failed to open file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	children, _ := s.listFileObjects(file)

	result := map[string]interface{}{
		"name":     filepath.Base(filePath),
		"path":     filePath,
		"type":     "file",
		"children": children,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) listFileObjects(file *hdf5.File) ([]map[string]interface{}, error) {
	var children []map[string]interface{}

	// Get number of objects in file
	n, err := file.NumObjects()
	if err != nil {
		return children, err
	}

	numObjects := int(n)
	for i := 0; i < numObjects; i++ {
		name, err := file.ObjectNameByIndex(uint(i))
		if err != nil {
			continue
		}

		// Try to open as group
		g, groupErr := file.OpenGroup(name)
		if groupErr == nil {
			defer g.Close()
			subChildren, _ := s.listGroupObjects(g)
			child := map[string]interface{}{
				"name":     name,
				"path":     g.Name(),
				"type":     "group",
				"children": subChildren,
			}
			children = append(children, child)
			continue
		}

		// Try to open as dataset
		dset, datasetErr := file.OpenDataset(name)
		if datasetErr == nil {
			defer dset.Close()

			space := dset.Space()
			defer space.Close()

			dims, _, err := space.SimpleExtentDims()
			if err != nil {
				continue
			}

			dtype, err := dset.Datatype()
			if err != nil {
				continue
			}
			defer dtype.Close()

			child := map[string]interface{}{
				"name":  name,
				"path":  dset.Name(),
				"type":  "dataset",
				"shape": dims,
				"dtype": s.getDatatypeName(dtype),
			}
			children = append(children, child)
		}
	}

	return children, nil
}

func (s *Server) listGroupObjects(group *hdf5.Group) ([]map[string]interface{}, error) {
	var children []map[string]interface{}

	// Recover from panics in HDF5
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic in listGroupObjects: %v", r)
		}
	}()

	n, err := group.NumObjects()
	if err != nil {
		return children, err
	}

	numObjects := int(n)
	for i := 0; i < numObjects; i++ {
		name, err := group.ObjectNameByIndex(uint(i))
		if err != nil {
			continue
		}

		// Try to open as group
		g, groupErr := group.OpenGroup(name)
		if groupErr == nil {
			defer g.Close()
			subChildren, _ := s.listGroupObjects(g)
			child := map[string]interface{}{
				"name":     name,
				"path":     g.Name(),
				"type":     "group",
				"children": subChildren,
			}
			children = append(children, child)
			continue
		}

		// Try to open as dataset with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Skipping dataset %s: %v", name, r)
				}
			}()

			dset, datasetErr := group.OpenDataset(name)
			if datasetErr == nil {
				defer dset.Close()

				space := dset.Space()
				defer space.Close()

				dims, _, err := space.SimpleExtentDims()
				if err != nil {
					log.Printf("Failed to get dims for %s: %v", name, err)
					return
				}

				dtype, err := dset.Datatype()
				if err != nil {
					log.Printf("Failed to get dtype for %s: %v", name, err)
					return
				}
				defer dtype.Close()

				child := map[string]interface{}{
					"name":  name,
					"path":  dset.Name(),
					"type":  "dataset",
					"shape": dims,
					"dtype": s.getDatatypeName(dtype),
				}
				children = append(children, child)
			}
		}()
	}

	return children, nil
}

func (s *Server) readDataset(w http.ResponseWriter, r *http.Request) {
	filePath := r.URL.Query().Get("file")
	datasetPath := r.URL.Query().Get("path")

	if filePath == "" || datasetPath == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in readDataset: %v", r)
			http.Error(w, "Failed to read dataset: "+fmt.Sprintf("%v", r), http.StatusInternalServerError)
		}
	}()

	file, err := hdf5.OpenFile(filePath, hdf5.F_ACC_RDONLY)
	if err != nil {
		http.Error(w, "Failed to open file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	dset, err := file.OpenDataset(datasetPath)
	if err != nil {
		http.Error(w, "Failed to open dataset: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer dset.Close()

	space := dset.Space()
	defer space.Close()

	dims, _, err := space.SimpleExtentDims()
	if err != nil {
		http.Error(w, "Failed to get dimensions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	dtype, err := dset.Datatype()
	if err != nil {
		http.Error(w, "Failed to get datatype: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dtype.Close()

	npoints := 1
	for _, d := range dims {
		npoints *= int(d)
	}

	var data interface{}
	if npoints > 100000 {
		data = map[string]interface{}{
			"truncated": true,
			"message":   fmt.Sprintf("Dataset too large (%d points). Use preview mode.", npoints),
		}
	} else {
		data = s.readDatasetValues(dset, dtype, npoints)
	}

	result := map[string]interface{}{
		"name":  filepath.Base(datasetPath),
		"path":  datasetPath,
		"data":  data,
		"shape": dims,
		"dtype": s.getDatatypeName(dtype),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *Server) readDatasetValues(dset *hdf5.Dataset, dtype *hdf5.Datatype, npoints int) interface{} {
	class := dtype.Class()
	size := dtype.Size()

	log.Printf("Reading dataset: class=%d, size=%d, npoints=%d", class, size, npoints)

	switch class {
	case 0: // H5T_INTEGER (class 0 in some versions)
		log.Printf("Reading as int (class 0)")
		if size == 8 {
			data := make([]int64, npoints)
			err := dset.Read(&data)
			if err != nil {
				log.Printf("Int64 read error: %v", err)
				return map[string]string{"error": err.Error()}
			}
			log.Printf("Successfully read %d int64 values", len(data))
			log.Printf("First 5 values: %v", data[:5])
			return data
		} else if size == 4 {
			data := make([]int32, npoints)
			err := dset.Read(&data)
			if err != nil {
				log.Printf("Int32 read error: %v", err)
				return map[string]string{"error": err.Error()}
			}
			log.Printf("Successfully read %d int32 values", len(data))
			return data
		}

	case 1: // H5T_FLOAT
		log.Printf("Reading as float64")
		data := make([]float64, npoints)
		if err := dset.Read(&data); err != nil {
			log.Printf("Float read error: %v", err)
			return map[string]string{"error": err.Error()}
		}
		return data

	case 2: // H5T_INTEGER
		log.Printf("Reading as int64 (class 2)")
		data := make([]int64, npoints)
		if err := dset.Read(&data); err != nil {
			log.Printf("Integer read error: %v", err)
			return map[string]string{"error": err.Error()}
		}
		return data

	case 3: // H5T_STRING
		log.Printf("Reading as string")
		data := make([]string, npoints)
		if err := dset.Read(&data); err != nil {
			log.Printf("String read error: %v", err)
			return map[string]string{"error": err.Error()}
		}
		return data

	default:
		log.Printf("Unsupported class: %d (size=%d)", class, size)
		return map[string]string{"error": fmt.Sprintf("Unsupported data type class: %d", class)}
	}

	return map[string]string{"error": "Failed to read data"}
}

func (s *Server) getDatatypeName(dtype *hdf5.Datatype) string {
	class := dtype.Class()
	size := dtype.Size()

	switch class {
	case 1: // H5T_FLOAT
		return fmt.Sprintf("float%d", size*8)
	case 2: // H5T_INTEGER
		return fmt.Sprintf("int%d", size*8)
	case 3: // H5T_STRING
		return "string"
	case 4: // H5T_BITFIELD
		return fmt.Sprintf("bitfield%d", size*8)
	case 5: // H5T_OPAQUE
		return "opaque"
	case 6: // H5T_COMPOUND
		return "compound"
	case 7: // H5T_REFERENCE
		return "reference"
	case 8: // H5T_ENUM
		return "enum"
	case 9: // H5T_ARRAY
		return "array"
	case 10: // H5T_VLEN
		return "vlen"
	default:
		return fmt.Sprintf("type_%d", class)
	}
}

func (s *Server) Start() error {
	s.setupRoutes()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	handler := c.Handler(s.router)

	log.Printf("HDF5 Agent Server starting on port %s", s.port)
	return http.ListenAndServe(":"+s.port, handler)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := NewServer(port)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
