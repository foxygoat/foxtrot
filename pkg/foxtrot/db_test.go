package foxtrot

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

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

func mustDB() *db {
	db, err := newDB(":memory:")
	if err != nil {
		panic(err)
	}
	return db
}

func TestCreateGetUser(t *testing.T) {
	db := mustDB()
	defer db.close()

	u := &User{Name: "alice", passwordHash: "###"}
	err := db.createUser(context.Background(), u)
	require.NoError(t, err)

	u2, err := db.getUser(context.Background(), "alice")
	require.NoError(t, err)
	require.Equal(t, u, u2)
}

func TestGetUserErr(t *testing.T) {
	db := mustDB()
	defer db.close()

	_, err := db.getUser(context.Background(), "MISSING")
	require.Error(t, err)
}

func TestCreateUserErr(t *testing.T) {
	db := mustDB()
	defer db.close()

	err := db.createUser(context.Background(), &User{Name: "alice"})
	require.Error(t, err) // missing hash

	err = db.createUser(context.Background(), &User{passwordHash: "##"})
	require.Error(t, err) // missing name
}

func TestCreateGetRoom(t *testing.T) {
	db := mustDB()
	defer db.close()

	r := &Room{Name: "kitchen"}
	err := db.createRoom(context.Background(), r)
	require.NoError(t, err)

	r2, err := db.getRoom(context.Background(), "kitchen")
	require.NoError(t, err)
	require.Equal(t, r, r2)
}

func TestGetRoomErr(t *testing.T) {
	db := mustDB()
	defer db.close()

	_, err := db.getRoom(context.Background(), "MISSING-SHED")
	require.Error(t, err)
}

func TestCreateRoomErr(t *testing.T) {
	db := mustDB()
	defer db.close()

	err := db.createRoom(context.Background(), &Room{Name: ""})
	require.Error(t, err) // room name cannot be empty
}

func TestCreateQueryMessageSimple(t *testing.T) {
	db := mustDB()
	defer db.close()

	ctx := context.Background()
	require.NoError(t, db.createRoom(ctx, &Room{Name: "kitchen"}))
	require.NoError(t, db.createUser(ctx, &User{Name: "alice", passwordHash: "###"}))

	m := Message{Content: "hi", Room: "kitchen", Author: "alice", CreatedAt: now()}
	err := db.createMessage(ctx, &m)
	require.NoError(t, err)

	messages, err := db.queryMessages(ctx, "kitchen", -1, -1)
	require.NoError(t, err)
	require.Equal(t, 1, len(messages))

	got := messages[0]
	createdAt, err := time.Parse(time.RFC3339, got.CreatedAt)
	require.NoError(t, err)
	require.WithinDuration(t, time.Now(), createdAt, time.Second)

	want := &Message{ID: 101, Content: "hi", Room: "kitchen", Author: "alice", CreatedAt: got.CreatedAt}
	require.Equal(t, want, got)

	messages, err = db.queryMessages(ctx, "kitchen", 110, -1)
	require.NoError(t, err)
	require.Equal(t, 1, len(messages))
	require.Equal(t, want, messages[0])

	messages, err = db.queryMessages(ctx, "kitchen", -1, 10)
	require.NoError(t, err)
	require.Equal(t, 1, len(messages))
	require.Equal(t, want, messages[0])

	messages, err = db.queryMessages(ctx, "kitchen", 110, 10)
	require.NoError(t, err)
	require.Equal(t, 1, len(messages))
	require.Equal(t, want, messages[0])
}

func TestCreateQueryMessage(t *testing.T) {
	db := mustDB()
	defer db.close()

	ctx := context.Background()
	require.NoError(t, db.createRoom(ctx, &Room{Name: "kitchen"}))
	require.NoError(t, db.createRoom(ctx, &Room{Name: "shed"}))
	require.NoError(t, db.createUser(ctx, &User{Name: "alice", passwordHash: "###"}))
	require.NoError(t, db.createUser(ctx, &User{Name: "bob", passwordHash: "***"}))

	messages := []Message{
		{Content: "ouch", Room: "shed", Author: "bob", CreatedAt: now()},
		{Content: "hi", Room: "kitchen", Author: "bob", CreatedAt: now()},
		{Content: "hungry?", Room: "kitchen", Author: "alice", CreatedAt: now()},
		{Content: "🍪🧁", Room: "shed", Author: "alice", CreatedAt: now()},
	}
	for _, message := range messages {
		m := message
		require.NoError(t, db.createMessage(ctx, &m))
	}
	got, err := db.queryMessages(ctx, "kitchen", -1, -1)
	require.NoError(t, err)
	require.Equal(t, 2, len(got))
	require.Equal(t, "hungry?", got[0].Content)
	require.Equal(t, "hi", got[1].Content)

	got, err = db.queryMessages(ctx, "shed", 102, -1)
	require.NoError(t, err)
	require.Equal(t, 1, len(got))
	require.Equal(t, "ouch", got[0].Content)

	got, err = db.queryMessages(ctx, "shed", -1, 2)
	require.NoError(t, err)
	require.Equal(t, 2, len(got))
	require.Equal(t, "🍪🧁", got[0].Content)
	require.Equal(t, "ouch", got[1].Content)
}

func TestCreateMessageErr(t *testing.T) {
	db := mustDB()
	defer db.close()

	ctx := context.Background()
	require.NoError(t, db.createRoom(ctx, &Room{Name: "kitchen"}))
	require.NoError(t, db.createUser(ctx, &User{Name: "alice", passwordHash: "###"}))

	m := Message{Content: "bla", Author: "alice", Room: "kitchen", CreatedAt: now()}
	err := db.createMessage(ctx, &m)
	require.NoError(t, err)

	m = Message{Content: "", Author: "alice", Room: "kitchen", CreatedAt: now()}
	err = db.createMessage(ctx, &m)
	require.Error(t, err)

	m = Message{Content: "bla", Author: "MISSING", Room: "kitchen", CreatedAt: now()}
	err = db.createMessage(ctx, &m)
	require.Error(t, err)

	m = Message{Content: "bla", Author: "alice", Room: "MISSING", CreatedAt: now()}
	err = db.createMessage(ctx, &m)
	require.Error(t, err)

	m = Message{Content: "bla", Author: "alice", Room: "kitchen"}
	err = db.createMessage(ctx, &m)
	require.Error(t, err)
}
