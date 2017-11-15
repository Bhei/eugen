package eugen

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
)

type staticFilesHandler struct {
	cache *Cache
}

func (s staticFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.String()[1:]
	fmt.Println("acces to static files at path: ", path)
	cached, ok := s.cache.staticFilesCache[path]

	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Keep-Alive", "timeout=5, max=100")
	w.Header().Set("Cache-Control", "max-age=2592000") // 30 days
	w.Header().Set("Etag", cached.etag)
	AddSecurityHeaders(w)

	encHeader := r.Header.Get("Accept-Encoding")

	if strings.Contains(encHeader, "br") && cached.contentBr != nil {
		serveBr(cached, w, r)
	} else if strings.Contains(encHeader, "gzip") && cached.contentGzip != nil {
		serveGzip(cached, w, r)
	} else {
		serveRaw(cached, w, r)
	}
}

func serveBr(cached *cachedFile, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Encoding", "br")
	reader := bytes.NewReader(cached.contentBr.Bytes())
	http.ServeContent(w, r, cached.filename, cached.mtime, reader)
}

func serveGzip(cached *cachedFile, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Encoding", "gzip")
	reader := bytes.NewReader(cached.contentGzip.Bytes())
	http.ServeContent(w, r, cached.filename, cached.mtime, reader)
}

func serveRaw(cached *cachedFile, w http.ResponseWriter, r *http.Request) {
	reader := bytes.NewReader(cached.contentRaw.Bytes())
	http.ServeContent(w, r, cached.filename, cached.mtime, reader)
}
