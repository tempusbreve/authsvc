package authsvc // import "breve.us/authsvc"

import "net/http"

func RegisterAPI(verbose bool) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/oauth/", newOAuthHandler())
	mux.Handle("/api/", newAPIHandler())
	mux.Handle("/login/", newLoginHandler(nil))
	return mux
}

func newAPIHandler() http.Handler {
	return &api{}
}

type api struct {
}

func (a *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSONCode(http.StatusBadRequest, w, err.Error())
		return
	}

	switch p := r.URL.Path; p {
	case "/api/v4/user":
		a.handleUser(w, r)
	default:
		obj := struct {
			Url       string
			QueryKeys []string
			Values    interface{}
		}{Url: r.URL.String(), QueryKeys: queryKeys(r.Form), Values: r.Form}

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
