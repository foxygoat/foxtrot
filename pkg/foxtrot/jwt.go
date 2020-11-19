// This file contains a cut down implementation of JWTs (JSON Web
// Tokens).  It only works with the HS256 algorithm (symmetric cypher).
// This JWT implementation is for a server that issues and validates its
// own tokens.

package foxtrot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"foxygo.at/s/errs"
)

var (
	encodedJWTHeader = jsonBase64Encode(map[string]string{"alg": "HS256"})

	errJWT          = errors.New("invalid JWT Token")
	errJWTEncoding  = fmt.Errorf("%w: bad encoding", errJWT)
	errJWTExpired   = fmt.Errorf("%w: expired", errJWT)
	errJWTSignature = fmt.Errorf("%w: invalid signature", errJWT)
)

type jwtPayload struct {
	Sub string `json:"sub"` // subject: user name
	Exp int64  `json:"exp"` // expiry in unix epoche seconds
}

func validateJWT(jwt string, secret []byte) error {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return errs.Errorf("%v: invalid format, expected 2 '.' got %d", errJWT, len(parts)-1)
	}
	expectedSignature := sign(parts[0]+"."+parts[1], secret)
	if parts[2] != expectedSignature {
		return errJWTSignature
	}
	if parts[0] != encodedJWTHeader {
		return errs.Errorf("%v: invalid JWT header", errJWT)
	}
	payload := parts[1]
	b, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return errs.Errorf("%v: base64: %v", errJWTEncoding, err)
	}
	p := jwtPayload{}
	if err := json.Unmarshal(b, &p); err != nil {
		return errs.Errorf("%v: json: %v", errJWTEncoding, err)
	}
	if p.Exp <= time.Now().Unix() {
		return errJWTExpired
	}
	return nil
}

func newJWT(sub string, exp int64, secret []byte) string {
	payload := jwtPayload{Sub: sub, Exp: exp}
	j := encodedJWTHeader + "." + jsonBase64Encode(payload)
	signature := sign(j, secret)
	return j + "." + signature
}

func sign(s string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	_, _ = h.Write([]byte(s))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func jsonBase64Encode(v interface{}) string {
	b, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(b)
}
