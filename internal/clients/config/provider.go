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

import "github.com/hashicorp/terraform-plugin-framework/types"

type ProviderConfiguration struct {
	Elasticsearch []ElasticsearchConnection `tfsdk:"elasticsearch"`
	Kibana        []KibanaConnection        `tfsdk:"kibana"`
	Fleet         []FleetConnection         `tfsdk:"fleet"`
}

type ElasticsearchConnection struct {
	Username               types.String `tfsdk:"username"`
	Password               types.String `tfsdk:"password"`
	APIKey                 types.String `tfsdk:"api_key"`
	BearerToken            types.String `tfsdk:"bearer_token"`
	ESClientAuthentication types.String `tfsdk:"es_client_authentication"`
	Endpoints              types.List   `tfsdk:"endpoints"`
	Headers                types.Map    `tfsdk:"headers"`
	Insecure               types.Bool   `tfsdk:"insecure"`
	CAFile                 types.String `tfsdk:"ca_file"`
	CAData                 types.String `tfsdk:"ca_data"`
	CertFile               types.String `tfsdk:"cert_file"`
	KeyFile                types.String `tfsdk:"key_file"`
	CertData               types.String `tfsdk:"cert_data"`
	KeyData                types.String `tfsdk:"key_data"`
}

type KibanaConnection struct {
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	APIKey      types.String `tfsdk:"api_key"`
	BearerToken types.String `tfsdk:"bearer_token"`
	Endpoints   types.List   `tfsdk:"endpoints"`
	Insecure    types.Bool   `tfsdk:"insecure"`
	CACerts     types.List   `tfsdk:"ca_certs"`
}

type FleetConnection struct {
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	APIKey      types.String `tfsdk:"api_key"`
	BearerToken types.String `tfsdk:"bearer_token"`
	Endpoint    types.String `tfsdk:"endpoint"`
	Insecure    types.Bool   `tfsdk:"insecure"`
	CACerts     types.List   `tfsdk:"ca_certs"`
}
