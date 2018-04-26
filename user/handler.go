package user // import "breve.us/authsvc/user"

import (
	"net/http"

	"github.com/gorilla/mux"

	"breve.us/authsvc/common"
)

// RegisterAPI returns a router for the api.
func RegisterAPI(root string, verbose bool) http.Handler {
	a := &handler{root: root}
	mx := mux.NewRouter()
	mx.Path(root).HandlerFunc(a.handleUser).Methods("GET")
	return mx
}

type handler struct {
	root string
}

func (a *handler) handleUser(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		common.JSONStatusResponse(http.StatusBadRequest, w, err.Error())
		return
	}

	u := &Details{
		ID:       1,
		Username: "johnweldon",
		Email:    "johnweldon4+hardcoded@gmail.com",
		Name:     "John Weldon",
		State:    "active",
	}

	common.JSONResponse(w, u)
}

// Details describes the information about a user.
type Details struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	State    string `json:"state"`
}
