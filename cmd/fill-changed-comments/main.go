package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/eisenstatdavid/tools/internal/diff"
	"github.com/eisenstatdavid/tools/internal/numbers"
	"github.com/eisenstatdavid/tools/internal/rewrite"
	"github.com/eisenstatdavid/tools/internal/scanner"
)

const maxCol = 80

func main() {
	diffs, err := diff.Parse(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	fail := false
	for _, d := range diffs {
		if d.DstPath == os.DevNull {
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

var (
	lineCommentRegexp = regexp.MustCompile(`^([\t ]*(?:#|//))([\t ][^\n]*)`)
	nextCommentRegexp = regexp.MustCompile(`^[\t ]*//[\t ]*[Nn]ext(?:(?:[\t ]+available)?(?:[\t ]+)(?:id|tag))?:[\t ]+\d+[\t ]*\n$`)
	listRegexp        = regexp.MustCompile(`^[\t ]*// (?:[-*]|[0-9]\.) `)
)

func isFillableLineComment(t string) bool {
	return lineCommentRegexp.MatchString(t) && !nextCommentRegexp.MatchString(t) && !listRegexp.MatchString(t)
}

func rewriteChangedLines(changes []diff.Interval, r io.Reader, w io.Writer) error {
	bw := bufio.NewWriter(w)
	s := scanner.Make(r)
	i := 0
	for s.Scan() {
		for i < len(changes) && changes[i].Stop <= s.Line() {
			i++
		}
		lines := []string{s.Text()}
		rewrite := rewriteCode
		t := s.Text()
		if isFillableLineComment(t) {
			rewrite = rewriteComment
			for s.Scan() {
				t := s.Text()
				if !isFillableLineComment(t) {
					s.Unscan()
					break
				}
				lines = append(lines, s.Text())
			}
		}
		if i < len(changes) && s.Line() >= changes[i].Start {
			for j, line := range lines {
				lines[j] = numbers.Normalize(line)
			}
			lines = rewrite(lines)
		}
		for _, line := range lines {
			if _, err := bw.WriteString(line); err != nil {
				return err
			}
		}
	}
	if s.Err() != nil {
		return s.Err()
	}
	return bw.Flush()
}

var tokenRegexp = regexp.MustCompile(strings.Join([]string{
	`(?:#|//)[^\n]*`,                        // line comment
	`/\*(?:[^*]|\*+[^*/])*(?:$|\*+(?:$|/))`, // general comment
	`'(?:[^\n'\\]|\\[^\n])*(?:$|\\$|')`,     // rune literal
	"`[^`]*(?:$|`)",                         // raw string literal
	`"(?:[^\n"\\]|\\[^\n])*(?:$|\\$|")`,     // interpreted string literal
}, "|"))

func rewriteCode(lines []string) []string {
	for i, line := range lines {
		lines[i] = tokenRegexp.ReplaceAllStringFunc(
			strings.TrimRightFunc(line, unicode.IsSpace),
			func(tok string) string {
				return strings.Join(strings.Fields(tok), " ")
			}) + "\n"
	}
	return lines
}

func rewriteComment(lines []string) []string {
	var (
		prefix string
		words  []string
	)
	for i, line := range lines {
		m := lineCommentRegexp.FindStringSubmatch(line)
		if i == 0 {
			prefix = m[1]
		}
		words = append(words, strings.Fields(m[2])...)
	}
	lines = nil
	for i := 0; i < len(words); {
		j := i
		for col := advanceString(0, prefix); j < len(words); j++ {
			col = advanceRune(col, ' ')
			col = advanceString(col, words[j])
			if col > maxCol {
				break
			}
		}
		if j == i {
			j++
		}
		var b bytes.Buffer
		_, _ = b.WriteString(prefix)
		for ; i < j; i++ {
			_, _ = b.WriteRune(' ')
			_, _ = b.WriteString(words[i])
		}
		_, _ = b.WriteRune('\n')
		lines = append(lines, b.String())
	}
	return lines
}

func advanceString(col uint64, s string) uint64 {
	for _, r := range s {
		col = advanceRune(col, r)
	}
	return col
}

func advanceRune(col uint64, r rune) uint64 {
	if r == '\t' {
		col |= 1
	}
	return col + 1
}
