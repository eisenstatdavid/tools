package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func slurp() string {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

var rightSingleQuotationMark = regexp.MustCompile("\u2019")

func normalizeApostrophes(s string) string {
	return rightSingleQuotationMark.ReplaceAllString(s, "'")
}

var (
	upperUpperLower = regexp.MustCompile("([[:upper:]])([[:upper:]][[:lower:]])")
	lowerUpper      = regexp.MustCompile("([[:lower:]])([[:upper:]])")
)

func cleaveCamelCase(s string) string {
	s = upperUpperLower.ReplaceAllString(s, `$1 $2`)
	return lowerUpper.ReplaceAllString(s, `$1 $2`)
}

var word = regexp.MustCompile("[[:alpha:]](?:'?[[:alpha:]])*")

func splitWords(s string) []string {
	return word.FindAllString(s, -1)
}

func count(a []string) map[string]int {
	cnt := make(map[string]int)
	for _, w := range a {
		cnt[w] += 1
	}
	return cnt
}

func keys(m map[string]int) []string {
	var a []string
	for w, _ := range m {
		a = append(a, w)
	}
	return a
}

func runAspell(words []string, extraArgs ...string) []string {
	args := []string{"list", "--mode=none"}
	cmd := exec.Command("aspell", append(args, extraArgs...)...)
	var b bytes.Buffer
	for _, word := range words {
		_, _ = b.WriteString(word)
		_ = b.WriteByte('\n')
	}
	cmd.Stdin = &b
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return splitWords(string(out))
}

func set(a []string) map[string]bool {
	m := make(map[string]bool)
	for _, w := range a {
		m[w] = true
	}
	return m
}

var collator = collate.New(language.AmericanEnglish, collate.Loose)
var buffer collate.Buffer

func key(w string) string {
	return string(collator.KeyFromString(&buffer, w))
}

func collatedCount(a []string) map[string]int {
	cnt := make(map[string]int)
	for _, w := range a {
		cnt[key(w)] += 1
	}
	return cnt
}

func main() {
	s := slurp()
	s = normalizeApostrophes(s)
	s = cleaveCamelCase(s)
	a := splitWords(s)
	cnt := count(a)
	cCnt := collatedCount(a)
	iffy := runAspell(keys(cnt))
	sort.Strings(iffy)
	sort.SliceStable(iffy, func(i, j int) bool {
		return cnt[iffy[i]] > cnt[iffy[j]]
	})
	bad := set(runAspell(iffy, "--ignore-case", "--run-together"))
	for _, w := range iffy {
		c := cnt[w]
		cc := cCnt[key(w)]
		out := fmt.Sprintf("%7d %s", c, w)
		if cc > c {
			out = fmt.Sprintf("%s (%d)", out, cc)
		}
		if bad[w] {
			out = fmt.Sprintf("\x1b[33m%s\x1b[0m", out)
		}
		_, err := fmt.Println(out)
		if err != nil {
			log.Fatal(err)
		}
	}
}
