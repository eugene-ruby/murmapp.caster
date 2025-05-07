package main

import (
	"log"

	"murmappcaster/internal"
)

func main() {
	if err := internal.Run(); err != nil {
		log.Fatalf("❌ fatal error: %v", err)
	}
}
