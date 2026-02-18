package main

import (
	"log"
	"os"

	"github.com/htrandev/metrics/pkg/crypto"
)

func main() {
	if err := crypto.Generate(""); err != nil {
		log.Printf("genarating keys: %s", err.Error())
		os.Exit(1)
	}
}