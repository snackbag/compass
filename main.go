package main

import (
	"fmt"
	"github.com/snackbag/compass/compass"
)

func main() {
	server := compass.NewServer()
	server.SetSessionSecret("dev")

	server.AddRoute("/", func(request compass.Request) compass.Response {
		session := request.GetSession()
		if session == nil {
			return compass.Text("No session")
		}

		session.WriteString("bruh", "grahhh")
		session.WriteBool("imabool", true)
		session.Commit()

		resp := compass.Text(fmt.Sprintf("Hey, your IP is %s and you sent a %s request", request.IP, request.Method))
		return resp
	})

	server.AddRoute("/set", func(request compass.Request) compass.Response {
		session := request.GetSession()
		if session != nil {
			return compass.Text("You already have a set session")
		}

		resp := compass.Text("Session set!")
		resp.SetSession(compass.NewSession(&server))

		return resp
	})

	server.AddRoute("/clear", func(request compass.Request) compass.Response {
		session := request.GetSession()
		if session == nil {
			return compass.Text("You do not have any session saved")
		}

		resp := compass.Text("Cleared session")
		resp.ClearSession()

		return resp
	})

	server.AddRoute("/test", func(request compass.Request) compass.Response {
		return compass.Redirect("https://google.com/")
	})

	server.SetNotFoundHandler(func(request compass.Request) compass.Response {
		return compass.TextWithCode("woah, that's not found", 404)
	})

	server.AddRoute("/test2", func(request compass.Request) compass.Response {
		ctx := compass.NewTemplateContext(request)
		ctx.SetVariable("test", false)

		return compass.Fill("example.html", ctx, server)
	})

	server.AddRoute("/route-test/<part1>/@<part2>/<part3>wow", func(request compass.Request) compass.Response {
		part1 := request.GetParam("part1")
		part2 := request.GetParam("part2")
		part3 := request.GetParam("part3")

		return compass.Text(fmt.Sprintf("P1 %s (@%s) P3: %s", part1, part2, part3))
	})

	server.Start()
}
