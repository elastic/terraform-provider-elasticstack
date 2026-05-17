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

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	sdkdiag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func PutComponentTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, template *models.ComponentTemplate) sdkdiag.Diagnostics {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	templateBytes, err := json.Marshal(template)
	if err != nil {
		return sdkdiag.FromErr(err)
	}

	_, err = typedClient.Cluster.PutComponentTemplate(template.Name).Raw(bytes.NewReader(templateBytes)).Do(ctx)
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	return nil
}

// GetComponentTemplate returns a component template by name.
func GetComponentTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) (*types.ClusterComponentTemplate, sdkdiag.Diagnostics) {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, sdkdiag.FromErr(err)
	}
	res, err := typedClient.Cluster.GetComponentTemplate().Name(templateName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, sdkdiag.FromErr(err)
	}
	if len(res.ComponentTemplates) != 1 {
		detail := fmt.Sprintf("Elasticsearch API returned %d when requested '%s' component template.", len(res.ComponentTemplates), templateName)
		return nil, diagutil.SDKErrorDiag("Wrong number of templates returned", detail)
	}
	tpl := res.ComponentTemplates[0]
	return &tpl, nil
}

func DeleteComponentTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) sdkdiag.Diagnostics {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return sdkdiag.FromErr(err)
	}
	_, err = typedClient.Cluster.DeleteComponentTemplate(templateName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return sdkdiag.FromErr(err)
	}
	return nil
}

func PutIndexTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, template *models.IndexTemplate) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
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

func GetIndexTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) (*types.IndexTemplateItem, fwdiags.Diagnostics) {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	res, err := typedClient.Indices.GetIndexTemplate().Name(templateName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}
	if len(res.IndexTemplates) != 1 {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(
				"Wrong number of templates returned",
				fmt.Sprintf("Elasticsearch API returned %d when requested '%s' template.", len(res.IndexTemplates), templateName),
			),
		}
	}
	tpl := res.IndexTemplates[0]
	return &tpl, nil
}

func DeleteIndexTemplate(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, templateName string) fwdiags.Diagnostics {
	typedClient, err := apiClient.GetESClient()
	if err != nil {
		return diagutil.FrameworkDiagFromError(err)
	}
	_, err = typedClient.Indices.DeleteIndexTemplate(templateName).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return diagutil.FrameworkDiagFromError(err)
	}
	return nil
}
