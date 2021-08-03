package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/eisenstatdavid/tools/internal/diff"
	"github.com/eisenstatdavid/tools/internal/rewrite"
)

func main() {
	diffs, err := diff.Parse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	fail := false
	for _, d := range diffs {
		if d.DstPath == os.DevNull || strings.HasSuffix(d.DstPath, ".go") {
			continue
		}
		if err := rewrite.File(d.DstPath, func(r io.Reader, w io.Writer) error {
			return rewriteChangedLines(d.DstChanges, r, w)
		}); err != nil {
			log.Print(err)
			fail = true
		}
	}
	if fail {
		os.Exit(1)
	}
}

const stringExpr = `"(?:[^\n"\\]|\\[^\n])*"`

var (
	stringRegexp  = regexp.MustCompile(stringExpr)
	stringsRegexp = regexp.MustCompile(stringExpr + `(?:\s*` + stringExpr + `)*`)
)

func rewriteChangedLines(changes []diff.Interval, r io.Reader, w io.Writer) error {
	content, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(w)
	line := uint64(1)
	i := 0
	for {
		loc := stringsRegexp.FindIndex(content)
		if loc == nil {
			break
		}
		prefix := content[:loc[0]]
		match := content[loc[0]:loc[1]]
		content = content[loc[1]:]
		for _, b := range prefix {
			if b == '\n' {
				line++
			}
		}
		for i < len(changes) && changes[i].Stop <= line {
			i++
		}
		for _, b := range match {
			if b == '\n' {
				line++
			}
		}
		if _, err := bw.Write(prefix); err != nil {
			return err
		}
		if i < len(changes) && line >= changes[i].Start {
			match = squash(match)
		}
		if _, err := bw.Write(match); err != nil {
			return err
		}
	}
	if _, err := bw.Write(content); err != nil {
		return err
	}
	return bw.Flush()
}

func squash(match []byte) []byte {
	result := append(make([]byte, 0, len(match)), '"')
	for _, s := range stringRegexp.FindAll(match, -1) {
		result = append(result, s[1:len(s)-1]...)
	}
	return append(result, '"')
}
