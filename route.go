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

type Route struct {
	parts     []routePart
	partIdMap map[string]int
	handler   func(request Request) Response

	repr string
}

func (r *Route) ToString() string {
	return r.repr
}

func (r *Route) matchesRawParts(split []string) bool {
	for i, str := range split {
		part := r.parts[i]
		if !strings.HasPrefix(str, part.prefix) {
			return false
		}

		if !strings.HasSuffix(str, part.suffix) {
			return false
		}

		minLen := len(part.prefix) + len(part.suffix)
		maxLen := minLen

		if part.id != "" {
			maxLen = 2048
		}

		if len(str) < minLen || len(str) > maxLen {
			return false
		}
	}

	return true
}

func (s *Server) AddRoute(path string, handler func(request Request) Response) {
	parts := createParts(path)
	length := len(parts)

	if length < 1 {
		s.Logger.Error(fmt.Sprintf("Skipped adding route %q, because it seems to be empty", path))
		return
	}

	if _, ok := s.routes[length]; !ok {
		s.routes[length] = make([]*Route, 0)
	}

	partIdMap := make(map[string]int)
	for i, part := range parts {
		partIdMap[part.id] = i
	}

	route := &Route{
		parts:     parts,
		partIdMap: partIdMap,
		handler:   handler,

		repr: path,
	}

	s.routes[length] = append(s.routes[length], route)
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

		id := strings.ToLower(match[1]) // name inside <>
		id = id[1 : len(id)-1]
		rawId := match[0] // full <...>
		idx := strings.Index(raw, rawId)

		prefix := raw[:idx]
		suffix := raw[idx+len(rawId):]

		parts = append(parts, routePart{id: id, prefix: prefix, suffix: suffix})
	}

	if len(parts) == 0 {
		parts = append(parts, routePart{id: "", prefix: "", suffix: ""})
	}

	return parts
}

func (s *Server) FindRoute(path string) *Route {
	split := strings.Split(path, "/")
	if strings.HasPrefix(path, "/") && len(split) > 1 {
		split = split[1:]
	}

	candidates, ok := s.routes[len(split)]
	if !ok {
		return nil
	}

	for _, candidate := range candidates {
		if !candidate.matchesRawParts(split) {
			continue
		}

		return candidate
	}

	return nil
}
