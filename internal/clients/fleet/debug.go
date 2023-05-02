package fleet

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
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
		tflog.Debug(ctx, fmt.Sprintf(logReqMsg, d.name, utils.PrettyPrintJSONLines(reqData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API request dump error: %#v", d.name, err))
	}

	resp, err := d.transport.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	respData, err := httputil.DumpResponse(resp, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logRespMsg, d.name, utils.PrettyPrintJSONLines(respData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API response dump error: %#v", d.name, err))
	}

	return resp, nil
}
