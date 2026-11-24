package omnigraph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
)

type Stream[T any] struct {
	body    io.ReadCloser
	scanner *bufio.Scanner
	current T
	err     error
}

func newStream[T any](body io.ReadCloser) *Stream[T] {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	return &Stream[T]{
		body:    body,
		scanner: scanner,
	}
}

func (s *Stream[T]) Next() bool {
	if s.err != nil {
		return false
	}
	for s.scanner.Scan() {
		line := bytes.TrimSpace(s.scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var item T
		if err := json.Unmarshal(line, &item); err != nil {
			s.err = err
			return false
		}
		s.current = item
		return true
	}
	if err := s.scanner.Err(); err != nil {
		s.err = err
	}
	return false
}

func (s *Stream[T]) Item() T {
	return s.current
}

func (s *Stream[T]) Err() error {
	return s.err
}

func (s *Stream[T]) Close() error {
	return s.body.Close()
}
