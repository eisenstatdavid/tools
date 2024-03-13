package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"unicode"
)

func main() {
	rd := bufio.NewReader(os.Stdin)
	w := bufio.NewWriter(os.Stdout)
	defer func() {
		if err := w.Flush(); err != nil {
			log.Fatal(err)
		}
	}()
	state := 0
	for {
		r, size, err := rd.ReadRune()
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		if size <= 0 {
			break
		}
		if !unicode.IsSpace(r) {
			switch state {
			case 0, 1:
			case 2:
				if _, err := w.WriteRune(' '); err != nil {
					log.Fatal(err)
				}
			case 3:
				if _, err := w.WriteRune('\n'); err != nil {
					log.Fatal(err)
				}
			case 4:
				if _, err := w.WriteString("\n\n"); err != nil {
					log.Fatal(err)
				}
			case 5:
				if _, err := w.WriteString("\n\t"); err != nil {
					log.Fatal(err)
				}
			}
			state = 1
			if _, err := w.WriteRune(r); err != nil {
				log.Fatal(err)
			}
		} else if r == '\n' {
			switch state {
			case 0:
			case 1:
				state = 3
			case 2:
				state = 3
			case 3:
				state = 4
			case 4:
			case 5:
				state = 3
			}
		} else {
			switch state {
			case 0:
			case 1:
				state = 2
			case 2:
			case 3:
				state = 5
			case 4:
			case 5:
				state = 3
			}
		}
	}
	if state > 0 {
		if _, err := w.WriteRune('\n'); err != nil {
			log.Fatal(err)
		}
	}
}
