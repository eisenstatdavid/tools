package diff

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/eisenstatdavid/tools/internal/scanner"
)

type Interval struct {
	Start, Stop uint64
}

func (i Interval) empty() bool {
	return i.Start >= i.Stop
}

type Diff struct {
	DstPath    string
	DstChanges []Interval
}

func Parse(r io.Reader) ([]Diff, error) {
	p := parser{}
	s := scanner.Make(r)
	for s.Scan() {
		if err := p.feedLine(s.Text()); err != nil {
			s.SetErr(err)
		}
	}
	if s.Err() != nil {
		return nil, s.Err()
	}
	return p.close()
}

type parser struct {
	diffs                               []Diff
	readyForDstPath, readyForHunkHeader bool
	hunkHeader
}

func (p *parser) feedLine(s string) error {
	if p.inHunk() {
		if err := p.feedInHunk(s); err != nil {
			return fmt.Errorf("in hunk: %v", err)
		}
		return nil
	}
	return p.feedNotInHunk(s)
}

func (p *parser) inHunk() bool {
	return !p.src.empty() || !p.dst.empty()
}

func (p *parser) feedInHunk(s string) error {
	switch s[0] {
	case '-':
		if p.src.empty() {
			return errors.New("unexpected src text")
		}
		p.changeDst(0)
		p.src.Start++
	case '+':
		if p.dst.empty() {
			return errors.New("unexpected dst text")
		}
		p.changeDst(1)
		p.dst.Start++
	case ' ':
		if p.src.empty() || p.dst.empty() {
			return errors.New("unexpected context")
		}
		p.src.Start++
		p.dst.Start++
	case '\\':
	default:
		return errors.New("unexpected comment")
	}
	return nil
}

func (p *parser) changeDst(count uint64) {
	d := &p.diffs[len(p.diffs)-1]
	d.DstChanges = appendInterval(d.DstChanges, Interval{p.dst.Start, p.dst.Start + count})
}

func appendInterval(intervals []Interval, i Interval) []Interval {
	if n := len(intervals); n > 0 && intervals[n-1].Stop == i.Start {
		intervals[n-1].Stop = i.Stop
		return intervals
	}
	return append(intervals, i)
}

func (p *parser) feedNotInHunk(s string) error {
	switch s[0] {
	case '-':
		if _, err := parsePath(s); err != nil {
			return err
		}
		if p.readyForDstPath {
			return errors.New("unexpected src path")
		}
		p.readyForHunkHeader = false
		p.diffs = append(p.diffs, Diff{})
		p.readyForDstPath = true
	case '+':
		path, err := parsePath(s)
		if err != nil {
			return err
		}
		if !p.readyForDstPath {
			return errors.New("unexpected dst path")
		}
		p.readyForDstPath = false
		p.diffs[len(p.diffs)-1].DstPath = path
		p.hunkHeader = hunkHeader{}
		p.readyForHunkHeader = true
	case '@':
		h, err := parseHunkHeader(s)
		if err != nil {
			return err
		}
		if !p.readyForHunkHeader {
			return errors.New("unexpected hunk header")
		}
		if h.src.Start <= p.src.Stop {
			return errors.New("src interval not strictly after previous interval")
		}
		if h.dst.Start <= p.dst.Stop {
			return errors.New("dst interval not strictly after previous interval")
		}
		if h.src.Start-p.src.Stop != h.dst.Start-p.dst.Stop {
			return errors.New("unequal skip lengths")
		}
		p.hunkHeader = h
	case ' ':
		return errors.New("unexpected context")
	case '\\':
	default:
		if p.readyForDstPath {
			return errors.New("unexpected comment")
		}
		p.readyForHunkHeader = false
	}
	return nil
}

var pathRegexp = regexp.MustCompile(`^(?:---|\+\+\+) ([^\t\n"\\]+)[\t\n]`)

func parsePath(s string) (string, error) {
	m := pathRegexp.FindStringSubmatch(s)
	if m == nil {
		return "", errors.New("invalid diff header")
	}
	return m[1], nil
}

type hunkHeader struct {
	src, dst Interval
}

var hunkHeaderRegexp = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

func parseHunkHeader(s string) (hunkHeader, error) {
	m := hunkHeaderRegexp.FindStringSubmatch(s)
	if m == nil {
		return hunkHeader{}, errors.New("invalid hunk header")
	}
	src, err := parseInterval(m[1], m[2])
	if err != nil {
		return hunkHeader{}, fmt.Errorf("src interval: %v", err)
	}
	dst, err := parseInterval(m[3], m[4])
	if err != nil {
		return hunkHeader{}, fmt.Errorf("dst interval: %v", err)
	}
	if src.empty() && dst.empty() {
		return hunkHeader{}, errors.New("empty hunk")
	}
	return hunkHeader{src, dst}, nil
}

func parseInterval(s, c string) (Interval, error) {
	start, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return Interval{}, fmt.Errorf("start: %v", err)
	}
	count, err := parseCount(c)
	if err != nil {
		return Interval{}, fmt.Errorf("count: %v", err)
	}
	if count == 0 {
		start++
	}
	if start == 0 {
		return Interval{}, errors.New("start out of range")
	}
	stop := start + count
	if stop < start {
		return Interval{}, errors.New("start + count out of range")
	}
	return Interval{start, stop}, nil
}

func parseCount(c string) (uint64, error) {
	if c == "" {
		return 1, nil
	}
	return strconv.ParseUint(c, 10, 64)
}

func (p *parser) close() ([]Diff, error) {
	if p.inHunk() || p.readyForDstPath {
		return nil, errors.New("unexpected end of input")
	}
	return p.diffs, nil
}
