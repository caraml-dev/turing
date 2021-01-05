package web_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	tu "github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/gojek/turing/api/turing/web"
	"github.com/stretchr/testify/require"
)

func startTestHTTPServer(mux *http.ServeMux, address string) *http.Server {
	srv := &http.Server{
		Addr:    address,
		Handler: mux,
	}

	go func() {
		_ = srv.ListenAndServe()
	}()

	return srv
}

func TestFileHandler(t *testing.T) {
	mux := http.NewServeMux()

	filePath := filepath.Join("..", "testdata", "cluster", "servicebuilder", "router_version_basic.json")
	mux.Handle("/path", web.FileHandler(filePath, true))
	mux.Handle("/not-found", web.FileHandler(fmt.Sprintf("%d.file", time.Now().Unix()), false))

	srv := startTestHTTPServer(mux, ":9999")
	defer func() {
		_ = srv.Shutdown(context.Background())
	}()

	resp, httpErr := http.DefaultClient.Get("http://localhost:9999/path")
	require.NoError(t, httpErr)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "no-cache, no-store, must-revalidate", resp.Header.Get("Cache-Control"))
	require.Equal(t, "0", resp.Header.Get("Expires"))

	respBytes, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.NoError(t, err)

	data, _ := tu.ReadFile(filePath)
	require.Equal(t, data, respBytes)

	resp, httpErr = http.DefaultClient.Get("http://localhost:9999/not-found")
	require.NoError(t, httpErr)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	_ = resp.Body.Close()
}
