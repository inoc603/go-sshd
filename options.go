package sshd

import (
	"github.com/inoc603/go-sshd/auth"
	"github.com/pkg/errors"
)

func WithAddress(addr string) Option {
	return func(s *Server) error {
		s.addr = addr
		return nil
	}
}

func WithAuth(a interface{}) Option {
	return func(s *Server) error {
		if pkAuth, ok := a.(auth.PublicKeyAuth); ok {
			s.pkAuth = append(s.pkAuth, pkAuth)
			return nil
		}

		if pwAuth, ok := a.(auth.PasswordAuth); ok {
			s.pwAuth = append(s.pwAuth, pwAuth)
			return nil
		}

		return errors.Errorf("invalid auth middleware")
	}
}

func WithHostFile(f string) Option {
	return func(s *Server) error {
		s.hostkeyFile = f
		return nil
	}
}

func WithUserStore(us auth.UserStore) Option {
	return func(s *Server) error {
		s.userStore = us
		return nil
	}
}

func WithRecorder(r RecorderFactory) Option {
	return func(s *Server) error {
		s.getRecorder = r
		return nil
	}
}
