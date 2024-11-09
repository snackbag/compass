package main

import (
	"compass/compass"
	"fmt"
)

func main() {
	server := compass.NewServer()

	server.AddRoute("/", func(request compass.Request) string {
		return fmt.Sprintf("Hey, your IP is %s and you sent a %s request", request.IP, request.Method)
	})

	server.Start()
}
