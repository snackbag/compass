package compass

import (
	"os"
	"path/filepath"
)

func Fill(template string, server Server) Response {
	byteBody, err := os.ReadFile(filepath.Join(server.TemplatesDirectory, template))
	if err != nil {
		panic(err)
	}

	body := string(byteBody)
	return Text(body)
}
