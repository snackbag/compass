package main

import (
	"fmt"
	"github.com/snackbag/compass/v2"
)

// the example is kinda big. cookies are actually really simple
// to use, but I wanted to make this example a bit more complex to
// show what's possible

func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())

	server.AddRoute("/set/<message>", func(request compass.Request) compass.Response {
		_, ok := request.GetCookie("exampleCookie")
		if ok {
			return compass.Text("You already have a cookie. Eat it first")
		}

		message, ok := request.GetRouteParam("message")
		if !ok {
			return compass.Text("No message found")
		}

		resp := compass.Text("There you go!")
		resp.SetCookie(compass.Cookie{Name: "exampleCookie", Value: message})
		return resp
	})

	server.AddRoute("/get", func(request compass.Request) compass.Response {
		value, ok := request.GetCookie("exampleCookie")
		if !ok {
			return compass.Text(fmt.Sprintf("You don't have an example cookie yet. Here's everything we know: %s", request.GetCookies()))
		}

		return compass.Text(fmt.Sprintf("Look at that, you have a neat little cookie. Good job! It says %q", value))
	})

	server.AddRoute("/eat", func(request compass.Request) compass.Response {
		_, ok := request.GetCookie("exampleCookie")
		if !ok {
			return compass.Text("You need to have a cookie before you can eat it!")
		}

		resp := compass.Text("Mmmmmm that was so yummy")
		resp.RemoveCookie("exampleCookie")
		return resp
	})

	server.MustRun()
}
