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
	"context"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const tagListMaxPerPage float32 = 100

// TagListMaxPerPage returns the maximum page size used when listing tags.
func TagListMaxPerPage() float32 {
	return tagListMaxPerPage
}

// TagDetail is the unwrapped tag payload from single-tag API responses.
type TagDetail struct {
	ID          string
	Name        string
	Color       string
	Description *string
	CreatedAt   *string
	UpdatedAt   *string
	Managed     *bool
}

// TagListResult is a page of tags from GET /api/tags.
type TagListResult struct {
	Tags  []TagDetail
	Total float32
	Page  float32
}

// GetTag reads a tag by ID. Returns (nil, nil) on HTTP 404.
func GetTag(ctx context.Context, client *Client, spaceID, id string) (*TagDetail, diag.Diagnostics) {
	resp, err := client.API.GetTagsIdWithResponse(ctx, id, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.ErrDiag(fmt.Sprintf("HTTP request failed reading tag %q", id), err)
	}

	tagResp, diags := HandleGetTypedResponse(resp.StatusCode(), resp.Body,
		func() *struct {
			Data kbapi.KibanaHTTPAPIsKbnTagsAttributes `json:"data"`
			Id   string                                `json:"id"`
			Meta kbapi.KibanaHTTPAPIsKbnAsCodeMeta     `json:"meta"`
		} {
			if resp.JSON200 == nil {
				return nil
			}
			return &struct {
				Data kbapi.KibanaHTTPAPIsKbnTagsAttributes `json:"data"`
				Id   string                                `json:"id"`
				Meta kbapi.KibanaHTTPAPIsKbnAsCodeMeta     `json:"meta"`
			}{
				Data: resp.JSON200.Data,
				Id:   resp.JSON200.Id,
				Meta: resp.JSON200.Meta,
			}
		})
	if diags.HasError() || tagResp == nil {
		return nil, diags
	}

	return tagDetailFromResponse(tagResp.Id, tagResp.Data, tagResp.Meta), nil
}

// CreateTag creates a tag via POST /api/tags.
func CreateTag(ctx context.Context, client *Client, spaceID string, body kbapi.PostTagsJSONRequestBody) (*TagDetail, diag.Diagnostics) {
	resp, err := client.API.PostTagsWithResponse(ctx, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.ErrDiag("HTTP request failed creating tag", err)
	}

	createResp, diags := HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *struct {
			Data kbapi.KibanaHTTPAPIsKbnTagsAttributes `json:"data"`
			Id   string                                `json:"id"`
			Meta kbapi.KibanaHTTPAPIsKbnAsCodeMeta     `json:"meta"`
		} {
			if resp.JSON201 == nil {
				return nil
			}
			return &struct {
				Data kbapi.KibanaHTTPAPIsKbnTagsAttributes `json:"data"`
				Id   string                                `json:"id"`
				Meta kbapi.KibanaHTTPAPIsKbnAsCodeMeta     `json:"meta"`
			}{
				Data: resp.JSON201.Data,
				Id:   resp.JSON201.Id,
				Meta: resp.JSON201.Meta,
			}
		}, http.StatusCreated)
	if diags.HasError() {
		return nil, diags
	}

	return tagDetailFromResponse(createResp.Id, createResp.Data, createResp.Meta), nil
}

// UpsertTag creates or updates a tag via PUT /api/tags/{id}.
func UpsertTag(ctx context.Context, client *Client, spaceID, id string, body kbapi.PutTagsIdJSONRequestBody) (*TagDetail, diag.Diagnostics) {
	resp, err := client.API.PutTagsIdWithResponse(ctx, id, body, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.ErrDiag(fmt.Sprintf("HTTP request failed upserting tag %q", id), err)
	}

	upsertResp, diags := HandleMutateTypedResponse(resp.StatusCode(), resp.Body,
		func() *struct {
			Data kbapi.KibanaHTTPAPIsKbnTagsAttributes `json:"data"`
			Id   string                                `json:"id"`
			Meta kbapi.KibanaHTTPAPIsKbnAsCodeMeta     `json:"meta"`
		} {
			switch {
			case resp.JSON200 != nil:
				return &struct {
					Data kbapi.KibanaHTTPAPIsKbnTagsAttributes `json:"data"`
					Id   string                                `json:"id"`
					Meta kbapi.KibanaHTTPAPIsKbnAsCodeMeta     `json:"meta"`
				}{
					Data: resp.JSON200.Data,
					Id:   resp.JSON200.Id,
					Meta: resp.JSON200.Meta,
				}
			case resp.JSON201 != nil:
				return &struct {
					Data kbapi.KibanaHTTPAPIsKbnTagsAttributes `json:"data"`
					Id   string                                `json:"id"`
					Meta kbapi.KibanaHTTPAPIsKbnAsCodeMeta     `json:"meta"`
				}{
					Data: resp.JSON201.Data,
					Id:   resp.JSON201.Id,
					Meta: resp.JSON201.Meta,
				}
			default:
				return nil
			}
		}, http.StatusOK, http.StatusCreated)
	if diags.HasError() {
		return nil, diags
	}

	return tagDetailFromResponse(upsertResp.Id, upsertResp.Data, upsertResp.Meta), nil
}

// DeleteTag deletes a tag by ID. HTTP 404 is treated as success.
func DeleteTag(ctx context.Context, client *Client, spaceID, id string) diag.Diagnostics {
	resp, err := client.API.DeleteTagsIdWithResponse(ctx, id, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return diagutil.ErrDiag(fmt.Sprintf("HTTP request failed deleting tag %q", id), err)
	}

	return diagutil.HandleStatusResponse(resp.StatusCode(), resp.Body, http.StatusOK, http.StatusNotFound)
}

// ListTags fetches a single page of tags from GET /api/tags.
func ListTags(ctx context.Context, client *Client, spaceID string, params *kbapi.GetTagsParams) (*TagListResult, diag.Diagnostics) {
	resp, err := client.API.GetTagsWithResponse(ctx, params, kibanautil.SpaceAwarePathRequestEditor(spaceID))
	if err != nil {
		return nil, diagutil.ErrDiag("HTTP request failed listing tags", err)
	}

	listResp, diags := HandleGetTypedResponse(resp.StatusCode(), resp.Body,
		func() *kbapi.GetTagsResponse {
			if resp.JSON200 == nil {
				return nil
			}
			return resp
		})
	if diags.HasError() || listResp == nil || listResp.JSON200 == nil {
		return nil, diags
	}

	payload := listResp.JSON200
	result := &TagListResult{
		Total: payload.Meta.Total,
	}
	if payload.Meta.Page != nil {
		result.Page = *payload.Meta.Page
	}

	for _, item := range payload.Data {
		result.Tags = append(result.Tags, *tagDetailFromResponse(item.Id, item.Data, item.Meta))
	}

	return result, nil
}

func tagDetailFromResponse(id string, data kbapi.KibanaHTTPAPIsKbnTagsAttributes, meta kbapi.KibanaHTTPAPIsKbnAsCodeMeta) *TagDetail {
	detail := &TagDetail{
		ID:    id,
		Name:  data.Name,
		Color: data.Color,
	}
	if data.Description != nil {
		desc := *data.Description
		detail.Description = &desc
	}
	if meta.CreatedAt != nil {
		createdAt := *meta.CreatedAt
		detail.CreatedAt = &createdAt
	}
	if meta.UpdatedAt != nil {
		updatedAt := *meta.UpdatedAt
		detail.UpdatedAt = &updatedAt
	}
	if meta.Managed != nil {
		managed := *meta.Managed
		detail.Managed = &managed
	}
	return detail
}
