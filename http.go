package authsvc // import "breve.us/authsvc"

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func queryKeys(v url.Values) []string {
	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	return keys
}

func writeRedirect(w http.ResponseWriter, r *http.Request, uri string, additionalParams map[string]string) {
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
	http.Redirect(w, r, dest.String(), http.StatusSeeOther)
}

func writeJSON(w http.ResponseWriter, o interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(o); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": "%v"}`, err)
	}
}

func writeJSONCode(code int, w http.ResponseWriter, o interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if o != nil {
		if err := json.NewEncoder(w).Encode(o); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%v"}`, err)
		}
	}
}
