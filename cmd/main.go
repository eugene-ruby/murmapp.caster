package main

import (
	"log"

	"github.com/eugene-ruby/murmapp.caster/internal"
)

func main() {
	if err := internal.Run(); err != nil {
		log.Fatalf("❌ fatal error: %v", err)
	}
}
