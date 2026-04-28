package main

import (
	"fmt"
	"github.com/snackbag/compass/v2"
	"os"
	"path/filepath"
	"strings"
)

func handleIndex(request compass.Request) compass.Response {
	data, err := os.ReadFile(filepath.Join("template", "index.html"))
	if err != nil {
		return compass.InternalError("failed to read index.html")
	}

	data = []byte(strings.ReplaceAll(string(data), "<Items/>", createTable()))

	return compass.ServeBytes(data, "index.html")
}

func createTable() string {
	entries, err := os.ReadDir("uploads")
	if err != nil {
		return ""
	}

	rv := ``

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join("uploads", entry.Name())
		res, err := os.Stat(path)
		if err != nil {
			server.Logger.Error(fmt.Sprintf("failed to stat upload %s", entry.Name()))
			continue
		}

		rv += "<tr>"
		rv += fmt.Sprintf(
			`<td><a href="/download/%s">%s</a></td><td>%s</td></tr>`,
			entry.Name(),
			entry.Name(),
			res.ModTime().Format("02-Jan-2006 15:04"),
		)
	}

	return rv
}
