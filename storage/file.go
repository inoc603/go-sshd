package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/errors"
)

func NewFileStorage(dir string) (*StorageFile, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, errors.Wrap(err, "create output directory")
	}
	return &StorageFile{dir}, nil
}

type StorageFile struct {
	dir string
}

func (s *StorageFile) New(ctx ssh.Context) (io.WriteCloser, error) {
	fname := fmt.Sprintf("%d_%s_%s.jsonl",
		time.Now().Unix(),
		ctx.User(),
		ctx.RemoteAddr().String(),
	)

	return os.OpenFile(filepath.Join(s.dir, fname), os.O_CREATE|os.O_WRONLY, 0660)
}
