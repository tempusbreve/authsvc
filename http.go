package authsvc // import "breve.us/authsvc"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func queryKeys(v url.Values) []string {
	var keys []string
	for k, _ := range v {
		keys = append(keys, k)
	}
	return keys
}

func writeRedirect(w http.ResponseWriter, uri string, additionalParams map[string]string) {
	dest, err := url.Parse(uri)
	if err != nil {
		writeJSONCode(http.StatusInternalServerError, w, err)
		return
	}
	q := dest.Query()
	for k, v := range additionalParams {
		q.Set(k, v)
	}
	dest.RawQuery = q.Encode()
	w.Header().Set("Location", dest.String())
	w.WriteHeader(http.StatusTemporaryRedirect)
	return
}

func writeJSON(w http.ResponseWriter, o interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if err := json.NewEncoder(w).Encode(o); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%v"}`, err)
	}
}

func writeJSONCode(code int, w http.ResponseWriter, o interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	if o != nil {
		if err := json.NewEncoder(w).Encode(o); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%v"}`, err)
		}
	}
}