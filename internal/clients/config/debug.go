// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package config

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/terraform-provider-elasticstack/internal/debugutils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const logRespMsg = `%s API Response for [%s] Details:
---[ RESPONSE ]--------------------------------------
%s
-----------------------------------------------------`

var _ elastictransport.Logger = &debugLogger{}

type debugLogger struct {
	Name string
}

func (l *debugLogger) LogRoundTrip(req *http.Request, resp *http.Response, err error, _ time.Time, duration time.Duration) error {
	ctx := req.Context()
	requestID := "<nil>"
	if req != nil {
		requestID = fmt.Sprintf("%s %s", req.Method, req.URL)
	}
	tflog.Debug(ctx, fmt.Sprintf("%s request [%s] executed. Took %s. %#v", l.Name, requestID, duration, err))

	if req != nil && req.Body != nil {
		l.logRequest(ctx, req, requestID)
	}

	if resp != nil && resp.Body != nil {
		l.logResponse(ctx, resp, requestID)
	}

	if resp == nil {
		tflog.Debug(ctx, fmt.Sprintf("%s response for [%s] is nil", l.Name, requestID))
	}

	return nil
}

func (l *debugLogger) logRequest(ctx context.Context, req *http.Request, requestID string) {
	defer req.Body.Close()

	reqData, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf("%s request [%s] dump:\n%s", l.Name, requestID, debugutils.PrettyPrintJSONLines(reqData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API request dump error: %#v", l.Name, err))
	}
}

func (l *debugLogger) logResponse(ctx context.Context, resp *http.Response, requestID string) {
	defer resp.Body.Close()

	respData, err := httputil.DumpResponse(resp, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logRespMsg, l.Name, requestID, debugutils.PrettyPrintJSONLines(respData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API response for [%s] dump error: %#v", l.Name, requestID, err))
	}
}

func (l *debugLogger) RequestBodyEnabled() bool  { return true }
func (l *debugLogger) ResponseBodyEnabled() bool { return true }
