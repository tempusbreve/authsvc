package authsvc // import "breve.us/authsvc"

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterAPI returns a router for the api.
func RegisterAPI(root string, verbose bool) http.Handler {
	mx := mux.NewRouter()
	mx.Handle(root, &api{root: root})
	return mx
}

type api struct {
	root string
}

func (a *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSONCode(http.StatusBadRequest, w, err.Error())
		return
	}

	switch p := r.URL.Path; p {
	case a.root + "v4/user":
		a.handleUser(w, r)
	default:
		obj := struct {
			URL       string
			QueryKeys []string
			Values    interface{}
		}{URL: r.URL.String(), QueryKeys: queryKeys(r.Form), Values: r.Form}

		writeJSON(w, obj)
	}
}

func (a *api) handleUser(w http.ResponseWriter, r *http.Request) {
	u := &user{
		ID:       1,
		Username: "johnweldon",
		Email:    "johnweldon4+hardcoded@gmail.com",
		Name:     "John Weldon",
		State:    "active",
	}

	writeJSON(w, u)
}

type user struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	State    string `json:"state"`
}
