package numbers

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

const ePattern = `(\d+(?:\.\d+)?)[Ee]([+-]?\d+)`
const fPattern = `\d+\.\d+`
const uPattern = `\d+`

var eRegexp = regexp.MustCompile(`^` + ePattern + `$`)
var fRegexp = regexp.MustCompile(`^` + fPattern + `$`)
var uRegexp = regexp.MustCompile(`^` + uPattern + `$`)
var numberRegexp = regexp.MustCompile(strings.Join([]string{ePattern, fPattern, uPattern}, "|"))

func formatF(f float64) string {
	var s = strconv.FormatFloat(f, 'f', -1, 64)
	if strings.ContainsRune(s, '.') {
		return s
	}
	return strconv.FormatFloat(f, 'f', 1, 64)
}

func normalizeMatch(s string) string {
	if uRegexp.MatchString(s) {
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		return strconv.FormatUint(u, 10)
	}
	if fRegexp.MatchString(s) {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			log.Fatal(err)
		}
		return formatF(f)
	}
	m := eRegexp.FindStringSubmatch(s)
	sig, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		log.Fatal(err)
	}
	exp, err := strconv.ParseInt(m[2], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%se%d", formatF(sig), exp)
}

func Normalize(s string) string {
	return numberRegexp.ReplaceAllStringFunc(s, normalizeMatch)
}
