package main

import (
	"net/http"

	"github.com/Bhei/eugen/eugen"
)

type MyWebServer struct {
	server *eugen.Server
	//db
	//...
}

func (s *MyWebServer) mainPage(w http.ResponseWriter, r *http.Request) {
	//s.server.AddSaneHTMLHeaders(w)
	//s.server.AddSecurityHeaders(w)

	//serve static index.html file
	s.server.ServeCachedFile("index.html", w, r)
}

func main() {
	var config eugen.Config

	//create your routing setup
	myMux := http.NewServeMux()
	myMux.HandleFunc("/", mainPage)
	//...
	config.Mux = myMux

	//set Path to static content
	config.WebResourcePath = "static"

	//cache all files in WebResourcePath with following file extensions
	extensions := []string{".html", ".css", ".js"}
	config.CacheFileExtensions = extensions

	//configure domains
	domains := []string{"example.com", "example.net", "example.org"}
	config.Domains = domains
	//create the Server
	srv := eugen.New(config)

	//run it!
	srv.Start()
}
