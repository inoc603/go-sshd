package auth

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
)

type LocalPublickKeyAuth struct {
	sync.RWMutex
	keys map[string]ssh.PublicKey
}

func NewLocalPublicKeyAuth(file string) (*LocalPublickKeyAuth, error) {
	s := &LocalPublickKeyAuth{
		keys: make(map[string]ssh.PublicKey),
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "open authorized key file")
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		pk, _, _, _, err := ssh.ParseAuthorizedKey(scanner.Bytes())
		if err != nil {
			return nil, errors.Wrap(err, "parse key")
		}
		s.keys[string(pk.Marshal())] = pk
	}

	return s, nil
}

func (s *LocalPublickKeyAuth) Auth(ctx ssh.Context, key ssh.PublicKey) bool {
	s.RLock()
	defer s.RUnlock()

	k, ok := s.keys[string(key.Marshal())]
	if !ok {
		return false
	}

	return ssh.KeysEqual(k, key)
}

type LocalUserStore struct {
}

func mustUint32(s string) uint32 {
	i, _ := strconv.ParseUint(s, 10, 32)
	return uint32(i)
}

func (s *LocalUserStore) Get(name string) (*User, error) {
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, errors.Wrap(err, "Faile to open /etc/passwd")
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if parts[0] != name {
			continue
		}
		return &User{
			Name:  name,
			UID:   mustUint32(parts[2]),
			GID:   mustUint32(parts[3]),
			Home:  parts[5],
			Shell: parts[6],
		}, nil
	}

	return nil, errors.Errorf("user %s not found", name)
}
