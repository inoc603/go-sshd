package storage

import (
	"io"

	"github.com/gliderlabs/ssh"
)

type Storage interface {
	New(ctx ssh.Context) (io.WriteCloser, error)
}
