package httpcall

type RequestOptions struct {
	Method    string
	Path      string
	Body      any
	Headers   map[string]string
	SessionID string
}
