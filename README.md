# Compass

A Go http wrapper for quick development.

### Features

- Kinda extensible
- Session management
- Resource management
- Questionable caching
- One dependency

Internally, there is also:

- Documentation on how things work and are implemented [/contributor-docs](/contributor-docs)
- ~~Testing suite~~ hmmmmmm

### Why?

There are probably better wrappers out there. Heck, even better Go web server implementations. And
faster, too. This project was started to move from Python (Flask) to Go, but no suitable
replacement was found. Therefore, it was decided to make our own, because reinventing the wheel is
always such a good and well-thought-out decision™

## Documentation

There are docstrings everywhere, and if you want to contribute you should check out our [contributor
documentation](/contributor-docs). Otherwise, here's a quick introduction.

### The basic idea

Every route handler is a function. It gets a `Request` and returns a `Response`. Compass takes
that `Response` and does the actual writing to the client, so you never touch `http.ResponseWriter`.

```go
func handleHello(r compass.Request) compass.Response {
return compass.Text("hello")
}
```

`compass.Text("hello")` builds a response value, and Compass does the rest, just like Flask.

### Getting started

```go
server := compass.NewServer(compass.NewStandardConfiguration())

server.AddRoute("/", func (r compass.Request) compass.Response {
return compass.Text("hello")
})

server.MustRun() // listens on :3000
```

`NewStandardConfiguration()` defaults:

| Option                | Default      | Note                                      |
|-----------------------|--------------|-------------------------------------------|
| `Port`                | `3000`       |                                           |
| `AssetDir`            | `"assets"`   | root for static files                     |
| `StaticUrl`           | `"/static"`  | URL prefix for static files               |
| `CompassDir`          | `".compass"` | where sessions and other state are stored |
| `SessionExpiryTime`   | `259200000`  | ms; 72 hours                              |
| `SessionTickInterval` | `300000`     | ms; how often we check for session expiry |

You can also dump or load the config struct in JSON, because the configuration has corresponding 
field tags.

### Routes

Wrap a segment in `< >` to make it a parameter:

```go
server.AddRoute("/users/<id>", func (r compass.Request) compass.Response {
    id, _ := r.GetRouteParam("id")
    return compass.Text("user: " + id)
})
```

Routes accept `GET` only by default. Change that on the returned route:

```go
server.AddRoute("/items", handler).AllowedMethods = []string{"get", "post"}
```

### Responses

```go
compass.Text("hello")                    // 200, just text
compass.TextWithCode("nope", 403)        // any status
compass.JsonMarshal(myStruct)            // marshals to JSON, 200
compass.Redirect("/login", false)        // 303 redirect
compass.DownloadFile("report.pdf", path) // triggers a file download
compass.ServeFile(path, "photo.jpg")     // serves picture
```

Headers and cookies go directly on the response value:

```go
resp := compass.JsonMarshal(data)
resp.Headers["X-Request-Id"] = "abc"
resp.SetCookie(compass.Cookie{Name: "theme", Value: "light"})
return resp
```

### Sessions

Sessions are stored as JSON files in `.compass/session/`, so they survive restarts.

```go
session, err := server.CreateSession()
resp := compass.Redirect("/", false)
resp.SetSession(session)
return resp

// get it back on the next request
session, ok := r.GetSession(server)
 
// read a value
name, err := compass.SessionGet[string](session, "name")
pfpLink := compass.SessionGetOrDefault[string](session, "profilePicture", "/static/default.png")
 
// write a value
tx := session.BeginTx()
tx.Set("name", "alice")
tx.Commit()
```

### Static files

Anything in `assets/static/` is served under `/static/` automatically. Both paths are
configurable.

### CORS

```go
return compass.JsonMarshal(data).WithCORS(compass.AllowAll())
```

`AllowAll()` is wide open, which is fine locally, but probably not for production. For production,
set only what you need:

```go
policy := compass.CORSPolicy{Origin: "https://example.com", Methods: []string{"get", "post"}}
return compass.Text("I'm a very secretive text!").WithCORS(policy)
```

### Custom 404 / 405

```go
server.NotFoundHandler = func(r compass.Request) compass.Response {
    return compass.TextWithCode("not found", 404)
}
server.MethodNotAllowedHandler = func(r compass.Request) compass.Response {
    return compass.TextWithCode("method not allowed", 405)
}
```
