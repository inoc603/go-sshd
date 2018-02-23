package asciicast

import (
	"encoding/json"
	"io"
	"time"
)

// Recorder records terminal session with the v2 asciicast file format
type Recorder struct {
	elapsedTime time.Duration
	lastWrite   time.Time
	encoder     *json.Encoder
	output      io.Writer
}

func NewRecorder(output io.Writer) *Recorder {
	return &Recorder{
		output:  output,
		encoder: json.NewEncoder(output),
	}
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
	Theme         struct {
		Fg      string `json:"fg,omitempty"`
		Bg      string `json:"bg,omitempty"`
		Palette string `json:"palette,omitempty"`
	} `json:"theme,omitempty"`
}

func (r *Recorder) WriteHeader(s Header) error {
	return r.encoder.Encode(s)
}

func (r *Recorder) Log(src string, content []byte) error {
	if r.lastWrite.IsZero() {
		r.lastWrite = time.Now()
	} else {
		delta := time.Now().Sub(r.lastWrite)
		r.lastWrite = r.lastWrite.Add(delta)
		r.elapsedTime += delta
	}

	return r.encoder.Encode([]interface{}{
		r.elapsedTime.Seconds(),
		src,
		string(content),
	})
}
