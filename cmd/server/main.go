// Command server runs the fantasy waiver assistant.
package main

import (
	"log"

	"fantasy_waiver_assistant/internal/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
