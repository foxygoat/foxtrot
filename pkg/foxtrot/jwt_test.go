package foxtrot

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewJWT(t *testing.T) {
	header := "eyJhbGciOiJIUzI1NiJ9."

	payload := "eyJzdWIiOiJnb2F0IiwiZXhwIjoxNjEzNTU1MjI3fQ."
	signature := "ATD0jdWI-5droEwBOLqlRh-958cuKxFTGxkMLTp_7_A"
	want := header + payload + signature
	got := newJWT("goat", 1613555227, []byte("SECRET"))
	require.Equal(t, want, got)

	signature2 := "fMDHBsSwemtFhEVhwUS34TRNP-6rgO3paOOQAOQcKkY"
	want = header + payload + signature2
	got = newJWT("goat", 1613555227, nil)
	require.Equal(t, want, got)

	signature3 := "XaCBhddqPOvmX-2oJyVD2qjXEl9tSuHKSF10jCRbOVY"
	payload3 := "eyJzdWIiOiJGT1giLCJleHAiOjE2MTM1NTUyMjd9."
	want = header + payload3 + signature3
	got = newJWT("FOX", 1613555227, nil)
	require.Equal(t, want, got)
}

func TestValidateJWT(t *testing.T) {
	secret := []byte("SECRET")
	j := "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJnb2F0IiwiZXhwIjoxNjEzNTU1MjI3fQ.ATD0jdWI-5droEwBOLqlRh-958cuKxFTGxkMLTp_7_A"
	require.NoError(t, validateJWT(j, secret))

	errIs(t, validateJWT(j, nil), errJWTSignature)
	errIs(t, validateJWT("", nil), errJWT)
	errIs(t, validateJWT(j[1:], secret), errJWT)

	j = "BAD-HEADER.abcd.-IpzLuSB2BxkeP3IKo-Rs-6fdjbKaPnd15t73nUs-cU"
	errIs(t, validateJWT(j, secret), errJWT)

	j = "eyJhbGciOiJIUzI1NiJ9.NOT-b64-💥.vZry1kTJ_kGEOGxlBFgZIrzBpVD9yI2gn1P8SRHNf1w"
	errIs(t, validateJWT(j, secret), errJWTEncoding)

	p := base64.RawStdEncoding.EncodeToString([]byte(`{ "BAD JSON`))
	j = "eyJhbGciOiJIUzI1NiJ9." + p + ".W7Ity7twIM0iPJqjjTQ0r3sPyfFj5v5l81eIdFB5jVg"
	errIs(t, validateJWT(j, secret), errJWTEncoding)

	pastExp := time.Now().Add(-10 * time.Second).Unix()
	j = newJWT("fox", pastExp, secret)
	errIs(t, validateJWT(j, secret), errJWTExpired)
}

func errIs(t *testing.T, err, targetErr error) {
	t.Helper()
	require.Error(t, err)
	requireErrIs(t, err, targetErr)
}
