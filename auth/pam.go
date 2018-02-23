package auth

import (
	"github.com/gliderlabs/ssh"
	"github.com/msteinert/pam"
	"github.com/pkg/errors"
)

type PamPasswordAuth struct {
	module string
}

func NewPamPasswordAuth(module string) *PamPasswordAuth {
	return &PamPasswordAuth{
		module: module,
	}
}

func (a *PamPasswordAuth) Auth(ctx ssh.Context, password string) bool {
	t, err := pam.StartFunc(a.module, ctx.User(), func(s pam.Style, msg string) (string, error) {
		switch s {
		case pam.PromptEchoOff:
			return password, nil
		}
		return "", errors.New("Unrecognized message style")
	})

	if err != nil {
		return false
	}

	return t.Authenticate(0) == nil
}
