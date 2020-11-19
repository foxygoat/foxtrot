package foxtrot

import (
	"context"
	"errors"
	"time"

	"foxygo.at/s/errs"
	"golang.org/x/crypto/bcrypt"
)

var (
	errAuth         = errors.New("user authentication error")
	errPasswordHash = errors.New("password hash creation err")
)

type authenticator struct {
	db     *db
	secret []byte
}

func (a *authenticator) register(ctx context.Context, u *User, password string) error {
	hashWithSalt, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errs.New(errPasswordHash, err)
	}
	u.passwordHash = string(hashWithSalt)
	if err := a.db.createUser(ctx, u); err != nil {
		return err
	}
	u.jwt = a.newJWT(u.Name)
	return nil
}

func (a *authenticator) login(ctx context.Context, name, password string) (*User, error) {
	u, err := a.db.getUser(ctx, name)
	if err != nil {
		_ = bcrypt.CompareHashAndPassword([]byte(""), []byte(password))
		return nil, errs.New(errAuth, err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.passwordHash), []byte(password)); err != nil {
		return nil, errs.New(errAuth, err)
	}
	u.jwt = a.newJWT(u.Name)
	return u, nil
}

func (a *authenticator) newJWT(sub string) string {
	// Arbitrarily chosen expiry of three months
	exp := time.Now().AddDate(0, 3, 0).Unix()
	return newJWT(sub, exp, a.secret)
}

func (a *authenticator) validateJWT(jwt string) error {
	return validateJWT(jwt, a.secret)
}
