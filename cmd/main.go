package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/gliderlabs/ssh"
	sshd "github.com/inoc603/go-sshd"
	"github.com/inoc603/go-sshd/asciicast"
	"github.com/inoc603/go-sshd/auth"
	"github.com/inoc603/go-sshd/storage"
	"github.com/pkg/errors"
)

func exitOnErr(err error, msg string) {
	if err != nil {
		logrus.WithError(err).Fatalln(msg)
	}
}

func main() {
	pkAuth, err := auth.NewLocalPublicKeyAuth("/root/.ssh/authorized_keys")
	exitOnErr(err, "Failed to create publick key auth")

	store, err := storage.NewFileStorage("tmp/output")
	exitOnErr(err, "Failed to create output storage")

	newAsciinemaRecorder := func(s ssh.Session) (sshd.Recorder, error) {
		output, err := store.New(s.Context().(ssh.Context))
		if err != nil {
			return nil, errors.Wrap(err, "open output")
		}
		return asciicast.NewRecorder(output), nil
	}

	server, err := sshd.NewServer(
		sshd.WithAddress(":2222"),
		sshd.WithAuth(pkAuth),
		sshd.WithAuth(auth.NewPamPasswordAuth("passwd")),
		sshd.WithUserStore(&auth.LocalUserStore{}),
		sshd.WithHostFile("/etc/ssh/ssh_host_rsa_key"),
		sshd.WithRecorder(newAsciinemaRecorder),
	)
	exitOnErr(err, "Failed to create ssh server")

	exitOnErr(server.Start(), "Server stopped")
}
