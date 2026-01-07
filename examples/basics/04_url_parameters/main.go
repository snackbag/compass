package main

import (
	"fmt"
	"github.com/snackbag/compass"
)

func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())

	server.AddRoute("/@<username>", func(request compass.Request) compass.Response {
		param := request.GetRouteParam("username") // returns empty if not found
		return compass.Text(param)
	})

	err := server.Run()
	if err != nil {
		server.Logger.Error(fmt.Sprintf("Failed to start: %s", err))
	}
}
