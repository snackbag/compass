package compass

import "net/url"

type Request struct {
	Method    string
	IP        string
	URL       url.URL
	UserAgent string
}
