package sshd

import (
	"io"

	"github.com/gliderlabs/ssh"
)

type Recorder interface {
	Start(ssh.Session) (io.WriteCloser, error)
}

type DummyRecorder struct{}

func (r *DummyRecorder) Start(ssh.Session) (io.WriteCloser, error) {
	return r, nil
}

func (r *DummyRecorder) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (r *DummyRecorder) Close() error {
	return nil
}
