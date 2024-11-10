package compass

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FillParser struct {
	Contents string

	col  int
	line int
}

func (parser *FillParser) Convert() string {
	converted := ""

	// none = 0, var = 1, action = 2
	inPart := 0
	lastChars := ""

	for _, currentRune := range parser.Contents {
		char := string(currentRune)

		if inPart == 0 && lastChars == "" && char == "<" {
			lastChars += char
			continue
		}

		if inPart == 0 && char == "$" {
			if lastChars != "<" {
				converted += char
				continue
			}

			inPart = 1
			lastChars = ""
			continue
		} else if inPart == 0 && lastChars == "<" {
			converted += lastChars + char
			lastChars = ""
			continue
		}

		if inPart == 1 && strings.HasSuffix(lastChars, "/>") {
			lastChars = strings.TrimSuffix(lastChars, "/>")
			converted += fmt.Sprintf("{{ VARIABLE %s }}", lastChars)
			converted += char
			inPart = 0
			continue
		}

		if inPart == 1 {
			lastChars += char
			continue
		}

		converted += char
	}

	return converted
}

func Fill(template string, server Server) Response {
	byteBody, err := os.ReadFile(filepath.Join(server.TemplatesDirectory, template))
	if err != nil {
		panic(err)
	}

	body := string(byteBody)
	parser := FillParser{Contents: body, col: -1, line: 0}
	return Text(parser.Convert())
}
