package foxtrot

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func requireErrIs(t *testing.T, err, target error) {
	t.Helper()
	require.Error(t, err)
	require.Error(t, target)
	require.Truef(t, errors.Is(err, target), "want %v, got %v", target, err)
}
