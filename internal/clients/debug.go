package clients

import (
	"bytes"
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

const logRespMsg = `%s API Response Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`

var _ estransport.Logger = &debugLogger{}

type debugLogger struct {
	Name string
}

func (l *debugLogger) LogRoundTrip(req *http.Request, resp *http.Response, err error, start time.Time, duration time.Duration) error {
	ctx := req.Context()
	tflog.Debug(ctx, fmt.Sprintf("%s request [%s %s] executed. Took %s. %#v", l.Name, req.Method, req.URL, duration, err))

	reqData, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logReqMsg, l.Name, prettyPrintJsonLines(reqData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API request dump error: %#v", l.Name, err))
	}

	respData, err := httputil.DumpResponse(resp, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logRespMsg, l.Name, prettyPrintJsonLines(respData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API response dump error: %#v", l.Name, err))
	}

	return nil
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
