package main

import "github.com/snackbag/compass"

func handleDownload(request compass.Request) compass.Response {
	return compass.Text("Download")
}
