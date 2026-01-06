package compass

import (
	"fmt"
	"regexp"
	"strings"
)

var routeParamRegex = regexp.MustCompile("(<.+>)")

type routePart struct {
	id     string
	prefix string
	suffix string
}

func createParts(path string) []routePart {
	split := strings.Split(path, "/")
	parts := make([]routePart, 0)

	for _, raw := range split {
		if raw == "" {
			continue
		}

		match := routeParamRegex.FindStringSubmatch(raw)
		if match == nil {
			parts = append(parts, routePart{id: "", prefix: raw, suffix: ""})
			continue
		}

		id := match[1] // name inside <>
		id = id[1 : len(id)-1]
		rawId := match[0] // full <...>
		idx := strings.Index(raw, rawId)

		prefix := raw[:idx]
		suffix := raw[idx+len(rawId):]

		parts = append(parts, routePart{id: id, prefix: prefix, suffix: suffix})
	}

	return parts
}

func (s *Server) AddRoute(path string, handler func(request Request) Response) {
	fmt.Println(createParts(path))
}
