package main

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
)

func normalizeNumber(data []byte) []byte {
	s := string(data)
	if i, err := strconv.ParseUint(s, 10, 64); err == nil {
		return []byte(strconv.FormatUint(i, 10))
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatal(err)
	}
	return []byte(strconv.FormatFloat(f, 'f', -1, 64))
}

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	data = regexp.MustCompile(`\d+(?:\.\d+)?|\.\d+`).
		ReplaceAllFunc(data, normalizeNumber)
	if _, err := os.Stdout.Write(data); err != nil {
		log.Fatal(err)
	}
}
