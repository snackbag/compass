package main

import (
	"compass/compass"
	"fmt"
)

func main() {
	server := compass.NewServer()

	server.AddRoute("/", func(request compass.Request) compass.Response {
		return compass.Text(fmt.Sprintf("Hey, your IP is %s and you sent a %s request", request.IP, request.Method))
	})

	server.AddRoute("/test", func(request compass.Request) compass.Response {
		return compass.Redirect("https://google.com/")
	})

	server.SetNotFoundHandler(func(request compass.Request) compass.Response {
		return compass.TextWithCode("woah, that's not found", 404)
	})

	server.AddRoute("/test2", func(request compass.Request) compass.Response {
		ctx := compass.NewTemplateContext(request)
		ctx.SetVariable("test", "I'm a test")

		return compass.Fill("example.html", ctx, server)
	})

	server.Start()
}
