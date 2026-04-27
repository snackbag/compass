package main

import (
	"encoding/json"
	"fmt"
	"github.com/snackbag/compass"
	"os"
)

type Storage struct {
	Files map[string]File `json:"files"`
}

type File struct {
	Path       string `json:"path"`
	UploadDate string `json:"upload_date"`
}

var storage *Storage
var server *compass.Server

func main() {
	server = compass.NewServer(compass.NewStandardConfiguration())

	dat, err := os.ReadFile("storage.json")
	if err == nil {
		var st Storage
		err = json.Unmarshal(dat, &st)
		if err != nil {
			server.Logger.Error(fmt.Sprintf("Failed to unmarshal storage.json (read): %s", err))
		}
	} else {
		storage = &Storage{Files: make(map[string]File)}
		DumpStorage()
	}

	server.AddRoute("/", handleIndex)
	server.AddRoute("/upload", handleUpload).AllowedMethods = []string{"get", "post"}
	server.AddRoute("/download/<file>", handleDownload)

	server.MustRun()
}

func DumpStorage() {
	dat, err := json.Marshal(storage)
	if err != nil {
		server.Logger.Error(fmt.Sprintf("Failed to marshal storage.json (dump): %s", err))
	}

	os.WriteFile("storage.json", dat, 0755)
}
