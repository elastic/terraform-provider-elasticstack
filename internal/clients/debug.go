package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7/estransport"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const logReqMsg = `%s API Request Details:
---[ REQUEST ]---------------------------------------
%s
-----------------------------------------------------`

const logRespMsg = `%s API Response for [%s] Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`

var _ estransport.Logger = &debugLogger{}

type debugLogger struct {
	Name string
}

func (l *debugLogger) LogRoundTrip(req *http.Request, resp *http.Response, err error, start time.Time, duration time.Duration) error {
	ctx := req.Context()
	requestId := "<nil>"
	if req != nil {
		requestId = fmt.Sprintf("%s %s", req.Method, req.URL)
	}
	tflog.Debug(ctx, fmt.Sprintf("%s request [%s] executed. Took %s. %#v", l.Name, requestId, duration, err))

	if req != nil {
		l.logRequest(ctx, req, requestId)
	}

	if resp != nil {
		l.logResponse(ctx, resp, requestId)
	}

	if resp == nil {
		tflog.Debug(ctx, fmt.Sprintf("%s response for [%s] is nil", l.Name, requestId))
	}

	return nil
}

func (l *debugLogger) logRequest(ctx context.Context, req *http.Request, requestId string) {
	defer req.Body.Close()

	reqData, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logReqMsg, l.Name, prettyPrintJsonLines(reqData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API request dump error: %#v", l.Name, err))
	}
}

func (l *debugLogger) logResponse(ctx context.Context, resp *http.Response, requestId string) {
	defer resp.Body.Close()

	respData, err := httputil.DumpResponse(resp, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logRespMsg, l.Name, requestId, prettyPrintJsonLines(respData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API response for [%s] dump error: %#v", l.Name, requestId, err))
	}
}

func (l *debugLogger) RequestBodyEnabled() bool  { return true }
func (l *debugLogger) ResponseBodyEnabled() bool { return true }

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
