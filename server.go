package sshd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/Sirupsen/logrus"
	"github.com/gliderlabs/ssh"
	"github.com/inoc603/go-sshd/auth"
	"github.com/inoc603/go-sshd/pipe"
	"github.com/kr/pty"
	"github.com/pkg/errors"
)

type Option func(s *Server) error

type Server struct {
	addr        string
	hostkeyFile string
	userStore   auth.UserStore
	pkAuth      []auth.PublicKeyAuth
	pwAuth      []auth.PasswordAuth
	getRecorder RecorderFactory
}

func NewServer(opts ...Option) (*Server, error) {
	s := &Server{
		addr:      ":22",
		userStore: &auth.DummyUserStore{},
		getRecorder: func(ssh.Session) (Recorder, error) {
			return &DummyRecorder{}, nil
		},
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Server) authPublicKey(ctx ssh.Context, key ssh.PublicKey) bool {
	for _, a := range s.pkAuth {
		if a.Auth(ctx, key) {
			return true
		}
	}
	return false
}

func (s *Server) authPassword(ctx ssh.Context, password string) bool {
	for _, a := range s.pwAuth {
		if a.Auth(ctx, password) {
			return true
		}
	}
	return false
}

func (s *Server) Start() error {
	var opts []ssh.Option

	opts = append(opts,
		ssh.PublicKeyAuth(s.authPublicKey),
		ssh.PasswordAuth(s.authPassword),
	)

	if s.hostkeyFile != "" {
		opts = append(opts, ssh.HostKeyFile(s.hostkeyFile))
	}

	return ssh.ListenAndServe(s.addr, s.handleSSH, opts...)
}

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func (s *Server) startCommand(cmd *exec.Cmd, session ssh.Session) error {
	ptyReq, winCh, isPty := session.Pty()
	if !isPty {
		return errors.Errorf("no pty requested")
	}

	rec, err := s.getRecorder(session)
	if err != nil {
		return errors.Wrap(err, "failed to create recorder")
	}

	cmd.Env = append(os.Environ(), fmt.Sprintf("TERM=%s", ptyReq.Term))
	f, err := pty.Start(cmd)
	if err != nil {
		return errors.Wrap(err, "start pty")
	}

	go func() {
		for win := range winCh {
			setWinsize(f, win.Width, win.Height)
		}
	}()

	stdin := pipe.New(rec.WriteInput)
	go io.Copy(f, stdin.Reader())
	go io.Copy(stdin.Writer(), session)

	stdout := pipe.New(rec.WriteOutput)
	go io.Copy(session, stdout.Reader())
	go io.Copy(stdout.Writer(), f)

	return cmd.Wait()
}

func (s *Server) handleSession(session ssh.Session) error {
	user, err := s.userStore.Get(session.User())
	if err != nil {
		return errors.Wrap(err, "find user")
	}

	shell := user.Shell
	logrus.WithFields(logrus.Fields{
		"user":       session.User(),
		"shell":      user.Shell,
		"session_id": session.Context().(ssh.Context).SessionID(),
	}).Infoln("Session started")

	cmd := exec.Command(shell)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: user.UID,
			Gid: user.GID,
		},
	}
	cmd.Args = []string{
		shell,
	}

	return errors.Wrap(s.startCommand(cmd, session), "running command")
}

func (s *Server) handleSSH(session ssh.Session) {
	l := logrus.WithFields(logrus.Fields{
		"user":       session.User(),
		"session_id": session.Context().(ssh.Context).SessionID(),
	})
	if err := s.handleSession(session); err != nil {
		io.WriteString(session, err.Error()+"\n")
		l.WithError(err).Errorln("Session ended")
		session.Exit(1)
		return
	}

	l.Infoln("Session ended")
	session.Exit(0)
}
