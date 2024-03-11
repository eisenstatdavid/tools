package scanner

import (
	"bufio"
	"fmt"
	"io"
)

type Scanner struct {
	br           *bufio.Reader
	rescan, done bool
	line         uint64
	text         string
	err          error
}

func Make(r io.Reader) Scanner {
	return Scanner{br: bufio.NewReader(r)}
}

func (s *Scanner) Err() error {
	if s.err != nil {
		return scannerError{s.line, s.err}
	}
	return nil
}

func (s *Scanner) Line() uint64 {
	return s.line
}

func (s *Scanner) Scan() bool {
	if s.rescan {
		s.rescan = false
		s.line++
		return true
	}
	if s.done {
		return false
	}
	s.line++
	switch text, err := s.br.ReadString('\n'); err {
	case io.EOF:
		s.done = true
		if text == "" {
			return false
		}
		fallthrough
	case nil:
		s.text = text
		return true
	default:
		s.SetErr(err)
		return false
	}
}

func (s *Scanner) SetErr(err error) {
	s.done = true
	s.err = err
}

func (s *Scanner) Text() string {
	return s.text
}

func (s *Scanner) Unscan() {
	s.rescan = true
	s.line--
}

type scannerError struct {
	line uint64
	err  error
}

func (e scannerError) Error() string {
	return fmt.Sprintf("line %d: %v", e.line, e.err)
}
