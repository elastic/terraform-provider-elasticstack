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

package provider

import (
	"context"
	"slices"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestPluginFrameworkResourcesIncludeIndexSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := &Provider{version: AccTestVersion}

	var registeredTypeNames []string
	for _, newResource := range p.Resources(ctx) {
		r := newResource()
		var res fwresource.MetadataResponse
		r.Metadata(ctx, fwresource.MetadataRequest{ProviderTypeName: "elasticstack"}, &res)
		registeredTypeNames = append(registeredTypeNames, res.TypeName)
	}

	if !slices.Contains(registeredTypeNames, "elasticstack_elasticsearch_index_settings") {
		t.Fatal(`provider does not register "elasticstack_elasticsearch_index_settings"`)
	}
}
