package web

import (
	"net/http"
	"path"
)

func FileHandler(path string, disableCaching bool) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if disableCaching {
			// Ref: https://create-react-app.dev/docs/production-build/#static-file-caching
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Expires", "0") // For proxies
		}
		http.ServeFile(w, r, path)
	}

	return http.HandlerFunc(fn)
}

func ServeReactApp(
	mux *http.ServeMux,
	homepage string,
	appDir string,
) {
	appDirFs := http.FileServer(http.Dir(appDir))
	reactEntryHandler := FileHandler(path.Join(appDir, "index.html"), true)

	mux.Handle(
		homepage+"/",
		http.StripPrefix(homepage, fallbackIfNotFoundHandler(appDirFs, reactEntryHandler)))

	mux.Handle("/", reactEntryHandler)
}

// Wrapper ResponseWriter to capture response Status Code
type notFoundResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *notFoundResponseWriter) WriteHeader(status int) {
	w.status = status
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *notFoundResponseWriter) Write(p []byte) (int, error) {
	if w.status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	}
	return len(p), nil
}

func fallbackIfNotFoundHandler(h http.Handler, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nfrw := &notFoundResponseWriter{ResponseWriter: w}
		h.ServeHTTP(nfrw, r)
		if nfrw.status == http.StatusNotFound {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fallback.ServeHTTP(w, r)
		}
	}
}
