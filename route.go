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

	AllowedMethods []string

	repr string
}

// ToString returns the original string representation of the route
// as it was defined when added to the server.
func (r *Route) ToString() string {
	return r.repr
}

// matchesRawParts checks whether the given split URL path matches
// this route's structure.
//
// Each segment is validated against its corresponding routePart by
// verifying prefix and suffix constraints. If the part defines a
// parameter (id != ""), the middle section may vary in length;
// otherwise, the segment must match exactly.
//
// Returns true if all parts match, false otherwise.
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

// AddRoute registers a new route on the server.
//
// The path is split into segments (by "/") and converted into an
// internal structure used for matching requests later. Segments can
// contain dynamic parameters using angle brackets, for example:
//
//	"/users/<id>"
//
// Each segment of the path is analyzed:
//   - Static segments (e.g. "users") must match exactly.
//   - Dynamic segments (e.g. "<id>") capture a value from the URL.
//   - Mixed segments (e.g. "file-<name>.txt") extract only the dynamic part,
//     while enforcing the surrounding prefix ("file-") and suffix (".txt").
//
// Routes are grouped by the number of segments, so only routes with
// the same structure length are compared during lookup.
//
// Parameter names are taken from inside "< >" and converted to lowercase.
//
// If the given path results in no usable segments, the route is ignored.
func (s *Server) AddRoute(path string, handler func(request Request) Response) *Route {
	parts := createParts(path)
	length := len(parts)

	if length < 1 {
		s.Logger.Error(fmt.Sprintf("Skipped adding route %q, because it seems to be empty", path))
		return nil
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

		repr:           path,
		AllowedMethods: []string{"get"},
	}

	s.routes[length] = append(s.routes[length], route)
	return route
}

// createParts breaks a route path into individual parts used for matching.
//
// Each segment of the path is analyzed:
//   - Static segments (e.g. "users") must match exactly.
//   - Dynamic segments (e.g. "<id>") capture a value from the URL.
//   - Mixed segments (e.g. "file-<name>.txt") extract only the dynamic part,
//     while enforcing the surrounding prefix ("file-") and suffix (".txt").
//
// Parameter names are taken from inside "< >" and converted to lowercase.
//
// If the path is empty, a single empty part is returned so the rest of
// the system can still operate consistently.
func createParts(path string) []routePart {
	split := splitUrlPath(path)
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

// FindRoute attempts to match a given path to a registered route.
//
// The path is split into segments and only routes with the same
// segment count are considered. Each candidate route is checked
// against the path using its matching rules.
//
// Returns the first matching route, or nil if no match is found.
func (s *Server) FindRoute(path string) *Route {
	split := splitUrlPath(path)

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
