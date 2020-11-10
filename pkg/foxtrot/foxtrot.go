// Package foxtrot provides core foxtrot data structures.
package foxtrot

// User is a core foxtrot data structure representing a user entry in
// the database as well as the data related to a user as presented in
// the web UI.
type User struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatarURL,omitempty"`

	passwordHash string
	avatar       []byte
}
