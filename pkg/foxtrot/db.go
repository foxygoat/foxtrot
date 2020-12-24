package foxtrot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"foxygo.at/s/errs"
)

var errDBInitialisation = errors.New("database initialisation error")

type db struct {
	conn *sql.DB
}

func newDB(dsnURI string) (*db, error) {
	conn, err := sql.Open("sqlite3", dsnURI)
	if err != nil {
		return nil, errs.New(errDBInitialisation, err)
	}

	db := &db{conn: conn}
	if err := db.setupSchema(); err != nil {
		db.close()
		return nil, err
	}

	return db, nil
}

func (db *db) close() {
	_ = db.conn.Close()
}

func (db *db) setupSchema() error {
	// Enforce foreign key constraints
	if _, err := db.conn.Exec("PRAGMA foreign_keys=1"); err != nil {
		return errs.New(errDBInitialisation, err)
	}
	selectVersionStr := "SELECT version FROM schema"
	version := ""
	err := db.conn.QueryRow(selectVersionStr).Scan(&version)
	expectedVersion := "v0.0.1"
	if err == nil && version != expectedVersion {
		return errs.Errorf("%v: bad version '%s' expected '%s'", errDBInitialisation, version, expectedVersion)
	} else if err == nil {
		return nil
	}
	// Assume an error means the schema has not been created, so create it now.
	if _, err := db.conn.Exec(schema); err != nil {
		return errs.New(errDBInitialisation, err)
	}
	if _, err := db.conn.Exec(sampleData); err != nil {
		return errs.New(errDBInitialisation, err)
	}
	return nil
}

func (db *db) getUser(ctx context.Context, name string) (*User, error) {
	u := User{Name: name}
	stmt := "SELECT password_hash FROM users WHERE name = ?"
	err := db.conn.QueryRowContext(ctx, stmt, name).Scan(&u.passwordHash)
	if err != nil {
		return nil, fmt.Errorf("db: cannot get user '%s': %w", name, err)
	}
	return &u, nil
}

func (db *db) createUser(ctx context.Context, u *User) error {
	stmt := "INSERT INTO users(name, password_hash, avatar) VALUES (?, ?, ?)"
	if _, err := db.conn.ExecContext(ctx, stmt, u.Name, u.passwordHash, u.avatar); err != nil {
		return fmt.Errorf("db: cannot create user '%s': %w", u.Name, err)
	}
	return nil
}

func (db *db) getRoom(ctx context.Context, name string) (*Room, error) {
	r := Room{Name: name}
	stmt := "SELECT name FROM rooms WHERE name = ?"
	err := db.conn.QueryRowContext(ctx, stmt, name).Scan(&r.Name)
	if err != nil {
		return nil, fmt.Errorf("db: cannot get room '%s': %w", name, err)
	}
	return &r, nil
}

func (db *db) createRoom(ctx context.Context, r *Room) error {
	stmt := "INSERT INTO rooms(name) VALUES (?)"
	if _, err := db.conn.ExecContext(ctx, stmt, r.Name); err != nil {
		return fmt.Errorf("db: cannot create room '%s': %w", r.Name, err)
	}
	return nil
}

// queryMessages returns a list Messages for given room. A maximum of
// limit messages is returned, or all messages if limit is set to -1
// Only messages the came before given beforeID are returned or messages
// up until the most recent one if beforeID is -1.
func (db *db) queryMessages(ctx context.Context, room string, beforeID, limit int) ([]*Message, error) {
	stmt := `SELECT id, content, created_at, room, author FROM messages WHERE room = ?`
	args := []interface{}{room}
	if beforeID != -1 {
		stmt += " AND id < ?"
		args = append(args, beforeID)
	}
	stmt += " ORDER BY id DESC"
	if limit != -1 {
		stmt += " LIMIT ?"
		args = append(args, limit)
	}
	rows, err := db.conn.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, fmt.Errorf("db: query messages for room '%s' before ID '%d': execute query: %w", room, beforeID, err)
	}
	defer rows.Close() //nolint:errcheck
	messages, err := rowsToMessages(rows)
	if err != nil {
		return nil, fmt.Errorf("db: query messages for room '%s' before '%d': %w", room, beforeID, err)
	}
	return messages, nil
}

func rowsToMessages(rows *sql.Rows) ([]*Message, error) {
	var messages []*Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Content, &m.CreatedAt, &m.Room, &m.Author); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		messages = append(messages, &m)
	}
	return messages, nil
}

func (db *db) createMessage(ctx context.Context, m *Message) error {
	stmt := "INSERT INTO messages(content, created_at, room, author) VALUES (?, ?, ?, ?)"
	if _, err := db.conn.ExecContext(ctx, stmt, m.Content, m.CreatedAt, m.Room, m.Author); err != nil {
		return fmt.Errorf("db: cannot create message '%#v': %w", m, err)
	}
	return nil
}
