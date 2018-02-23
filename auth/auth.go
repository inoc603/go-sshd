package auth

import (
	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
)

type PublicKeyAuth interface {
	Auth(ctx ssh.Context, key ssh.PublicKey) bool
}

type PasswordAuth interface {
	Auth(ctx ssh.Context, password string) bool
}

type User struct {
	Name  string
	UID   uint32
	GID   uint32
	Home  string
	Shell string
}

type UserStore interface {
	Get(name string) (*User, error)
}

type DummyUserStore struct{}

func (us *DummyUserStore) Get(name string) (*User, error) {
	return nil, errors.Errorf("user %s not found", name)
}
