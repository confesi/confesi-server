package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("missing arg")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--new-nonce":
		nonceGenerator()
	default:
		fmt.Println("invalid argument")
		fmt.Println("usage: [--new-nonce]")
		os.Exit(0)
	}
}

func nonceGenerator() {
	nonce := make([]byte, 12)

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		log.Fatal(err)
	}

	nonceStr := hex.EncodeToString(nonce)
	fmt.Println(nonceStr)
}
