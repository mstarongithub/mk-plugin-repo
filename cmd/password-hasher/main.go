package main

import (
	"fmt"
	"os"

	"github.com/ermites-io/passwd"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Bad arguments. Usage: $<binary name> <password to hash>")
	}
	hasher, _ := passwd.New(passwd.Argon2idDefault)
	hash, _ := hasher.Hash([]byte(os.Args[1]))
	fmt.Print(string(hash))
}
