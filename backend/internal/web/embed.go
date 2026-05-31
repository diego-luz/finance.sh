// Package web embeds the built frontend SPA bundle into the API binary so a
// single container serves both the static app and the JSON API on one port.
// Self-hosters point their own reverse proxy (Traefik, Caddy, Nginx Proxy
// Manager, ...) at this one HTTP port and let it terminate TLS.
//
// The real bundle is baked in at build time by the Docker build, which replaces
// the placeholder dist/ below with the output of `npm run build`. A minimal
// placeholder index.html is committed so `go build` (and CI) succeed without a
// frontend build present.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// Handler returns an http.Handler that serves the embedded SPA. It is meant to
// be mounted as the root catch-all ("/*") AFTER the API/docs/health routes, so
// those keep priority. Real files (hashed assets, the PWA service worker,
// manifest, icons) are served as-is; any other path falls back to index.html so
// the client-side router can resolve deep links (/transactions, /settings, ...).
func Handler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		// Compile-time embed guarantees dist/ exists; a failure here is a
		// programming error, not a runtime condition.
		panic("web: invalid embedded dist FS: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqPath := strings.TrimPrefix(r.URL.Path, "/")
		if reqPath == "" {
			reqPath = "index.html"
		}

		// Probe the embedded FS: if the path is not a real file, it's a
		// client-side route — serve the app shell so the SPA router takes over.
		if info, statErr := fs.Stat(sub, reqPath); statErr != nil || info.IsDir() {
			setCacheHeaders(w, "index.html")
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}

		setCacheHeaders(w, reqPath)
		fileServer.ServeHTTP(w, r)
	})
}

// setCacheHeaders sets the caching policy: content-hashed build assets are
// immutable and cached aggressively, while the app shell, service worker and
// manifest must always be revalidated so a deploy is picked up immediately.
func setCacheHeaders(w http.ResponseWriter, p string) {
	switch {
	case p == "index.html",
		p == "sw.js",
		p == "workbox-sw.js",
		p == "registerSW.js",
		p == "manifest.webmanifest":
		w.Header().Set("Cache-Control", "no-cache")
	case strings.HasPrefix(p, "assets/"):
		// Vite emits content-hashed filenames under assets/ — safe to cache for
		// a year and mark immutable.
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}
}
