package compass

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// SameSite controls the SameSite attribute of a cookie, which determines
// when the browser will send it on cross-site requests.
type SameSite int

const (
	// SameSiteDefault omits the SameSite attribute entirely.
	SameSiteDefault SameSite = iota
	// SameSiteLax allows the cookie on top-level navigations and same-site
	// requests. The recommended default for most applications.
	SameSiteLax
	// SameSiteStrict only sends the cookie on same-site requests.
	SameSiteStrict
	// SameSiteNone allows the cookie on all requests, including cross-site.
	// Requires Secure to be true, because browsers otherwise reject it.
	SameSiteNone
)

// Cookie represents an HTTP cookie to be sent to the client via Set-Cookie.
//
// Only Name and Value are required. All other fields have sensible defaults:
// Path defaults to "/", and all boolean flags default to false.
//
// To delete a cookie, use Response.DeleteCookie rather than constructing
// a deletion cookie manually.
type Cookie struct {
	Name  string
	Value string

	// MaxAge is the lifetime of the cookie in seconds.
	// 0 means no Max-Age attribute is set (session cookie).
	// A negative value deletes the cookie immediately.
	MaxAge int

	// Expires sets an absolute expiry time. Ignored if zero.
	// Prefer MaxAge over Expires for new code — MaxAge is unambiguous.
	Expires time.Time

	// Path scopes the cookie to a URL path prefix.
	// Defaults to "/" if empty.
	Path string

	// Domain scopes the cookie to a domain. Omitted if empty.
	Domain string

	// Secure instructs the browser to only send the cookie over HTTPS.
	Secure bool

	// HttpOnly prevents JavaScript from accessing the cookie via document.cookie.
	// Should be true for any cookie that does not need to be read client-side.
	HttpOnly bool

	// SameSite controls cross-site request behaviour.
	// Defaults to SameSiteDefault (attribute omitted) if not set.
	SameSite SameSite
}

// toHeader serialises the cookie to a valid Set-Cookie header value.
func (c Cookie) toHeader() string {
	path := c.Path
	if path == "" {
		path = "/"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s=%s", c.Name, c.Value)
	fmt.Fprintf(&b, "; Path=%s", path)

	if c.MaxAge < 0 {
		b.WriteString("; Max-Age=0")
	} else if c.MaxAge > 0 {
		fmt.Fprintf(&b, "; Max-Age=%d", c.MaxAge)
	}

	if !c.Expires.IsZero() {
		fmt.Fprintf(&b, "; Expires=%s", c.Expires.UTC().Format(http.TimeFormat))
	}

	if c.Domain != "" {
		fmt.Fprintf(&b, "; Domain=%s", c.Domain)
	}

	if c.Secure {
		b.WriteString("; Secure")
	}

	if c.HttpOnly {
		b.WriteString("; HttpOnly")
	}

	switch c.SameSite {
	case SameSiteLax:
		b.WriteString("; SameSite=Lax")
	case SameSiteStrict:
		b.WriteString("; SameSite=Strict")
	case SameSiteNone:
		b.WriteString("; SameSite=None")
	default:
	}

	return b.String()
}
