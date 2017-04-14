package webCache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Mux struct {
	cache    map[string][]byte
	errCache map[string][]byte
	errMatch *regexp.Regexp
}

func NewMux() *Mux {
	mux := &Mux{
		cache:    make(map[string][]byte),
		errCache: make(map[string][]byte),
	}
	r, err := regexp.Compile("/error/[\\d]+$")
	if err != nil {
		log.Fatalf("web.Cache.go >> NewMux() >> regexp.Compile() >> %v\n\n", err)
	}
	mux.errMatch = r
	mux.setCache()

	return mux
}

func (m *Mux) setCache() {
	if err := filepath.Walk("serve", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		path = strings.Replace(path, "serve/", "", 1)

		if info.Name() == "index.html" {
			shortPath := strings.Replace(path, "index.html", "", -1)
			if shortPath != "/" && strings.HasSuffix(shortPath, "/") {
				shortPath = shortPath[:len(shortPath)-1]
			}

			m.cache["/"+shortPath] = b
		}

		m.cache["/"+path] = b
		return nil
	}); err != nil {
		panic(err)
	}
}

var ERROR = `<html><head><title>Error %d</title></head><body style="text-align:center;"><h1>Error %d</h1><p>%s</o></body></html>`

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.errMatch.MatchString(r.URL.Path) {
		if err, ok := m.errCache[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "text/html; utf-8")
			fmt.Fprint(w, err)
			return
		}
		codestr := strings.Replace(r.URL.Path, "/error/", "", -1)
		code, err := strconv.Atoi(codestr)
		if err != nil {
			code = 500
		}
		w.Header().Set("Content-Type", "text/html; utf-8")
		fmt.Fprintf(w, ERROR, code, code, http.StatusText(code))
		return
	}
	if file, ok := m.cache[r.URL.Path]; ok {
		w.Write(file)
		return
	}
	http.Redirect(w, r, "/error/404", 303)
	return
}
