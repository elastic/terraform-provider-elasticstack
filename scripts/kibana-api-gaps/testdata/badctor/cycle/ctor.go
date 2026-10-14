package cycle

import "net/http"

func NewOpRequest(server string) (*http.Request, error) {
	return NewOpRequestWithBody(server)
}

func NewOpRequestWithBody(server string) (*http.Request, error) {
	return NewOpRequest(server)
}
