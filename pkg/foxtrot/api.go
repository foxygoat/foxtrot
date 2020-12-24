package foxtrot

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"foxygo.at/s/errs"
	"foxygo.at/s/httpe"
)

// api is a REST inspired HTTP API for accessing foxtrot data,
// registration and authentication
//
// /api/login POST
// /api/register POST
// /api/history?room=NAME[&before=MESSAGE_ID|TIMESTAMP&count=N]
//
// Not yet implemented:
// /api/user/NAME/
// /api/user/NAME/avatar
// /api/room/NAME # create new room.
type api struct {
	db   *db
	auth *authenticator
}

func newAPI(db *db, auth *authenticator) *api {
	a := &api{
		db:   db,
		auth: auth,
	}
	return a
}

func (a *api) wireRoutes(basePath string, mux *http.ServeMux) {
	mux.Handle(basePath+"/login", httpe.Must(httpe.Post, a.login))
	mux.Handle(basePath+"/register", httpe.Must(httpe.Post, a.register))
	mux.Handle(basePath+"/history", httpe.Must(httpe.Get, a.history))
	mux.Handle(basePath+"/_test_cleanup", httpe.Must(httpe.Delete, a.testCleanup))
}

type creds struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (a *api) login(w http.ResponseWriter, r *http.Request) error {
	c := creds{}
	defer r.Body.Close() //nolint: errcheck
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		return errs.Errorf("%v: JSON parse error: %v", httpe.ErrBadRequest, err)
	}
	u, err := a.auth.login(r.Context(), c.Name, c.Password)
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Bearer realm="Write access to foxtrot chat"`)
		return httpe.ErrUnauthorized
	}
	return json.NewEncoder(w).Encode(u)
}

func (a *api) register(w http.ResponseWriter, r *http.Request) error {
	c := creds{}
	defer r.Body.Close() //nolint: errcheck
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		return errs.Errorf("%v: JSON parse error: %v", httpe.ErrBadRequest, err)
	}
	u := User{Name: c.Name}
	if err := a.auth.register(r.Context(), &u, c.Password); err != nil {
		if errors.Is(err, errDBDuplicate) {
			return httpe.ErrBadRequest
		}
		return httpe.ErrInternalServerError
	}
	return json.NewEncoder(w).Encode(u)
}

func (a *api) history(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	room := q.Get("room")
	before := q.Get("before")
	beforeID := -1
	if before != "" {
		var err error
		if beforeID, err = strconv.Atoi(before); err != nil {
			return httpe.ErrBadRequest
		}
	}
	limit := 200
	messages, err := a.db.queryMessages(r.Context(), room, beforeID, limit)
	if err != nil {
		return httpe.ErrInternalServerError
	}
	return json.NewEncoder(w).Encode(messages)
}

const testUser = "$user"

func (a *api) testCleanup(_ http.ResponseWriter, r *http.Request) error {
	return a.db.deleteUser(r.Context(), testUser)
}
