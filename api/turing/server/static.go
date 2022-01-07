package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	logger "github.com/gojek/turing/api/turing/log"
	"github.com/gorilla/mux"
)

// SPAHandler implements the http.Handler interface.
// The path to the static directory and path to the
// index file within that static directory are used to
// serve the SPA in the given static directory.
type SPAHandler struct {
	ServingPath string
	StaticPath  string
	IndexPath   string
}

func NewSinglePageApplicationHandler(servingPath, appDir string) SPAHandler {
	return SPAHandler{
		ServingPath: servingPath,
		StaticPath:  appDir,
		IndexPath:   "index.html",
	}
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served
func (h SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	path = strings.TrimPrefix(path, h.ServingPath)

	// prepend the path with the path to the static directory
	path = filepath.Join(h.StaticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		indexPath := filepath.Join(h.StaticPath, h.IndexPath)

		// check if index files exists
		_, err = os.Stat(indexPath)
		if os.IsNotExist(err) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Ref: https://create-react-app.dev/docs/production-build/#static-file-caching
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Expires", "0") // For proxies

		http.ServeFile(w, r, indexPath)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.StripPrefix(
		h.ServingPath,
		http.FileServer(http.Dir(h.StaticPath)),
	).ServeHTTP(w, r)
}

func ServeSinglePageApplication(r *mux.Router, path string, appDir string) {
	r.PathPrefix(path).Handler(NewSinglePageApplicationHandler(path, appDir))
}

func serveByteArray(r *mux.Router, path string, bytes []bytes, contentType string) {
	r.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
        w.Header().Set("Content-type", contentType)
		_, err := w.Write(bytes)
		if err != nil {
			logger.Errorf("error writing openapi yaml: %s", err)
		}
	})
}

func ServeYAML(r *mux.Router, path string, bytes []byte) {
    serveByteArray(r *mux.Router, path string, bytes []byte(), "application/x-yaml")
}
