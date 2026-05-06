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

package kibanaoapi

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

// SourceMapArtifact is the data returned for each artifact in the list response.
type SourceMapArtifact struct {
	ID             string
	BundleFilepath string
	ServiceName    string
	ServiceVersion string
}

// UploadSourceMapOptions contains the parameters for uploading a source map.
type UploadSourceMapOptions struct {
	SpaceID        string
	BundleFilepath string
	ServiceName    string
	ServiceVersion string
	SourcemapBytes []byte
}

// UploadSourceMap uploads a source map to Kibana and returns the artifact ID.
func UploadSourceMap(ctx context.Context, client *Client, opts UploadSourceMapOptions) (string, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for _, field := range []struct{ name, value string }{
		{"bundle_filepath", opts.BundleFilepath},
		{"service_name", opts.ServiceName},
		{"service_version", opts.ServiceVersion},
	} {
		if err := mw.WriteField(field.name, field.value); err != nil {
			diags.AddError("Failed to build multipart form", "writing field "+field.name+": "+err.Error())
			return "", diags
		}
	}

	filePart, err := mw.CreateFormFile("sourcemap", "sourcemap.js.map")
	if err != nil {
		diags.AddError("Failed to build multipart form", "creating file part: "+err.Error())
		return "", diags
	}
	if _, err := filePart.Write(opts.SourcemapBytes); err != nil {
		diags.AddError("Failed to build multipart form", "writing sourcemap bytes: "+err.Error())
		return "", diags
	}
	if err := mw.Close(); err != nil {
		diags.AddError("Failed to build multipart form", "closing writer: "+err.Error())
		return "", diags
	}

	apiResp, err := client.API.UploadSourceMapWithBodyWithResponse(
		ctx,
		&kbapi.UploadSourceMapParams{
			ElasticApiVersion: kbapi.UploadSourceMapParamsElasticApiVersionN20231031,
		},
		mw.FormDataContentType(),
		&buf,
		kibanautil.SpaceAwarePathRequestEditor(opts.SpaceID),
	)
	if err != nil {
		diags.AddError("Failed to upload APM source map", err.Error())
		return "", diags
	}

	if apiResp.HTTPResponse.StatusCode >= 400 {
		diags.Append(diagutil.ReportUnknownHTTPError(apiResp.HTTPResponse.StatusCode, apiResp.Body)...)
		return "", diags
	}

	if apiResp.JSON200 == nil || apiResp.JSON200.Id == nil {
		diags.AddError("Unexpected response from APM source map upload", "Expected a non-nil id in the upload response.")
		return "", diags
	}

	return *apiResp.JSON200.Id, diags
}

// ListSourceMaps fetches one page of source map artifacts from Kibana.
func ListSourceMaps(ctx context.Context, client *Client, spaceID string, page, perPage float32) ([]SourceMapArtifact, fwdiags.Diagnostics) {
	var diags fwdiags.Diagnostics

	apiResp, err := client.API.GetSourceMapsWithResponse(
		ctx,
		&kbapi.GetSourceMapsParams{
			Page:              &page,
			PerPage:           &perPage,
			ElasticApiVersion: kbapi.GetSourceMapsParamsElasticApiVersionN20231031,
		},
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		diags.AddError("Failed to list APM source maps", err.Error())
		return nil, diags
	}

	if apiResp.HTTPResponse.StatusCode >= 400 {
		diags.Append(diagutil.ReportUnknownHTTPError(apiResp.HTTPResponse.StatusCode, apiResp.Body)...)
		return nil, diags
	}

	if apiResp.JSON200 == nil || apiResp.JSON200.Artifacts == nil {
		return nil, diags
	}

	var results []SourceMapArtifact
	for _, a := range *apiResp.JSON200.Artifacts {
		if a.Id == nil {
			continue
		}
		art := SourceMapArtifact{ID: *a.Id}
		if a.Body != nil {
			if a.Body.BundleFilepath != nil {
				art.BundleFilepath = *a.Body.BundleFilepath
			}
			if a.Body.ServiceName != nil {
				art.ServiceName = *a.Body.ServiceName
			}
			if a.Body.ServiceVersion != nil {
				art.ServiceVersion = *a.Body.ServiceVersion
			}
		}
		results = append(results, art)
	}
	return results, diags
}

// DeleteSourceMap deletes a source map artifact by ID. HTTP 404 is treated as success.
func DeleteSourceMap(ctx context.Context, client *Client, spaceID, artifactID string) fwdiags.Diagnostics {
	var diags fwdiags.Diagnostics

	apiResp, err := client.API.DeleteSourceMapWithResponse(
		ctx,
		artifactID,
		&kbapi.DeleteSourceMapParams{
			ElasticApiVersion: kbapi.N20231031,
		},
		kibanautil.SpaceAwarePathRequestEditor(spaceID),
	)
	if err != nil {
		diags.AddError("Failed to delete APM source map", err.Error())
		return diags
	}

	if apiResp.HTTPResponse.StatusCode == http.StatusNotFound {
		return diags
	}

	if apiResp.HTTPResponse.StatusCode >= 400 {
		diags.Append(diagutil.ReportUnknownHTTPError(apiResp.HTTPResponse.StatusCode, apiResp.Body)...)
	}

	return diags
}
