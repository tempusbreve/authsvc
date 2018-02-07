package authsvc // import "breve.us/authsvc"

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
)

func DebugLogger(w io.Writer, verbose bool) *logger {
	ll := log.New(w, "[http] ", log.LstdFlags|log.LUTC)
	return &logger{verbose: verbose, l: ll, o: w}
}

type logger struct {
	verbose bool
	o       io.Writer
	l       *log.Logger
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	l.requestLogger(r)
	rw, logResponse := l.responseLogger(r, w)
	defer logResponse()

	next.ServeHTTP(rw, r)
}

func (l *logger) requestLogger(r *http.Request) {
	l.l.Printf("(client %s) %s %s %s [%s]", queryIP(r), r.Host, r.Method, r.URL.Path, r.UserAgent())
	if l.verbose {
		b, err := httputil.DumpRequest(r, true)
		if err != nil {
			l.l.Printf("error dumping request: %v", err)
			return
		}
		fmt.Fprint(l.o, "|REQUEST|\n")
		if _, err = l.o.Write(b); err != nil {
			l.l.Printf("error writing dumped request: %v", err)
			return
		}
	}
}

func (l *logger) responseLogger(r *http.Request, w http.ResponseWriter) (http.ResponseWriter, func()) {
	rr := httptest.NewRecorder()
	return rr, func() {
		for k, v := range rr.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(rr.Code)

		out := []io.Writer{w}

		if l.verbose {
			fmt.Fprint(l.o, "|RESPONSE|\n")
			if err := rr.HeaderMap.Write(l.o); err != nil {
				l.l.Printf("error dumping response headers: %v", err)
			}
			switch ct := rr.HeaderMap.Get("Content-Type"); {
			case ct == "":
				// ignore no content type
			case strings.Contains(ct, "text/"):
				out = append(out, l.o)
			case strings.Contains(ct, "application/json"):
				out = append(out, l.o)
			default:
				l.l.Printf("not logging content type %q", ct)
			}
		}

		if _, err := rr.Body.WriteTo(io.MultiWriter(out...)); err != nil {
			l.l.Printf("error sending response: %v", err)
		}

		if l.verbose {
			fmt.Fprint(l.o, "\n")
		}

		l.l.Printf("(client %s) %d %s", queryIP(r), rr.Code, http.StatusText(rr.Code))
	}
}
