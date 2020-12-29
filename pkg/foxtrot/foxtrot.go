// Package foxtrot provides core foxtrot data structures.
package foxtrot

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"
)

// Config contains the DB and Authenticator configuration as kong
// struct tags to use as command line flags.
type Config struct {
	DSN        string `help:"SQLite Data Source Name" default:":memory:"`
	AuthSecret string `help:"JWT token generation secret. default: <RANDOM_STRING>" env:"FT_AUTH_SECRET"`

	Version Version `kong:"-"`
}

// Version contains build version information which is returned as JSON
// at the version HTTP endpoint.
type Version struct {
	CommitSha string `json:"commitSha"`
	Semver    string `json:"semver"`
}

// NewApp creates a new App struct for given config and wire it with
// given mux on /api.
func NewApp(cfg *Config, mux *http.ServeMux) (*App, error) {
	db, err := newDB(cfg.DSN)
	if err != nil {
		return nil, err
	}
	secret := []byte(cfg.AuthSecret)

	if len(secret) == 0 {
		secret = make([]byte, 16)
		if _, err := rand.Read(secret); err != nil {
			return nil, fmt.Errorf("NewApp auth secret initialisation: %w", err)
		}
	}
	auth := &authenticator{db: db, secret: secret}
	api := newAPI(db, auth, cfg.Version)
	api.wireRoutes("/api", mux)
	app := &App{db: db, auth: auth, api: api}
	return app, nil
}

// App is a top level data structure containing all relevant foxtrot parts,
// including database, authenticator and HTTP api structs.
type App struct {
	db   *db
	auth *authenticator
	api  *api
}

// User is a core foxtrot data structure representing a user entry in
// the database as well as the data related to a user as presented in
// the web UI.
type User struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatarURL,omitempty"`
	JWT       string `json:"jwt,omitempty"`

	passwordHash string
	avatar       []byte
}

// Room is a chat room identified by its name.
type Room struct {
	Name string `json:"name"`
}

// Message is a chat message.
type Message struct {
	ID        int    `json:"id"` // ordered by creation time
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
	Room      string `json:"room"`
	Author    string `json:"author"`
}

func now() string {
	return time.Now().Format(time.RFC3339)
}
