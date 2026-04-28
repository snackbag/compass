package main

import (
	"github.com/snackbag/compass/v2"
)

func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())

	server.AddRoute("/@<username>", func(request compass.Request) compass.Response {
		param, _ := request.GetRouteParam("username") // returns empty if not found
		return compass.Text(param)
	})

	server.MustRun()
}
