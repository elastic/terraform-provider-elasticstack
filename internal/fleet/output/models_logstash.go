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

func (model *outputModel) fromAPILogstashModel(ctx context.Context, data *kbapi.OutputLogstash) (diags diag.Diagnostics) {
	diags = model.fromAPICommonFields(ctx, commonOutputReadData{
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
	clearRemoteElasticsearchOnlyFields(model)
	return
}

func (model outputModel) toAPICreateLogstashModel(ctx context.Context) (kbapi.NewOutputUnion, diag.Diagnostics) {
	var diags diag.Diagnostics
	f := model.buildCommonNewOutput(ctx, &diags)
	if diags.HasError() {
		return kbapi.NewOutputUnion{}, diags
	}

	body := kbapi.NewOutputLogstash{
		Type:                 kbapi.KibanaHTTPAPIsNewOutputLogstashTypeLogstash,
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
	err := union.FromNewOutputLogstash(body)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.NewOutputUnion{}, diags
	}

	return union, diags
}

func (model outputModel) toAPIUpdateLogstashModel(ctx context.Context) (kbapi.UpdateOutputUnion, diag.Diagnostics) {
	var diags diag.Diagnostics
	f := model.buildCommonUpdateOutput(ctx, &diags)
	if diags.HasError() {
		return kbapi.UpdateOutputUnion{}, diags
	}

	body := kbapi.UpdateOutputLogstash{
		Type: func() *kbapi.KibanaHTTPAPIsUpdateOutputLogstashType {
			outputType := kbapi.Logstash
			return &outputType
		}(),
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
	err := union.FromUpdateOutputLogstash(body)
	if err != nil {
		diags.AddError(err.Error(), "")
		return kbapi.UpdateOutputUnion{}, diags
	}

	return union, diags
}
