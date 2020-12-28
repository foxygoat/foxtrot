package foxtrot

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	cfg := &Config{DSN: ":memory:"}
	mux := http.NewServeMux()
	app, err := NewApp(cfg, mux)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotEmpty(t, app.auth.secret)
	secret := app.auth.secret

	mux = http.NewServeMux()
	app, err = NewApp(cfg, mux)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotEmpty(t, app.auth.secret)
	require.NotEqual(t, secret, app.auth.secret)

	cfg.AuthSecret = "$ecret"
	mux = http.NewServeMux()
	app, err = NewApp(cfg, mux)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.Equal(t, cfg.AuthSecret, string(app.auth.secret))
}

func TestNewAppErr(t *testing.T) {
	cfg := &Config{DSN: "file:MISSING.db?mode=ro"}
	mux := http.NewServeMux()
	_, err := NewApp(cfg, mux)
	require.Error(t, err)
	requireErrIs(t, err, errDBInitialisation)
}

func requireErrIs(t *testing.T, err, target error) {
	t.Helper()
	require.Error(t, err)
	require.Error(t, target)
	require.Truef(t, errors.Is(err, target), "want %v, got %v", target, err)
}
