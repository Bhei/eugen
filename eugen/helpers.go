package eugen

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func toUTF8DigitRepresentation(text string) int {
	var value int
	for _, trune := range text {
		value += int(trune - '0')
	}
	return value
}

func isCompressable(filename string) bool {
	isImage := true

	for _, ext := range noCompressionExtensions {
		if strings.Contains(filename, ext) {
			isImage = false
			break
		}
	}

	log.Println("File: ", filename, " compressable: ", isImage)
	return isImage
}

func isWantedFile(file string, wantedFileExtensions []string) bool {
	wanted := false

	for _, ext := range wantedFileExtensions {
		if strings.Contains(file, ext) {
			wanted = true
			break
		}
	}

	return wanted
}

func handleCompression(filename string, data []byte) compressed {
	var c compressed
	if !isCompressable(filename) {
		c.bufRaw = bytes.NewBuffer(data)
		c.bufGzip = nil
		c.bufBr = nil

	} else {
		c.bufGzip = compressWithGzip(data)
		c.bufBr = compressWithBrotli(data)
		c.bufRaw = bytes.NewBuffer(data)
	}

	return c
}

func handleFile(filename string) (*cachedFile, error) {
	contents, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	comp := handleCompression(filename, contents)

	fstat, err := os.Stat(filename)

	if err != nil {
		return nil, err
	}

	cfile := &cachedFile{
		filename:    fstat.Name(),
		path:        filename,
		mtime:       fstat.ModTime(),
		contentGzip: comp.bufGzip,
		contentBr:   comp.bufBr,
		contentRaw:  comp.bufRaw}
	cfile.createEtag()
	return cfile, nil
}

func makeCache(files []cachedFile) *Cache {
	var cache Cache
	cache.staticFilesCache = make(map[string]*cachedFile)

	for _, file := range files {
		cfile, err := handleFile(file.path)

		if err != nil {
			log.Fatal("[Cache] Make Cache file error: ", file.path)

		}
		cache.staticFilesCache[file.path] = cfile
		log.Println("[Cache] added file ", cfile.filename)
	}

	log.Println("[Cache] created")
	return &cache
}

func walkDir(path string,
	isWantedFile func(string, []string) bool,
	wantedExtensions []string) ([]cachedFile, error) {

	var files []cachedFile
	finfos, err := ioutil.ReadDir(path)

	if err != nil {
		return files, err
	}

	for _, file := range finfos {

		fpath := path + "/"
		fpath += file.Name()

		if file.IsDir() {
			wfiles, err := walkDir(fpath, isWantedFile, wantedExtensions)
			if err != nil {
				return wfiles, err
			}

			files = append(files, wfiles...)
		} else {
			if isWantedFile(file.Name(), wantedExtensions) {
				cfile := cachedFile{filename: file.Name(), path: fpath,
					mtime: file.ModTime()}
				files = append(files, cfile)
			}
		}
	}

	return files, nil
}

func getAllFiles(path string, wantedExtensions []string) ([]cachedFile, error) {
	res, err := walkDir(path, isWantedFile, wantedExtensions)
	return res, err
}
