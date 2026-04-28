package main

import "github.com/snackbag/compass/v2"

func main() {
	server := compass.NewServer(compass.NewStandardConfiguration())

	server.AddRoute("/", func(request compass.Request) compass.Response {
		return compass.JsonString(`{"test": 123, "wow": ["a", "b", "c"]}`)
	})

	server.AddRoute("/object", func(request compass.Request) compass.Response {
		return compass.JsonMarshal(server.Config)
	})

	server.MustRun()
}
