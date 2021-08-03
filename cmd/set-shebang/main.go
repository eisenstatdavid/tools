package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/eisenstatdavid/tools/internal/rewrite"
)

func rewriteNeeded(shebang []byte, filename string) (needed bool, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	data := make([]byte, 0, len(shebang))
	for len(data) < cap(data) {
		count, err := file.Read(data[len(data):cap(data)])
		data = data[:len(data)+count]
		if err != nil {
			if err != io.EOF {
				return false, err
			}
			break
		}
	}
	return !bytes.Equal(data, shebang), nil
}

const shebangPrefix = "#!"

func rewriteIfNeeded(shebang []byte, filename string) error {
	needed, err := rewriteNeeded(shebang, filename)
	if err != nil {
		return err
	}
	if !needed {
		return nil
	}
	return rewrite.File(filename, func(r io.Reader, w io.Writer) error {
		b := bufio.NewReader(r)
		data, err := b.Peek(len(shebangPrefix))
		if err != nil && err != io.EOF {
			return err
		}
		if string(data) == shebangPrefix {
			for {
				_, isPrefix, err := b.ReadLine()
				if err != nil {
					if err != io.EOF {
						return err
					}
					break
				}
				if !isPrefix {
					break
				}
			}
		}
		if _, err := w.Write(shebang); err != nil {
			return err
		}
		_, err = io.Copy(w, b)
		return err
	})
}

func main() {
	var i int
	for i = 1; i < len(os.Args); i++ {
		if os.Args[i] == "--" {
			break
		}
	}
	if i < 2 || i >= len(os.Args) {
		fmt.Fprintln(os.Stderr, "usage: set-shebang utility_name [argument ...] -- [file ...]")
		os.Exit(1)
	}
	shebang := []byte(shebangPrefix + strings.Join(os.Args[1:i], " ") + "\n")
	fail := false
	for _, filename := range os.Args[i+1:] {
		if err := rewriteIfNeeded(shebang, filename); err != nil {
			log.Print(err)
			fail = true
		}
	}
	if fail {
		os.Exit(1)
	}
}
