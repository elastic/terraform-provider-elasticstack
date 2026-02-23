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

package integrationds

import (
	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type integrationDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Prerelease types.Bool   `tfsdk:"prerelease"`
	Version    types.String `tfsdk:"version"`
}

func (m *integrationDataSourceModel) populateFromAPI(pkgName string, packages []kbapi.PackageListItem) {
	m.Version = types.StringNull()
	for _, pkg := range packages {
		if pkg.Name == pkgName {
			m.Version = types.StringValue(pkg.Version)
			return
		}
	}
}
