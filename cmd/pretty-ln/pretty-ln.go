package main

import (
	"flag"
	"os"
	"path/filepath"
)

func main() {
	flag.Parse()
	for _, name := range flag.Args() {
		ln, err := os.Readlink(name)
		if err != nil {
			panic(err)
		}
		ln = filepath.Clean(ln)
		err = os.Symlink(filepath.Join(filepath.Dir(ln), filepath.Base(name)), name+"~")
		if err != nil {
			panic(err)
		}
		err = os.Rename(name+"~", name)
		if err != nil {
			panic(err)
		}
	}
}
