# CORS

**File:** `cors.go`

## Overview

CORS support in Compass is implemented as a plain value type that attaches HTTP headers to
a `Response`. There is no global middleware, no automatic preflight handling, and no hidden
configuration. Callers apply a policy explicitly where they want it.

## CORSPolicy

```go
type CORSPolicy struct {
  Origin      string
  Methods     []string
  Headers     []string
  Credentials bool
  MaxAge      int
}
```

Each field maps directly to a CORS response header:

| Field         | Header                             | Notes                                       |
|---------------|------------------------------------|---------------------------------------------|
| `Origin`      | `Access-Control-Allow-Origin`      | Omitted if empty string                     |
| `Methods`     | `Access-Control-Allow-Methods`     | Joined with `", "`; omitted if nil/empty    |
| `Headers`     | `Access-Control-Allow-Headers`     | Joined with `", "`; omitted if nil/empty    |
| `Credentials` | `Access-Control-Allow-Credentials` | Set to `"true"` only if the field is `true` |
| `MaxAge`      | `Access-Control-Max-Age`           | Omitted if zero or negative                 |

Fields at their zero value are silently skipped, but a partial policy is valid. This means
you can set only `Origin` if that is all you need.

## AllowAll

```go
func AllowAll() CORSPolicy
```

Returns a fully permissive policy for development or public APIs:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: *
```

`Credentials` is false and `MaxAge` is zero in this preset. Note that browsers will reject
a credentialed request (cookies, HTTP auth) when `Origin` is `"*"`. If you need
credentials, you must set a specific origin and set `Credentials: true`.

## Apply and WithCORS

There are two ways to attach a policy to a response, depending on style preference:

`policy.Apply(resp)` takes a `Response` by value and returns a new one with headers
added. Use this when you have a policy variable and want to apply it to multiple responses:

```go
policy := compass.CORSPolicy{Origin: "https://myapp.com"}
return policy.Apply(compass.JsonMarshal(data))
```

`resp.WithCORS(policy)` is the same operation in method form. Use this for inline
chaining:

```go
return compass.Text("ok").WithCORS(compass.AllowAll())
```

Both are equivalent. `WithCORS` calls `Apply` internally.

## Preflight requests

`OPTIONS` preflight requests are not handled automatically. A route must be registered for
`OPTIONS` explicitly if your application needs to respond to preflight checks. This is a
deliberate omission since automatic preflight handling requires the framework to understand
method-based routing, which Compass does not currently implement.

A simple manual approach until method routing is added:

```go
server.AddRoute("/api/data", func (r compass.Request) compass.Response {
  if r.Method == "options" {
    return compass.TextWithCode("", 204).WithCORS(compass.AllowAll())
  }
  // ... normal handler
})
```

When method-based routing is implemented, automatic preflight handling can be added to the
dispatch layer without changing the `CORSPolicy` API.

## Applying CORS globally

There is currently no global CORS middleware. The closest equivalent is a helper function in
your own application code:

```go
var policy = compass.CORSPolicy{Origin: "https://myapp.com"}

func withCors(resp compass.Response) compass.Response {
    return policy.Apply(resp)
}

// In handlers:
return withCors(compass.JsonMarshal(data))
```

If a middleware system is added to Compass in the future, global CORS would be a natural first
use case. The `CORSPolicy` type would not need to change, only the application point would
move from individual responses to a middleware layer.

## Security notes for contributors

- `Origin: "*"` and `Credentials: true` together are rejected by all major browsers. The
  framework does not enforce this constraint, meaning it is the caller's responsibility. Consider
  adding a validation check or at minimum a warning log if both are set.
- `AllowAll()` is explicitly named to make its permissiveness obvious. Do not rename it to
  something that sounds safe.
- `MaxAge` controls how long browsers cache preflight results. A high value (e.g. 86400)
  reduces preflight overhead but means policy changes take longer to propagate to clients.