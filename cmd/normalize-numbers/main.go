package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/eisenstatdavid/tools/internal/numbers"
)

func main() {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stdout.WriteString(numbers.Normalize(string(data))); err != nil {
		log.Fatal(err)
	}
}
