package user // import "breve.us/authsvc/user"

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"breve.us/authsvc/common"
)

// Options are user API handler options
type Options struct {
	Root    string
	Verbose bool
	Users   *Registry
}

// RegisterAPI returns a router for the api.
func RegisterAPI(opts Options) *mux.Router {
	a := &handler{opts: opts}
	mx := mux.NewRouter()
	mx.Path(opts.Root).HandlerFunc(a.handleUser).Methods("GET")
	if !strings.HasSuffix(opts.Root, "/") {
		mx.Path(opts.Root + "/").HandlerFunc(a.handleUser).Methods("GET")
	}
	return mx
}

type handler struct {
	opts Options
}

func (a *handler) handleUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		common.JSONStatusResponse(http.StatusBadRequest, w, err.Error())
		return
	}

	if username := common.GetUsername(r.Context()); username != "" {
		if u, err := a.opts.Users.Get(username); err == nil {
			common.JSONResponse(w, u.toFilteredMap())
			return
		}
	}

	common.JSONStatusResponse(http.StatusBadRequest, w, "not logged in")
}
