package foxtrot

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegisterLogin(t *testing.T) {
	db := mustDB()
	defer db.close()

	secret := []byte("$$$$$hhh!")
	a := authenticator{db: db, secret: secret}
	u := &User{Name: "Alice"}
	err := a.register(context.Background(), u, "Pa$$w0rd")
	require.NoError(t, err)

	u2, err := a.login(context.Background(), "Alice", "Pa$$w0rd")
	u.jwt = u2.jwt
	require.NoError(t, err)
	require.Equal(t, u, u2)
	require.NoError(t, a.validateJWT(u.jwt))

	_, err = a.login(context.Background(), "Alice", "WRONG-PASS")
	require.Error(t, err)
	require.Truef(t, errors.Is(err, errAuth), "want errAuth, got %v", err)
}

func TestRegisterErr(t *testing.T) {
	db := mustDB()
	defer db.close()

	secret := []byte("$$$$$hhh!")
	a := authenticator{db: db, secret: secret}

	u := &User{Name: "Alice"}
	err := a.register(context.Background(), u, "Pa$$w0rd")
	require.NoError(t, err)

	err = a.register(context.Background(), u, "Pa$$w0rd")
	require.Error(t, err)
}

func TestLoginErr(t *testing.T) {
	db := mustDB()
	defer db.close()

	a := authenticator{db: db}
	_, err := a.login(context.Background(), "Alice", "Pa$$w0rd")
	require.Error(t, err) // missing user
	require.Truef(t, errors.Is(err, errAuth), "want errAuth, got %v", err)
}