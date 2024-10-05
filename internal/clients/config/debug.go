package config

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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

var _ elastictransport.Logger = &debugLogger{}

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

	if req != nil && req.Body != nil {
		l.logRequest(ctx, req, requestId)
	}

	if resp != nil && resp.Body != nil {
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
		tflog.Debug(ctx, fmt.Sprintf(logReqMsg, l.Name, utils.PrettyPrintJSONLines(reqData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API request dump error: %#v", l.Name, err))
	}
}

func (l *debugLogger) logResponse(ctx context.Context, resp *http.Response, requestId string) {
	defer resp.Body.Close()

	respData, err := httputil.DumpResponse(resp, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logRespMsg, l.Name, requestId, utils.PrettyPrintJSONLines(respData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API response for [%s] dump error: %#v", l.Name, requestId, err))
	}
}

func (l *debugLogger) RequestBodyEnabled() bool  { return true }
func (l *debugLogger) ResponseBodyEnabled() bool { return true }
