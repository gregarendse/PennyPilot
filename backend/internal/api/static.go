package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func spaHandler(staticPath string) http.Handler {
	fileServer := http.FileServer(http.Dir(staticPath))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isBackendPath(r.URL.Path) {
			http.NotFound(w, r)
			return
		}

		requestedPath := filepath.Join(staticPath, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(requestedPath); err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, filepath.Join(staticPath, "index.html"))
	})
}

func isBackendPath(path string) bool {
	return path == "/healthz" ||
		path == "/api" || strings.HasPrefix(path, "/api/") ||
		path == "/auth" || strings.HasPrefix(path, "/auth/")
}
