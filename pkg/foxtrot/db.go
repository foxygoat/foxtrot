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
	return nil
}

func (db *db) getUser(ctx context.Context, name string) (*User, error) {
	u := User{Name: name}
	selectUserStmt := "SELECT password_hash FROM users WHERE name=?"
	err := db.conn.QueryRowContext(ctx, selectUserStmt, name).Scan(&u.passwordHash)
	if err != nil {
		return nil, fmt.Errorf("db: cannot get user '%s': %w", name, err)
	}
	return &u, nil
}

func (db *db) createUser(ctx context.Context, u *User) error {
	insertUserStr := "INSERT INTO users(name, password_hash, avatar) VALUES( ?, ?, ? )"
	if _, err := db.conn.ExecContext(ctx, insertUserStr, u.Name, u.passwordHash, u.avatar); err != nil {
		return fmt.Errorf("db: cannot create user '%s': %w", u.Name, err)
	}
	return nil
}
