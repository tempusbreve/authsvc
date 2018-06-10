package common // import "breve.us/authsvc/common"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

// Middleware describes common middleware
type Middleware interface {
	ServeHTTP(http.ResponseWriter, *http.Request, http.HandlerFunc)
}

// QueryParamKeys gets the keys from a url.Values collection.
func QueryParamKeys(v url.Values) []string {
	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	return keys
}

// Redirect initiates an http redirect response.
func Redirect(w http.ResponseWriter, r *http.Request, targetURL string, additionalParams map[string]string) {
	dest, err := url.Parse(targetURL)
	if err != nil {
		JSONStatusResponse(http.StatusInternalServerError, w, err)
		return
	}
	q := dest.Query()
	for k, v := range additionalParams {
		q.Set(k, v)
	}
	dest.RawQuery = q.Encode()
	http.Redirect(w, r, dest.String(), http.StatusSeeOther)
}

// JSONResponse writes the interface{} o as a JSON object response body.
func JSONResponse(w http.ResponseWriter, o interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(o); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if _, e2 := fmt.Fprintf(w, `{"error": "%v"}`, err); e2 != nil {
			// TODO: ?
		}
	}
}

// JSONStatusResponse writes the interface{} o as a JSON object response
// body with the givent http status code.
func JSONStatusResponse(code int, w http.ResponseWriter, o interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if o != nil {
		if err := json.NewEncoder(w).Encode(o); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			if _, e2 := fmt.Fprintf(w, `{"error": "%v"}`, err); e2 != nil {
				// TODO: ?
			}
		}
	}
}

// NewTemplateHandler loads each filename parameter as a template, and
// serves the first one to match, or returns 404 if one is not found.
func NewTemplateHandler(filenames ...string) http.Handler {
	return &templateHandler{t: template.Must(template.ParseFiles(filenames...))}
}

type templateHandler struct {
	t *template.Template
}

func (h *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, name := path.Split(r.URL.Path)
	if name == "" {
		name = "index.html"
	}
	if tmpl := h.t.Lookup(name); tmpl != nil {
		buf := &bytes.Buffer{}
		if err := tmpl.Execute(buf, nil); err == nil {
			http.ServeContent(w, r, name, time.Now().UTC(), bytes.NewReader(buf.Bytes()))
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	if _, err := fmt.Fprintf(w, "404 :: not found"); err != nil {
		// TODO: ?
	}
}

// NewStaticFileHandler tries each filename parameter and serves the
// first one successfully read, or returns 404 if none are found.
func NewStaticFileHandler(filenames ...string) http.Handler {
	if len(filenames) == 0 {
		filenames = []string{"index.html"}
	}
	return &staticFileHandler{filenames: filenames}
}

type staticFileHandler struct {
	filenames []string
}

func (h *staticFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, filename := range h.filenames {
		if fd, err := os.Open(filename); err == nil {
			defer func() {
				if err := fd.Close(); err != nil {
					// TODO: ?
				}
			}()
			if fi, err := fd.Stat(); err == nil {
				http.ServeContent(w, r, fi.Name(), fi.ModTime(), fd)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
	if _, err := fmt.Fprintf(w, "404 :: not found"); err != nil {
		// TODO: ?
	}
}

// MatchingFiles returns the names of all the files in dir with (case
// insensitive) extension matching one of extensions.
func MatchingFiles(dir string, extensions ...string) []string {
	templates := []string{}
	if files, err := ioutil.ReadDir(dir); err == nil {
		for _, fname := range files {
			lfn := strings.ToLower(fname.Name())
			for _, ext := range extensions {
				if strings.HasSuffix(lfn, strings.ToLower(ext)) {
					templates = append(templates, path.Join(dir, fname.Name()))
				}
			}
		}
	}
	return templates
}
