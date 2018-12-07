package asciicast

import (
	"encoding/json"
	"io"
	"time"
)

// Session records terminal session with the v2 asciicast file format
type Session struct {
	recordStdin   bool
	raw           bool
	idleTimeLimit time.Duration
	elapsedTime   time.Duration
	lastWrite     time.Time
	encoder       *json.Encoder
	output        io.WriteCloser
}

type Option func(r *Session)

func IdleTimeLimit(d time.Duration) Option {
	return func(r *Session) { r.idleTimeLimit = d }
}

func Stdin(b bool) Option {
	return func(r *Session) { r.recordStdin = b }
}

func Raw(b bool) Option {
	return func(r *Session) { r.raw = b }
}

func NewRecorder(output io.WriteCloser, opts ...Option) *Session {
	s := &Session{
		output:  output,
		encoder: json.NewEncoder(output),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type Theme struct {
	Fg      string `json:"fg,omitempty"`
	Bg      string `json:"bg,omitempty"`
	Palette string `json:"palette,omitempty"`
}

type Header struct {
	Version       int               `json:"version"`
	Width         int               `json:"width"`
	Height        int               `json:"height"`
	Timestamp     int64             `json:"timestamp,omitempty"`
	Duration      float64           `json:"duration,omitempty"`
	IdleTimeLimit float64           `json:"idle_time_limit,omitempty"`
	Command       string            `json:"command,omitempty"`
	Title         string            `json:"title,omitempty"`
	Env           map[string]string `json:"env,omitempty"`
	Theme         Theme             `json:"theme,omitempty"`
}

func (r *Session) WriteHeader(h Header) error {
	return r.encoder.Encode(h)
}

func (r *Session) log(src string, content []byte) error {
	if r.lastWrite.IsZero() {
		r.lastWrite = time.Now()
	} else {
		delta := time.Now().Sub(r.lastWrite)
		r.lastWrite = r.lastWrite.Add(delta)
		if r.idleTimeLimit > 0 && delta > r.idleTimeLimit {
			delta = r.idleTimeLimit
		}
		r.elapsedTime += delta
	}

	if r.raw {
		_, err := r.output.Write(content)
		return err
	}

	return r.encoder.Encode([]interface{}{
		r.elapsedTime.Seconds(),
		src,
		string(content),
	})
}

func (r *Session) Write(p []byte) (n int, err error) {
	return len(p), r.log("o", p)
}

func (r *Session) Close() error {
	return r.output.Close()
}
