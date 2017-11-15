package eugen

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/crypto/acme/autocert"
)

type Server struct {
	config Config
	cache  *Cache
}

type Config struct {
	WebResourcePath     string
	CacheFileExtensions []string
	Domains             []string
	Mux                 *http.ServeMux
}

func New(config Config) *Server {
	server := Server{config: config}
	return &server
}

func (s *Server) ServeCachedFile(name string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-XSS-Prodection", "1; mode=block")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

	encHeader := r.Header.Get("Accept-Encoding")

	cached, ok := s.cache.staticFilesCache[name]
	if !ok {
		http.NotFound(w, r)
		return
	}

	if strings.Contains(encHeader, "br") && cached.contentBr != nil {
		serveBr(cached, w, r)
	} else if strings.Contains(encHeader, "gzip") && cached.contentGzip != nil {
		serveGzip(cached, w, r)
	} else {
		serveRaw(cached, w, r)
	}
}

func (s *Server) Start() {
	s.cache = createCache(s.config.WebResourcePath,
		s.config.CacheFileExtensions)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("fsnotify error: ", err)
	}

	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					s.cache.updateCachedFile(event.Name)
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					s.cache.addCachedFile(event.Name)
				}

			case err := <-watcher.Errors:
				log.Println("watcher error: ", err)
			}
		}
	}()

	err = watcher.Add(s.config.WebResourcePath)

	if err != nil {
		log.Fatal("watcher path static/ error: ", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	if err != nil {
		log.Fatal("setup exchange server error: ", err)
	}

	s.config.Mux.Handle("/static/", staticFilesHandler{cache: s.cache})

	srv := &http.Server{
		Addr:              ":80",
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,

		//http redirect to https handler
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Connection", "close")
			url := "https://" + req.Host + req.URL.String()
			http.Redirect(w, req, url, http.StatusMovedPermanently)
		}),
	}

	//http server
	go func() { log.Fatal(srv.ListenAndServe()) }()

	certManager := autocert.Manager{
		Cache:      autocert.DirCache(".secret"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.config.Domains...),
	}

	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
			tls.CurveP521},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		},
		GetCertificate: certManager.GetCertificate,
	}

	//https server
	secureServer := &http.Server{
		Handler:           s.config.Mux,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		TLSConfig:         tlsConfig,
	}
	go func() {
		log.Fatal(secureServer.ListenAndServeTLS("", ""))
	}()

	<-c
	fmt.Println("Shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	secureServer.Shutdown(ctx)
	srv.Shutdown(ctx)

	log.Println("Server gracefully stopped")
}
