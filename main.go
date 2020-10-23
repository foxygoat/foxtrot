package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

func main() {
	rm := newRoom()
	go rm.Run()

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.Handle("/ws", rm)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// operation is a client operation on a chat room
type operation struct {
	payload []byte
}

// room is a chat room with multiple clients. Any message received from
// a client is sent to all the other clients.
type room struct {
	clients sync.Map
	in      chan operation
}

func newRoom() *room {
	return &room{
		in: make(chan operation),
	}
}

func (rm *room) Run() {
	for op := range rm.in {
		// receive an op from the input channel, send it to every
		// clients' output channel.
		rm.clients.Range(func(cl, _ interface{}) bool {
			select {
			case cl.(*client).out <- op:
			case <-cl.(*client).done:
			}
			return true
		})
	}
}

func (rm *room) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cl, err := newClient(w, r)
	if err != nil {
		return
	}

	go func() { cl.out <- operation{payload: []byte("Welcome to foxtrot!")} }()
	rm.clients.Store(cl, nil)
	cl.Run(rm.in)
	rm.clients.Delete(cl)
}

// client is a chat connection for a client.
type client struct {
	conn *websocket.Conn
	out  chan operation
	done chan struct{}
}

func newClient(w http.ResponseWriter, r *http.Request) (*client, error) {
	u := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := u.Upgrade(w, r, nil)

	if err != nil {
		log.Printf("upgrader.Upgrade: %v", err)
		return nil, err
	}

	cl := &client{
		conn: conn,
		out:  make(chan operation),
		done: make(chan struct{}),
	}

	return cl, nil
}

func (cl *client) Run(in chan<- operation) {
	go cl.readInput(in)
	go cl.sendOutput()

	<-cl.done // wait for channel to be closed

	close(cl.out)
	cl.conn.Close()
}

func (cl *client) readInput(in chan<- operation) {
	for {
		_, p, err := cl.conn.ReadMessage()
		if err != nil {
			log.Printf("conn.ReadMessage: %v", err)
			cl.Done()
			return
		}
		in <- operation{p}
	}
}

func (cl *client) sendOutput() {
	for op := range cl.out {
		err := cl.conn.WriteMessage(websocket.TextMessage, op.payload)
		if err != nil {
			log.Printf("conn.WriteMessage: %v", err)
			cl.Done()
			return
		}
	}
}

func (cl *client) Done() {
	select {
	case <-cl.done: // already closed
	default:
		close(cl.done)
	}
}
