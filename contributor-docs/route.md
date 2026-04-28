# Routing

**File:** `route.go`

## Overview

Routes are URL patterns mapped to handler functions. A pattern is split into segments
and each segment is one of:

- **Static** - must match exactly (`users`, `settings`)
- **Dynamic** - captures a variable value (`<id>`, `<name>`)
- **Mixed** - a static prefix and/or suffix around a capture (`file-<name>.txt`)

The server stores routes in a `map[int][]*Route` keyed by segment count. A three-segment
request only compares against three-segment routes.

## Data structures

### routePart

```go
type routePart struct {
    id     string
    prefix string
    suffix string
}
```

One per path segment. For a static segment, `id` is empty and `prefix` holds the literal
text. For a dynamic segment, `id` is the parameter name and `prefix`/`suffix` are the
surrounding text (empty strings for a bare `<param>`).

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
    handler   func(request Request) Response

    AllowedMethods []string

    repr string
}
```

`partIdMap` maps parameter names to their index in `parts` so `GetRouteParam` can find a
value in O(1). `AllowedMethods` defaults to `["get"]`. `repr` is the original path string
returned by `ToString()`.

## AddRoute

Parses the path into parts via `createParts`, then appends the route to
`s.routes[len(parts)]`.

If the path produces zero parts, the route is skipped and an error is logged. The server
keeps running.

Parameter names are lowercased. `<ID>` becomes `"id"`. Callers can always use lowercase
in `GetRouteParam`.

`AddRoute` returns `*Route` so you can configure it immediately:

```go
server.AddRoute("/items", handler).AllowedMethods = []string{"get", "post"}
```

## createParts

Calls `splitUrlPath`, then for each segment: if it has no `< >`, it's a static part. If
it does, the regex `(<.+>)` extracts the parameter token and `strings.Index` finds the
surrounding prefix and suffix.

The regex is compiled once at package level.

Only the first `< >` per segment is used. `<a>-<b>` would only capture `a`. If you need
two parameters in one segment, `createParts` is where to extend that.

## FindRoute

1. Splits the incoming path into segments.
2. Looks up routes with the same segment count.
3. Checks each candidate with `matchesRawParts`.
4. Returns the first match, or nil.

`matchesRawParts` checks each segment: it must start with `part.prefix` and end with
`part.suffix`. For static parts the length must be exactly `len(prefix) + len(suffix)`.
For dynamic parts the middle section can be up to 2048 characters.

The 2048 limit is a hard-coded ceiling against very long parameters. It's not configurable.

`FindRoute` only checks structure. Method checking happens in `handleRequest` using
`AllowedMethods`.

## GetRouteParam

Defined on `Request` in `request.go`, documented here because it's really about routing.

```go
func (r *Request) GetRouteParam(id string) (string, bool)
```

1. Looks up the name in `partIdMap` to get the segment index.
2. Splits `r.URL.Path` and gets the segment.
3. Strips prefix and suffix to isolate the captured value.
4. Returns `("", false)` for any failure: unknown name, nil route, out-of-bounds index.

The path is re-split on every call. For handlers that read many parameters this does
repeated work, but it hasn't been worth optimising at current scale.

## splitUrlPath

Splits a path on `"/"`, drops empty segments, and always returns at least one element
(`[""]` for the root path). Used by both route registration and `GetRouteParam`.

This lives in `server.go` rather than `routing.go` for historical reasons (it was needed
there first), but it is logically a routing utility. If the files are ever reorganised,
it belongs alongside the routing code.