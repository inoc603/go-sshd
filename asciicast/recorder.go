package asciicast

import (
	"encoding/json"
	"io"
	"time"
)

// Recorder records terminal session with the v2 asciicast file format
type Recorder struct {
	recordStdin   bool
	raw           bool
	idleTimeLimit time.Duration
	elapsedTime   time.Duration
	lastWrite     time.Time
	encoder       *json.Encoder
	output        io.Writer
}

type Option func(r *Recorder) error

func IdleTimeLimit(d time.Duration) Option {
	return func(r *Recorder) error {
		r.idleTimeLimit = d
		return nil
	}
}

func Stdin(b bool) Option {
	return func(r *Recorder) error {
		r.recordStdin = b
		return nil
	}
}

func Raw(b bool) Option {
	return func(r *Recorder) error {
		r.raw = b
		return nil
	}
}

func NewRecorder(output io.Writer, opts ...Option) *Recorder {
	return &Recorder{
		output:  output,
		encoder: json.NewEncoder(output),
	}
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

func (r *Recorder) log(src string, content []byte) error {
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

func (r *Recorder) WriteInput(b []byte) { r.log("i", b) }

func (r *Recorder) WriteOutput(b []byte) { r.log("o", b) }
