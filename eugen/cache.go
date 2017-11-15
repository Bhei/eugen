package eugen

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type cachedFile struct {
	filename    string
	path        string
	mtime       time.Time
	contentGzip *bytes.Buffer
	contentBr   *bytes.Buffer
	contentRaw  *bytes.Buffer
	etag        string
}

func (c *cachedFile) createEtag() {
	t := c.mtime
	c.etag = strings.Join([]string{
		fmt.Sprintf("%d%d%d%d%d", t.Year(),
			t.Month(), t.Day(), t.Hour(), t.Minute()),
		"-",
		strconv.Itoa(c.contentRaw.Len()),
		"-",
		fmt.Sprintf("%d", toUTF8DigitRepresentation(c.filename))}, "")
}

func createCache(staticpath string, cacheFileExtensions []string) *Cache {
	fileArray, cacheErr := getAllFiles(staticpath, cacheFileExtensions)

	if cacheErr != nil {
		log.Fatal("Cache error: ", cacheErr)
	}

	return makeCache(fileArray)
}

type Cache struct {
	staticFilesCache map[string]*cachedFile
}

func (c *Cache) addCachedFile(filename string) {
	//check if file path is already in the cache
	_, ok := c.staticFilesCache[filename]

	if ok {
		c.updateCachedFile(filename)
	} else {
		cfile, err := handleFile(filename)

		if err != nil {
			log.Println("[Cache] Adding file error: ", filename)
			return
		}
		c.staticFilesCache[filename] = cfile
	}
}

func (c *Cache) updateCachedFile(filename string) {
	cfile, err := handleFile(filename)

	if err != nil {
		log.Println("[Cache] Updating file error: ", filename)
		return
	}
	c.staticFilesCache[filename] = cfile
	log.Println("[Cache] updated file: ", filename)
}
