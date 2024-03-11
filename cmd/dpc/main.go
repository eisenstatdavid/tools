package main

import (
	"encoding/ascii85"
	"log"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	salt := []byte(os.Args[1])
	password, err := terminal.ReadPassword(syscall.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	cmd := exec.Command("pbcopy")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	w := ascii85.NewEncoder(stdin)
	password = append(password, '\n')
	hash := argon2.Key(password, salt, 16, 1<<16, 4, 16)
	_, err = w.Write(hash)
	if err != nil {
		log.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = stdin.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
