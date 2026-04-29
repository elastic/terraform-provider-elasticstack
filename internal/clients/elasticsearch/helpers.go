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

package elasticsearch

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// doFWWrite marshals body to JSON, obtains an ES client from apiClient, calls fn
// with the client and a reader over the serialised body, closes the response body,
// and returns framework diagnostics.
//
// Error messages:
//   - marshalErrMsg: reported when json.Marshal fails
//   - callErrMsg:    reported when fn itself returns an error
//   - responseErrMsg: passed to diagutil.CheckErrorFromFW for HTTP-level errors
func doFWWrite(
	apiClient *clients.ElasticsearchScopedClient,
	body any,
	marshalErrMsg, callErrMsg, responseErrMsg string,
	fn func(*elasticsearch.Client, io.Reader) (*esapi.Response, error),
) fwdiag.Diagnostics {
	b, err := json.Marshal(body)
	if err != nil {
		var diags fwdiag.Diagnostics
		diags.AddError(marshalErrMsg, err.Error())
		return diags
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		var diags fwdiag.Diagnostics
		diags.AddError("Unable to get Elasticsearch client", err.Error())
		return diags
	}

	res, err := fn(esClient, bytes.NewReader(b))
	if err != nil {
		var diags fwdiag.Diagnostics
		diags.AddError(callErrMsg, err.Error())
		return diags
	}
	defer res.Body.Close()

	return diagutil.CheckErrorFromFW(res, responseErrMsg)
}

// doSDKWrite marshals body to JSON, obtains an ES client from apiClient, calls fn
// with the client and a reader over the serialised body, closes the response body,
// and returns SDK diagnostics.
func doSDKWrite(
	apiClient *clients.ElasticsearchScopedClient,
	body any,
	responseErrMsg string,
	fn func(*elasticsearch.Client, io.Reader) (*esapi.Response, error),
) sdkdiag.Diagnostics {
	b, err := json.Marshal(body)
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	esClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	res, err := fn(esClient, bytes.NewReader(b))
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	defer res.Body.Close()

	return diagutil.CheckError(res, responseErrMsg)
}
