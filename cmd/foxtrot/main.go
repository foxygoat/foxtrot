package main

import (
	"log"
	"net/http"

	"foxygo.at/foxtrot/pkg/foxtrot"
	"github.com/alecthomas/kong"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// Semver holds the version exposed at api/version and passed via linker flag on CI.
	Semver = "undefined"
	// CommitSha holds the commit exposed at api/version and passed via linker flag on CI.
	CommitSha = "undefined"
)

func main() {
	cfg := &foxtrot.Config{
		Version: foxtrot.Version{Semver: Semver, CommitSha: CommitSha, App: "foxtrot"},
	}
	kong.Parse(cfg, kong.Description("Foxtrot Server"))

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("frontend/public")))
	if _, err := foxtrot.NewApp(cfg, mux); err != nil {
		log.Fatal(err)
	}

	port := ":8080"
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(port, handler(logHTTP(mux))); err != nil {
		log.Fatal(err)
	}
}

func handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws" {
			socketHandler(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func logHTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		h.ServeHTTP(ww, r)
		log.Printf("%d %-4s %s %s\n", ww.statusCode, r.Method, r.URL, r.RemoteAddr)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// upgrader holds the websocket connection.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// socketHandler echos websocket messages back to the client.
func socketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrader.Upgrade: %v", err)
		return
	}
	defer conn.Close() //nolint: errcheck

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("conn.ReadMessage: %v", err)
			return
		}
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Printf("conn.WriteMessage: %v", err)
			return
		}
	}
}
