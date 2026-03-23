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
	reqData, err := httputil.DumpRequestOut(r, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logReqMsg, d.name, PrettyPrintJSONLines(reqData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API request dump error: %#v", d.name, err))
	}

	resp, err := d.transport.RoundTrip(r)
	if err != nil {
		return resp, err
	}

	respData, err := httputil.DumpResponse(resp, true)
	if err == nil {
		tflog.Debug(ctx, fmt.Sprintf(logRespMsg, d.name, PrettyPrintJSONLines(respData)))
	} else {
		tflog.Debug(ctx, fmt.Sprintf("%s API response dump error: %#v", d.name, err))
	}

	return resp, nil
}
