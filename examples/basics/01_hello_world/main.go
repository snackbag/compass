package main

import (
	"github.com/snackbag/compass/v2"
)

func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())

	server.AddRoute("/", func(request compass.Request) compass.Response {
		return compass.Text("hi hello hey")
	})

	server.MustRun()
}
