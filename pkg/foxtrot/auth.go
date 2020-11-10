package foxtrot

import (
	"context"
	"errors"

	"foxygo.at/s/errs"
	"golang.org/x/crypto/bcrypt"
)

var (
	errAuth         = errors.New("user authentication error")
	errPasswordHash = errors.New("password hash creation err")
)

type authenticator struct {
	db *db
}

func (a *authenticator) register(ctx context.Context, u *User, password string) error {
	hashWithSalt, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errs.New(errPasswordHash, err)
	}
	u.passwordHash = string(hashWithSalt)
	return a.db.createUser(ctx, u)
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
	return u, nil
}
