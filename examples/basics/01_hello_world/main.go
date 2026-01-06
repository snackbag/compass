package main

import (
	"fmt"
	"github.com/snackbag/compass"
)

func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())

	server.AddRoute("/", func(request compass.Request) compass.Response {
		return compass.Text("hi hello hey")
	})

	err := server.Run()
	if err != nil {
		server.Logger.Error(fmt.Sprintf("Failed to start: %v", err))
	}
}
