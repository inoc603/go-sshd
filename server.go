package sshd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"github.com/Sirupsen/logrus"
	"github.com/gliderlabs/ssh"
	"github.com/inoc603/go-sshd/asciicast"
	"github.com/inoc603/go-sshd/auth"
	"github.com/inoc603/go-sshd/pipe"
	"github.com/inoc603/go-sshd/storage"
	"github.com/kr/pty"
	"github.com/pkg/errors"
)

type Option func(s *Server) error

type Server struct {
	addr        string
	hostkeyFile string
	outputStore storage.Storage
	userStore   auth.UserStore
	pkAuth      []auth.PublicKeyAuth
	pwAuth      []auth.PasswordAuth
}

func NewServer(opts ...Option) (*Server, error) {
	s := &Server{
		addr:      ":22",
		userStore: &auth.DummyUserStore{},
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *Server) Start() error {
	var opts []ssh.Option

	for _, a := range s.pkAuth {
		opts = append(opts, ssh.PublicKeyAuth(a.Auth))
	}

	for _, a := range s.pwAuth {
		opts = append(opts, ssh.PasswordAuth(a.Auth))
	}

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

	ctx, ok := session.Context().(ssh.Context)
	if !ok {
		panic("SHIT")
	}

	output, err := s.outputStore.New(ctx)
	if err != nil {
		return errors.Wrap(err, "open case file")
	}
	defer output.Close()

	rec := asciicast.NewRecorder(output)

	rec.WriteHeader(asciicast.Header{
		Env: map[string]string{
			"TERM": ptyReq.Term,
		},
		Width:     ptyReq.Window.Width,
		Height:    ptyReq.Window.Height,
		Version:   2,
		Timestamp: time.Now().Unix(),
	})

	go io.Copy(f, session)

	stdout := pipe.New(func(b []byte) {
		rec.Log("o", b)
	})
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

func WithStorage(store storage.Storage) Option {
	return func(s *Server) error {
		s.outputStore = store
		return nil
	}
}
