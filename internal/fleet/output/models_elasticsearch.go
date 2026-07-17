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

package output

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func (model *outputModel) fromAPIElasticsearchModel(ctx context.Context, data *kbapi.KibanaHTTPAPIsOutputResponseElasticsearch) diag.Diagnostics {
	return model.fromAPISimpleOutput(ctx, commonOutputReadData{
		id:                   data.Id,
		name:                 data.Name,
		outputType:           string(data.Type),
		hosts:                data.Hosts,
		caSha256:             data.CaSha256,
		caTrustedFingerprint: data.CaTrustedFingerprint,
		isDefault:            data.IsDefault,
		isDefaultMonitoring:  data.IsDefaultMonitoring,
		configYaml:           data.ConfigYaml,
		ssl:                  data.Ssl,
	})
}

func (model outputModel) toAPICreateElasticsearchModel(ctx context.Context) (kbapi.NewOutputUnion, diag.Diagnostics) {
	return model.toAPICreateSimpleOutput(ctx, func(f commonNewOutputBody) (kbapi.NewOutputUnion, error) {
		body := kbapi.KibanaHTTPAPIsNewOutputElasticsearch{
			Type:                 kbapi.KibanaHTTPAPIsNewOutputElasticsearchTypeElasticsearch,
			CaSha256:             f.CaSha256,
			CaTrustedFingerprint: f.CaTrustedFingerprint,
			ConfigYaml:           f.ConfigYaml,
			Hosts:                f.Hosts,
			Id:                   f.ID,
			IsDefault:            f.IsDefault,
			IsDefaultMonitoring:  f.IsDefaultMonitoring,
			Name:                 f.Name,
			Ssl:                  f.Ssl,
		}
		var union kbapi.NewOutputUnion
		return union, union.FromKibanaHTTPAPIsNewOutputElasticsearch(body)
	})
}

func (model outputModel) toAPIUpdateElasticsearchModel(ctx context.Context) (kbapi.UpdateOutputUnion, diag.Diagnostics) {
	return model.toAPIUpdateSimpleOutput(ctx, func(f commonUpdateOutputBody) (kbapi.UpdateOutputUnion, error) {
		outputType := kbapi.Elasticsearch
		body := kbapi.KibanaHTTPAPIsUpdateOutputElasticsearch{
			Type:                 &outputType,
			CaSha256:             f.CaSha256,
			CaTrustedFingerprint: f.CaTrustedFingerprint,
			ConfigYaml:           f.ConfigYaml,
			Hosts:                f.Hosts,
			IsDefault:            f.IsDefault,
			IsDefaultMonitoring:  f.IsDefaultMonitoring,
			Name:                 f.Name,
			Ssl:                  f.Ssl,
		}
		var union kbapi.UpdateOutputUnion
		return union, union.FromKibanaHTTPAPIsUpdateOutputElasticsearch(body)
	})
}
