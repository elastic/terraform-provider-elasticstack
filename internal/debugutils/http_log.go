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

package debugutils

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"

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

func NewDebugTransport(name string, transport http.RoundTripper) http.RoundTripper {
	return &debugRoundTripper{
		name:      name,
		transport: transport,
	}
}

func (d *debugRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {

	ctx := r.Context()
	LogHTTPRequest(ctx, d.name, r)

	resp, err := d.transport.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	LogHTTPResponse(ctx, d.name, resp)

	return resp, nil
}

// LogHTTPRequest dumps an outgoing HTTP request and emits it via tflog.Debug,
// pretty-printing any JSON body. name identifies the API/client the request
// belongs to and is used to label the log entry.
func LogHTTPRequest(ctx context.Context, name string, req *http.Request) {
	reqData, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logReqMsg, name, PrettyPrintJSONLines(reqData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API request dump error: %#v", name, err))
	}
}

// LogHTTPResponse dumps an HTTP response and emits it via tflog.Debug,
// pretty-printing any JSON body. name identifies the API/client the response
// belongs to and is used to label the log entry.
func LogHTTPResponse(ctx context.Context, name string, resp *http.Response) {
	respData, err := httputil.DumpResponse(resp, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logRespMsg, name, PrettyPrintJSONLines(respData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API response dump error: %#v", name, err))
	}
}
