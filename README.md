# eugen
The https Server with Security, Caching and automatic Certificate renewal


### What it does

1. provide a recommended TLS Configuration (Ciphers, Curves, TLS Version)
2. provide inmemory caching for webapp files, ignore certain image types which are already compressed
2.1 provide hot cache update for new files or changed files
3. provide automatic certificate renewal
4. provide access to the cache to serve e.g. index.html

### How to

1. Create static folder next to your binary
2. Put all your webapp files in the static folder
3. Config eugen
4. Start Server


```go
package main
import (	
  "github.com/Bhei/eugen/eugen"	
  "net/http"
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
    
    
    //create the Server	
    srv := eugen.New(config)	
    
    //run it!	
    srv.Start()
}
```
