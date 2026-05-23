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
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
)

func PutComponentTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, template *models.ComponentTemplate) fwdiags.Diagnostics {
	typedClient, d := apiClient.GetESClientDiag()
	if d.HasError() {
		return d
	}

	templateBytes, err := json.Marshal(template)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	_, err = typedClient.Cluster.PutComponentTemplate(template.Name).Raw(bytes.NewReader(templateBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

// GetComponentTemplate returns a component template by name.
//
// We use .Perform() instead of .Do() because the typed client decodes the
// response through *types.ComponentTemplateSummary, whose Settings field is
// map[string]IndexSettings and whose hand-coded UnmarshalJSON implementations
// silently drop any setting sub-key not modeled by the typed structs (e.g.
// index.search.slowlog.include) and coerce string-encoded scalars such as
// index.lifecycle.parse_origination_date "true" into typed bool. Decoding
// into models.ComponentTemplate preserves the raw JSON shape Elasticsearch
// returns. See issue #3124.
func GetComponentTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) (*models.ComponentTemplateResponse, fwdiags.Diagnostics) {
	typedClient, d := apiClient.GetESClientDiag()
	if d.HasError() {
		return nil, d
	}
	res, err := typedClient.Cluster.GetComponentTemplate().Name(templateName).Perform(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if d := diagutil.CheckHTTPErrorFromFW(res, "Unable to find component template on cluster"); d.HasError() {
		return nil, d
	}

	var resp struct {
		ComponentTemplates []models.ComponentTemplateResponse `json:"component_templates"`
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if len(resp.ComponentTemplates) != 1 {
		detail := fmt.Sprintf("Elasticsearch API returned %d when requested '%s' component template.", len(resp.ComponentTemplates), templateName)
		return nil, fwdiags.Diagnostics{fwdiags.NewErrorDiagnostic("Wrong number of templates returned", detail)}
	}
	tpl := resp.ComponentTemplates[0]
	return &tpl, nil
}

func DeleteComponentTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) fwdiags.Diagnostics {
	typedClient, d := apiClient.GetESClientDiag()
	if d.HasError() {
		return d
	}
	_, err := typedClient.Cluster.DeleteComponentTemplate(templateName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func PutIndexTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, template *models.IndexTemplate) fwdiags.Diagnostics {
	typedClient, d := apiClient.GetESClientDiag()
	if d.HasError() {
		return d
	}

	templateBytes, err := json.Marshal(template)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}

	_, err = typedClient.Indices.PutIndexTemplate(template.Name).Raw(bytes.NewReader(templateBytes)).Do(ctx)
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}

func GetIndexTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) (*models.IndexTemplateResponse, fwdiags.Diagnostics) {
	typedClient, d := apiClient.GetESClientDiag()
	if d.HasError() {
		return nil, d
	}

	// We use .Perform() instead of .Do() because the typed client decodes the
	// response through *types.IndexSettings, whose hand-coded UnmarshalJSON
	// implementations silently drop any field not present in the typed structs
	// (e.g. index.search.slowlog.include) and coerce string-encoded scalars to
	// their typed form (e.g. index.lifecycle.parse_origination_date "true" ->
	// bool true). Decoding into models.IndexTemplate preserves the raw JSON
	// shape Elasticsearch returns. See issue #3124.
	res, err := typedClient.Indices.GetIndexTemplate().Name(templateName).Perform(ctx)
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if d := diagutil.CheckHTTPErrorFromFW(res, "Unable to find index template on cluster"); d.HasError() {
		return nil, d
	}

	var resp models.IndexTemplatesResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	if len(resp.IndexTemplates) != 1 {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(
				"Wrong number of templates returned",
				fmt.Sprintf("Elasticsearch API returned %d when requested '%s' template.", len(resp.IndexTemplates), templateName),
			),
		}
	}
	tpl := resp.IndexTemplates[0]
	return &tpl, nil
}

func DeleteIndexTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) fwdiags.Diagnostics {
	typedClient, d := apiClient.GetESClientDiag()
	if d.HasError() {
		return d
	}
	_, err := typedClient.Indices.DeleteIndexTemplate(templateName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}
