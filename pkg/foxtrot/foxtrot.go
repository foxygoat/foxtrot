// Package foxtrot provides core foxtrot data structures.
package foxtrot

import "time"

// User is a core foxtrot data structure representing a user entry in
// the database as well as the data related to a user as presented in
// the web UI.
type User struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatarURL,omitempty"`

	passwordHash string
	avatar       []byte
	jwt          string
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
