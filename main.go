package main

import (
	"compass/compass"
)

func main() {
	server := compass.NewServer()

	server.AddRoute("/", func(request compass.Request) string {
		return "<h1>test</h1>"
	})

	server.Start()
}
