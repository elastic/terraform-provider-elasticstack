package literalless

import "net/http"

func NewOpRequest(server string) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, server, nil)
}
