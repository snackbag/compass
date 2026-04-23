package compass

import (
	"fmt"
	"strings"
)

// CORSPolicy defines the CORS headers to attach to a response.
//
// Fields left at their zero value are omitted from the response headers.
// Use AllowAll for a permissive default suitable for development.
type CORSPolicy struct {
	Origin      string
	Methods     []string
	Headers     []string
	Credentials bool
	MaxAge      int
}

// AllowAll returns a permissive CORSPolicy that allows any origin,
// common HTTP methods, and all headers.
//
// Suitable for development or fully public APIs. Not recommended for
// production unless you explicitly want open access.
func AllowAll() CORSPolicy {
	return CORSPolicy{
		Origin:  "*",
		Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		Headers: []string{"*"},
	}
}

// Apply attaches the policy's CORS headers to the given Response
// and returns it.
func (c CORSPolicy) Apply(r Response) Response {
	if c.Origin != "" {
		r.Headers["Access-Control-Allow-Origin"] = c.Origin
	}

	if len(c.Methods) > 0 {
		r.Headers["Access-Control-Allow-Methods"] = strings.Join(c.Methods, ", ")
	}

	if len(c.Headers) > 0 {
		r.Headers["Access-Control-Allow-Headers"] = strings.Join(c.Headers, ", ")
	}

	if c.Credentials {
		r.Headers["Access-Control-Allow-Credentials"] = "true"
	}

	if c.MaxAge > 0 {
		r.Headers["Access-Control-Max-Age"] = fmt.Sprintf("%d", c.MaxAge)
	}

	return r
}

// WithCORS attaches the given CORSPolicy to this response and returns it.
func (r Response) WithCORS(policy CORSPolicy) Response {
	return policy.Apply(r)
}
