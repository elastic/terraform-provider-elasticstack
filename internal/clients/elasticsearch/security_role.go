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
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
)

func PutRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, name string, role *types.Role) fwdiag.Diagnostics {
	typedClient, d := apiClient.GetESClientDiag()
	if d.HasError() {
		return d
	}

	req := typedClient.Security.PutRole(name)

	req.Applications(role.Applications...)
	req.Cluster(role.Cluster...)
	if role.Description != nil {
		req.Description(*role.Description)
	}
	if role.Global != nil {
		globalJSON, err := json.Marshal(role.Global)
		if err != nil {
			return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to marshal global privileges", err.Error())}
		}
		var global map[string]json.RawMessage
		if err := json.Unmarshal(globalJSON, &global); err != nil {
			return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to convert global privileges", err.Error())}
		}
		req.Global(global)
	}
	req.Indices(role.Indices...)
	req.Metadata(role.Metadata)
	req.RemoteIndices(role.RemoteIndices...)
	req.RunAs(role.RunAs...)

	_, err := req.Do(ctx)
	if err != nil {
		return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to create or update a role", err.Error())}
	}

	return nil
}

func GetRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rolename string) (*types.Role, fwdiag.Diagnostics) {
	typedClient, d := apiClient.GetESClientDiag()
	if d.HasError() {
		return nil, d
	}

	res, err := typedClient.Security.GetRole().Name(rolename).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil, nil
		}
		return nil, diagutil.FrameworkDiagFromError(err)
	}

	if role, ok := res[rolename]; ok {
		return &role, nil
	}
	return nil, fwdiag.Diagnostics{
		fwdiag.NewErrorDiagnostic(
			"Unable to find a role in the cluster",
			fmt.Sprintf(`Unable to find "%s" role in the cluster`, rolename),
		),
	}
}

func DeleteRole(ctx context.Context, apiClient *clients.ElasticsearchScopedClient, rolename string) fwdiag.Diagnostics {
	typedClient, d := apiClient.GetESClientDiag()
	if d.HasError() {
		return d
	}

	_, err := typedClient.Security.DeleteRole(rolename).Do(ctx)
	if err != nil {
		if IsNotFoundElasticsearchError(err) {
			return nil
		}
		return fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic("Unable to delete a role", err.Error())}
	}

	return nil
}
