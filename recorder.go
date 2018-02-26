package sshd

import "github.com/gliderlabs/ssh"

type RecorderFactory func(ssh.Session) (Recorder, error)

type Recorder interface {
	WriteInput(b []byte)
	WriteOutput(b []byte)
}

type DummyRecorder struct{}

func (r *DummyRecorder) WriteInput(b []byte)  {}
func (r *DummyRecorder) WriteOutput(b []byte) {}
