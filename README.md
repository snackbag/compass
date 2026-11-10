# Compass

A Go http wrapper featuring sessions and basic routes.

This project started when we began moving from Flask (which is Python) to Go, but there was no 
suitable replacement.

### The concept

Every route handler is a function. It gets a `Request` and returns a `Response`. Compass takes
that `Response` does all the boring backend stuff for you.

```go
func handleHello(r compass.Request) compass.Response {
    return compass.Text("hello")
}
```

### Getting started

```go
server := compass.NewServer(compass.NewStandardConfiguration())

server.AddRoute("/", func (r compass.Request) compass.Response {
    return compass.Text("hello")
})

server.MustRun()
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

You can also dump or load the config struct in JSON.

### Routes

Wrap a segment in `< >` to make it a parameter:

```go
server.AddRoute("/users/<id>", func (r compass.Request) compass.Response {
    id, _ := r.GetRouteParam("id")
    return compass.Text("user: " + id)
})
```

Compass manages route method safety for you.

```go
server.AddRoute("/items", handler).AllowedMethods = []string{"get", "post"}
```

### Responses

```go
compass.Text("hello")                    // 200, just text
compass.TextWithCode("nope", 403)        // any status
compass.HTML("<p>hiii</p>")              // 200, displayed as HTML
compass.JsonMarshal(myStruct)            // marshals to JSON, 200
compass.Redirect("/login", false)        // 303 redirect
compass.DownloadFile("report.pdf", path) // triggers a file download
compass.ServeFile(path, "photo.jpg")     // serves picture
```

Headers and cookies also work with the responses:

```go
resp := compass.JsonMarshal(data)
resp.Headers["X-Request-Id"] = "abc"
resp.SetCookie(compass.Cookie{Name: "theme", Value: "light"})
return resp
```

### Static files and templating

Anything in `assets/static/` is served under `/static/` automatically. Both paths are
configurable.

Compass does not come with its own templating engine. We recommend using something like 
[fasttemplate](https://github.com/valyala/fasttemplate) and serving its output as HTML. It is good practice to put your templates in 
the `assets/templates/` directory, but at the end of the day, it's up to you.

### Sessions

Sessions are stored as JSON files in `.compass/session/`

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

### CORS

```go
return compass.JsonMarshal(data).WithCORS(compass.AllowAll())
```

And custom policies

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
