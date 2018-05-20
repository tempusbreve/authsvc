package common // import "breve.us/authsvc/common"

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
			log.Printf("error writing error code to response: %v", e2)
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
				log.Printf("error writing error code to response: %v", e2)
			}
		}
	}
}
