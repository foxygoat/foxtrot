package foxtrot

import (
	"context"
	"database/sql"
	"errors"

	"foxygo.at/s/errs"
	"github.com/mattn/go-sqlite3"
)

var (
	errDBInitialisation = errors.New("db: initialisation error")
	errDBInternal       = errors.New("db: internal error")
	errDBNotFound       = errors.New("db: entry not found")
	errDBDuplicate      = errors.New("db: duplicate")
)

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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.Errorf("%s: getUser '%s': %v", errDBNotFound, name, err)
		}
		return nil, errs.New(errDBInternal, err)
	}
	return &u, nil
}

func (db *db) createUser(ctx context.Context, u *User) error {
	stmt := "INSERT INTO users(name, password_hash, avatar) VALUES (?, ?, ?)"
	if _, err := db.conn.ExecContext(ctx, stmt, u.Name, u.passwordHash, u.avatar); err != nil {
		sqliteErr := &sqlite3.Error{}
		if errors.As(err, sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
			return errs.Errorf("%v: user '%s': %v", errDBDuplicate, u.Name, err)
		}
		return errs.Errorf("%v: cannot create user '%s': %v", errDBInternal, u.Name, err)
	}
	return nil
}

func (db *db) deleteUser(ctx context.Context, name string) error {
	stmt := "DELETE FROM users WHERE name = ?"
	result, err := db.conn.ExecContext(ctx, stmt, name)
	if err != nil {
		return errs.Errorf("%v: cannot delete user '%s': %v", errDBInternal, name, err)
	}
	cnt, err := result.RowsAffected()
	if err != nil {
		return errs.Errorf("%v: cannot confirm deletion of user '%s': %v", errDBInternal, name, err)
	}
	if cnt == 0 {
		return errs.Errorf("%v: cannot delete user '%s'", errDBNotFound, name)
	}
	return nil
}

func (db *db) getRoom(ctx context.Context, name string) (*Room, error) {
	r := Room{Name: name}
	stmt := "SELECT name FROM rooms WHERE name = ?"
	err := db.conn.QueryRowContext(ctx, stmt, name).Scan(&r.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.Errorf("%s: getRoom '%s': %v", errDBNotFound, name, err)
		}
		return nil, errs.New(errDBInternal, err)
	}
	return &r, nil
}

func (db *db) createRoom(ctx context.Context, r *Room) error {
	stmt := "INSERT INTO rooms(name) VALUES (?)"
	if _, err := db.conn.ExecContext(ctx, stmt, r.Name); err != nil {
		sqliteErr := &sqlite3.Error{}
		if errors.As(err, sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
			return errs.Errorf("%v: room '%s': %v", errDBDuplicate, r.Name, err)
		}
		return errs.Errorf("%v: cannot create room '%s': %v", errDBInternal, r.Name, err)
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
		return nil, errs.Errorf("%v: QueryContext messages for room '%s' before '%d': %v", errDBInternal, room, beforeID, err)
	}
	defer rows.Close() //nolint:errcheck
	messages, err := rowsToMessages(rows)
	if err != nil {
		return nil, errs.Errorf("%v: rowsToMessages for room '%s' before '%d': %v", errDBInternal, room, beforeID, err)
	}
	return messages, nil
}

func rowsToMessages(rows *sql.Rows) ([]*Message, error) {
	messages := []*Message{}
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.Content, &m.CreatedAt, &m.Room, &m.Author); err != nil {
			return nil, errs.Errorf("scan row: %v", err)
		}
		messages = append(messages, &m)
	}
	return messages, nil
}

func (db *db) createMessage(ctx context.Context, m *Message) error {
	stmt := "INSERT INTO messages(content, created_at, room, author) VALUES (?, ?, ?, ?)"
	if _, err := db.conn.ExecContext(ctx, stmt, m.Content, m.CreatedAt, m.Room, m.Author); err != nil {
		return errs.Errorf("%v: cannot create message '%#v': %v", errDBInternal, m, err)
	}
	return nil
}
