package foxtrot

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestNewDBErr(t *testing.T) {
	_, err := newDB("file:MISSING.db?mode=ro")
	require.Error(t, err)
}

func TestSchemaSetup(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "foxtrot-*.db")
	require.NoError(t, err)
	dsn := tmpfile.Name()
	require.NoError(t, tmpfile.Close())
	defer os.Remove(dsn) //nolint:errcheck

	db, err := newDB(dsn)
	require.NoError(t, err)
	db.close()

	db, err = newDB(dsn)
	require.NoError(t, err)

	_, err = db.conn.Exec("UPDATE schema SET version='invalid_version'")
	require.NoError(t, err)

	_, err = newDB(dsn)
	require.Error(t, err)
	require.Truef(t, errors.Is(err, errDBInitialisation), "want %v, got %v", errDBInitialisation, err)
}

func TestCreateGetUser(t *testing.T) {
	db, err := newDB("")
	require.NoError(t, err)
	defer db.close()

	u := &User{Name: "alice", passwordHash: "###"}
	err = db.createUser(context.Background(), u)
	require.NoError(t, err)

	u2, err := db.getUser(context.Background(), "alice")
	require.NoError(t, err)
	require.Equal(t, u, u2)
}

func TestGetUserErr(t *testing.T) {
	db, err := newDB("")
	require.NoError(t, err)
	defer db.close()

	_, err = db.getUser(context.Background(), "MISSING")
	require.Error(t, err)
}

func TestCreateUserErr(t *testing.T) {
	db, err := newDB("")
	require.NoError(t, err)
	defer db.close()

	err = db.createUser(context.Background(), &User{Name: "alice"})
	require.Error(t, err) // missing hash

	err = db.createUser(context.Background(), &User{passwordHash: "##"})
	require.Error(t, err) // missing name
}
