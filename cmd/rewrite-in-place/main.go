package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/eisenstatdavid/tools/internal/rewrite"
)

func inParallel(parallelism int, f func(string) error, args []string) int {
	fails := uintptr(0)
	var wg sync.WaitGroup
	wg.Add(len(args))
	s := semaphore.NewWeighted(int64(parallelism))
	for _, arg := range args {
		go func(arg string) {
			defer wg.Add(-1)
			if err := s.Acquire(context.Background(), 1); err != nil {
				log.Print(err)
				return
			}
			defer s.Release(1)
			if err := f(arg); err != nil {
				log.Print(err)
				atomic.AddUintptr(&fails, 1)
			}
		}(arg)
	}
	wg.Wait()
	return int(fails)
}

func main() {
	var i int
	for i = 1; i < len(os.Args); i++ {
		if os.Args[i] == "--" {
			break
		}
	}
	if i < 2 || i >= len(os.Args) {
		fmt.Fprintln(os.Stderr, "usage: rewrite-in-place utility_name [argument ...] -- [file ...]")
		os.Exit(1)
	}
	fails := inParallel(runtime.NumCPU(), func(filename string) error {
		return rewrite.File(filename, func(r io.Reader, w io.Writer) error {
			cmd := exec.Command(os.Args[1], os.Args[2:i]...)
			cmd.Stdin = r
			cmd.Stdout = w
			cmd.Stderr = os.Stderr
			return cmd.Run()
		})
	}, os.Args[i+1:])
	const MAXFAILS = 125
	if fails > MAXFAILS {
		fails = MAXFAILS
	}
	os.Exit(fails)
}
