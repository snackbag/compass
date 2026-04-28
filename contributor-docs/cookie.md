# Cookie

**File:** `cookie.go`

## Overview

`Cookie` is a struct for building `Set-Cookie` response headers. Cookies are attached to
a `Response` via `SetCookie` and written by the framework before the response body.

Reading cookies from an incoming request is handled by `Request.GetCookie` and
`Request.GetCookies`, which delegate to the standard library.

## SameSite

```go
type SameSite int

const (
    SameSiteDefault SameSite = iota  // omit the attribute
    SameSiteLax
    SameSiteStrict
    SameSiteNone                     // requires Secure: true
)
```

`SameSiteDefault` (zero value) omits the `SameSite` attribute. Browsers typically apply
`Lax` behaviour when it's absent, but set it explicitly if you want a predictable result.

`SameSiteNone` requires `Secure: true`, because browsers reject it otherwise. The framework
doesn't enforce this, but logs a warning when writing the request.

## Cookie

```go
type Cookie struct {
    Name  string
    Value string

    MaxAge  int
    Expires time.Time

    Path     string
    Domain   string
    Secure   bool
    HttpOnly bool
    SameSite SameSite
}
```

Only `Name` and `Value` are required.

| Field      | Zero value behaviour                            |
|------------|-------------------------------------------------|
| `MaxAge`   | `Max-Age` omitted (session cookie)              |
| `Expires`  | `Expires` omitted                               |
| `Path`     | defaults to `"/"` in `toHeader`                 |
| `Domain`   | `Domain` omitted                                |
| `Secure`   | flag omitted                                    |
| `HttpOnly` | flag omitted                                    |
| `SameSite` | attribute omitted                               |

A negative `MaxAge` serialises to `Max-Age=0`, which deletes the cookie. That's what
`Response.RemoveCookie` uses.

Prefer `MaxAge` over `Expires`. `MaxAge` is relative to the client's clock and
unambiguous. `Expires` depends on the server and client clocks being in sync.

## toHeader

An unexported method that serialises a `Cookie` to a `Set-Cookie` header value string.
Called by `writeCookies` in `server.go`. The order is: `name=value`, then `Path`, then
the remaining attributes.

## Reading cookies

```go
value, ok := request.GetCookie("session_id")
all := request.GetCookies() // map[string]string
```

Incoming cookies only carry name and value. No `Path`, `Expires`, or `HttpOnly`.