# Routing

**File:** `routing.go`

## Overview

The routing system converts URL path strings into structured `Route` values that can be
efficiently matched against incoming request paths. It supports three kinds of path
segments:

- **Static** - must match exactly (`users`, `settings`)
- **Dynamic** - captures a variable value (`<id>`, `<name>`)
- **Mixed** - a static prefix and/or suffix around a dynamic capture (`file-<name>.txt`)

Routes are stored in a `map[int][]*Route` on the server, keyed by segment count. This
means a three-segment path like `/users/42/profile` only ever compares against routes
that also have three segments and not every route in the system.

## Data structures

### routePart

```go
type routePart struct {
  id     string
  prefix string
  suffix string
}
```

One `routePart` per path segment. For a static segment, `id` is empty and `prefix` holds
the literal string. For a dynamic segment, `id` holds the parameter name and `prefix`/`suffix`
hold the surrounding literal text (empty strings for a bare `<param>`).

Examples:

| Segment           | id       | prefix    | suffix   |
|-------------------|----------|-----------|----------|
| `users`           | `""`     | `"users"` | `""`     |
| `<id>`            | `"id"`   | `""`      | `""`     |
| `file-<name>.txt` | `"name"` | `"file-"` | `".txt"` |

### Route

```go
type Route struct {
  parts     []routePart
  partIdMap map[string]int
  handler   func (request Request) Response
  repr      string
}
```

`parts` is the parsed segment slice. `partIdMap` maps parameter names to their index in
`parts`, enabling O(1) lookup in `GetRouteParam`. `repr` stores the original path string
so `ToString()` can return it without reconstructing it.

## Registration: AddRoute

`AddRoute(path, handler)` parses the path into parts via `createParts`, then appends the
resulting `Route` to the appropriate slice in `s.routes[len(parts)]`.

If the parsed result has zero parts (e.g. an empty string was passed), the route is
skipped and an error is logged. This is a soft failure and the server continues running.

Parameter names inside `< >` are lowercased. `<ID>` and `<Id>` both become `"id"`.
This normalisation happens in `createParts` so `GetRouteParam` callers can always use
lowercase names without thinking about it.

## Parsing: createParts

`createParts` calls `splitUrlPath` to break the path into segments, then processes each:

1. If the segment contains no `< >`, it becomes a static part.
2. If it does, the regex `(<.+>)` extracts the parameter token. `strings.Index` locates
   it within the raw segment to derive the prefix and suffix.

The regex is compiled once at package level (`routeParamRegex`) to avoid recompiling on
every registration.

Only the first `< >` match per segment is used. A segment like `<a>-<b>` would only
capture `a`. This is a current limitation. If multi-param segments are needed in the future,
`createParts` is the place to extend.

## Matching: FindRoute

```go
func (s *Server) FindRoute(path string) *Route
```

1. `splitUrlPath(path)` breaks the incoming path into segments.
2. Only routes with `len(split)` segments are considered.
3. Each candidate is checked via `matchesRawParts`.
4. The first match wins.

`matchesRawParts` checks each segment against its corresponding `routePart`:

- The segment must have `part.prefix` as a prefix and `part.suffix` as a suffix.
- For static parts (`id == ""`), no additional characters are allowed between prefix and
  suffix. The minimum and maximum length are both `len(prefix) + len(suffix)`.
- For dynamic parts, the length cap is 2048 characters, giving the middle section room
  to vary.

The 2048 limit is a safety ceiling against unreasonably long URL parameters. It is not
configurable today. If a use case requires longer parameters, that constant is the thing
to change or expose.

## Parameter extraction: GetRouteParam

`GetRouteParam` is defined on `Request` (in `request.go`) but depends entirely on route
internals, so it is documented here.

```go
func (r *Request) GetRouteParam(id string) string
```

1. Looks up the parameter name in `r.Route.partIdMap` to get the segment index.
2. Splits `r.URL.Path` and retrieves the segment at that index.
3. Strips `part.prefix` and `part.suffix` from the segment to isolate the captured value.

**Current limitations:**

- Returns `""` for all failure cases: unknown parameter name, nil route, out-of-bounds
  index. Callers cannot distinguish "parameter not present" from "parameter is empty
  string". A future improvement would be to return `(string, bool)` following Go convention.
- The path is re-split on every call. For handlers that extract multiple parameters, this
  means repeated work. A small optimisation would be to split once and cache on the
  `Request`, but this has not been necessary at current scale.

## splitUrlPath

```go
func splitUrlPath(path string) []string
```

A shared utility used by both the router and `GetRouteParam`. It splits on `"/"`,
discards empty segments (handling leading, trailing, and doubled slashes), and guarantees
at least one element by returning `[""]` for the root path.

This lives in `server.go` rather than `routing.go` for historical reasons (it was needed
there first), but it is logically a routing utility. If the files are ever reorganised,
it belongs alongside the routing code.