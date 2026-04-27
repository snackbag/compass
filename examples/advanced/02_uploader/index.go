package main

import "github.com/snackbag/compass"

func handleIndex(request compass.Request) compass.Response {
	return compass.Text("Index")
}
