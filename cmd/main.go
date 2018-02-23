package main

import (
	"github.com/Sirupsen/logrus"
	sshd "github.com/inoc603/go-sshd"
	"github.com/inoc603/go-sshd/auth"
	"github.com/inoc603/go-sshd/storage"
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

	server, err := sshd.NewServer(
		sshd.WithAddress(":2222"),
		sshd.WithAuth(pkAuth),
		sshd.WithAuth(auth.NewPamPasswordAuth("passwd")),
		sshd.WithUserStore(&auth.LocalUserStore{}),
		sshd.WithHostFile("/etc/ssh/ssh_host_rsa_key"),
		sshd.WithStorage(store),
	)
	exitOnErr(err, "Failed to create ssh server")

	exitOnErr(server.Start(), "Server stopped")
}
