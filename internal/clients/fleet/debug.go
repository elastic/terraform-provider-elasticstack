package fleet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const logReqMsg = `%s API Request Details:
---[ REQUEST ]---------------------------------------
%s
-----------------------------------------------------`

const logRespMsg = `%s API Response Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`

var _ http.RoundTripper = &debugRoundTripper{}

type debugRoundTripper struct {
	name      string
	transport http.RoundTripper
}

func newDebugTransport(name string, transport http.RoundTripper) *debugRoundTripper {
	return &debugRoundTripper{
		name:      name,
		transport: transport,
	}
}

func (d *debugRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()
	reqData, err := httputil.DumpRequestOut(r, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logReqMsg, d.name, prettyPrintJsonLines(reqData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API request dump error: %#v", d.name, err))
	}

	resp, err := d.transport.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	respData, err := httputil.DumpResponse(resp, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logRespMsg, d.name, prettyPrintJsonLines(respData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API response dump error: %#v", d.name, err))
	}

	return resp, nil
}

// prettyPrintJsonLines iterates through a []byte line-by-line,
// transforming any lines that are complete json into pretty-printed json.
func prettyPrintJsonLines(b []byte) string {
	parts := strings.Split(string(b), "\n")
	for i, p := range parts {
		if b := []byte(p); json.Valid(b) {
			var out bytes.Buffer
			if err := json.Indent(&out, b, "", " "); err != nil {
				continue
			}
			parts[i] = out.String()
		}
		// Mask Authorization header value
		if strings.Contains(strings.ToLower(p), "authorization:") {
			kv := strings.Split(p, ": ")
			if len(kv) != 2 {
				continue
			}
			kv[1] = strings.Repeat("*", len(kv[1]))
			parts[i] = strings.Join(kv, ": ")
		}
	}
	return strings.Join(parts, "\n")
}
