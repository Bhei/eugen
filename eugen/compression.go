package eugen

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"gopkg.in/kothar/brotli-go.v0/enc"
)

type compressed struct {
	bufBr   *bytes.Buffer
	bufRaw  *bytes.Buffer
	bufGzip *bytes.Buffer
}

var noCompressionExtensions []string

func init() {
	noCompressionExtensions = []string{".jpeg", ".jpg", ".png", ".bmp", ".tiff", ".pdf"}
}

func withBrotli(w http.ResponseWriter, payload []byte) {
	w.Header().Set("Content-Encoding", "br")
	w.Header().Set("Content-Type", "text/html")
	log.Println("Serving html with Brotli")
	fmt.Fprint(w, compressWithBrotli(payload))
}

func withGzip(w http.ResponseWriter, payload []byte) {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "text/html")
	log.Println("Serving html with gzip")
	fmt.Fprint(w, compressWithGzip(payload))
}

func withRaw(w http.ResponseWriter, payload []byte) {
	log.Println("Serving raw html")
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, payload)
}

func DetectEncoding(r *http.Request) func(w http.ResponseWriter, payload []byte) {
	//Return compression function
	encHeader := r.Header.Get("Accept-Encoding")
	if strings.Contains(encHeader, "br") {
		return withBrotli
	} else if strings.Contains(encHeader, "gzip") {
		return withGzip
	} else {
		return withRaw
	}
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func compressWithGzip(b []byte) *bytes.Buffer {
	buf := &bytes.Buffer{}
	zw := gzip.NewWriter(buf)
	_, err := zw.Write(b)

	if err != nil {
		log.Println("Gzip compression error: ", err)
		return buf
	}

	err = zw.Close()

	if err != nil {
		log.Println("Gzip compression error: ", err)
		return buf
	}

	return buf
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func compressWithBrotli(input []byte) *bytes.Buffer {
	params := enc.NewBrotliParams()
	// brotli supports quality values from 0 to 11 included
	// 0 is the fastest, 11 is the most compressed but slowest
	params.SetQuality(11)
	compressed, _ := enc.CompressBuffer(params, input, make([]byte, 0))
	buf := bytes.NewBuffer(compressed)
	return buf
}
