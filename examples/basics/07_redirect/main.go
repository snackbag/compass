package main

import "github.com/snackbag/compass"

func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())

	server.AddRoute("/target", func(request compass.Request) compass.Response {
		return compass.Text("Woohooo! You've been redirected :D")
	})

	server.AddRoute("/", func(request compass.Request) compass.Response {
		return compass.Redirect("/target", false)
	})

	server.MustRun()
}
