package compass

import "slices"

type Route struct {
	segments []segment
	methods  []string
	handler  func(ctx *Context)
}

type segment struct {
	prefix string
	suffix string
}

func (r *Route) SetAllowance(method string, allow bool) {
	if allow && !slices.Contains(r.methods, method) {
		r.methods = append(r.methods, method)
	} else if allow {
		r.methods = slices.DeleteFunc(r.methods, func(s string) bool { return s == method })
	}
}
