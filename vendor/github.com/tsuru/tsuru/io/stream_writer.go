// Copyright 2016 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/docker/docker/pkg/jsonmessage"
	"golang.org/x/crypto/ssh/terminal"
)

type streamWriter struct {
	w         io.Writer
	b         []byte
	formatter Formatter
}

type syncWriter struct {
	w  io.Writer
	mu sync.Mutex
}

func (w *syncWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.w.Write(b)
}

var ErrInvalidStreamChunk = errors.New("invalid stream chunk")

type Formatter interface {
	Format(out io.Writer, data []byte) error
}

func NewStreamWriter(w io.Writer, formatter Formatter) *streamWriter {
	if formatter == nil {
		formatter = &SimpleJsonMessageFormatter{}
	}
	return &streamWriter{w: &syncWriter{w: w}, formatter: formatter}
}

func (w *streamWriter) Remaining() []byte {
	return w.b
}

func (w *streamWriter) Write(b []byte) (int, error) {
	w.b = append(w.b, b...)
	writtenCount := len(b)
	for len(w.b) > 0 {
		parts := bytes.SplitAfterN(w.b, []byte("\n"), 2)
		err := w.formatter.Format(w.w, parts[0])
		if err != nil {
			if err == ErrInvalidStreamChunk {
				if len(parts) == 1 {
					return writtenCount, nil
				} else {
					err = fmt.Errorf("Unparseable chunk: %q", parts[0])
				}
			}
			return writtenCount, err
		}
		if len(parts) == 1 {
			w.b = []byte{}
		} else {
			w.b = parts[1]
		}
	}
	return writtenCount, nil
}

type SimpleJsonMessage struct {
	Message string
	Error   string `json:",omitempty"`
}

type SimpleJsonMessageFormatter struct {
	pipeReader io.Reader
	pipeWriter io.WriteCloser
}

func likeJSON(str string) bool {
	data := bytes.TrimSpace([]byte(str))
	return len(data) > 1 && data[0] == '{' && data[len(data)-1] == '}'
}

type withFd interface {
	Fd() uintptr
}

type withFD interface {
	FD() uintptr
}

func (f *SimpleJsonMessageFormatter) Format(out io.Writer, data []byte) error {
	if len(data) == 0 || (len(data) == 1 && data[0] == '\n') {
		return nil
	}
	var msg SimpleJsonMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return ErrInvalidStreamChunk
	}
	if msg.Error != "" {
		return errors.New(msg.Error)
	}
	if likeJSON(msg.Message) {
		if f.pipeWriter == nil {
			f.pipeReader, f.pipeWriter = io.Pipe()
			var fd uintptr
			switch v := out.(type) {
			case withFd:
				fd = v.Fd()
			case withFD:
				fd = v.FD()
			}
			isTerm := terminal.IsTerminal(int(fd))
			go jsonmessage.DisplayJSONMessagesStream(f.pipeReader, out, fd, isTerm, nil)
		}
		f.pipeWriter.Write([]byte(msg.Message))
	} else {
		if f.pipeWriter != nil {
			f.pipeWriter.Close()
			f.pipeWriter = nil
			f.pipeReader = nil
		}
		out.Write([]byte(msg.Message))
	}
	return nil
}

type SimpleJsonMessageEncoderWriter struct {
	*json.Encoder
}

func (w *SimpleJsonMessageEncoderWriter) Write(msg []byte) (int, error) {
	err := w.Encode(SimpleJsonMessage{Message: string(msg)})
	if err != nil {
		return 0, err
	}
	return len(msg), nil
}
