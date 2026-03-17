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

package parameter

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/kibana/synthetics"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	kibanaClient := synthetics.GetKibanaClient(r, response.Diagnostics)
	if kibanaClient == nil {
		return
	}

	var plan tfModelV0
	diags := request.State.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceID := plan.ID.ValueString()

	compositeID, dg := tryReadCompositeID(resourceID)
	response.Diagnostics.Append(dg...)
	if response.Diagnostics.HasError() {
		return
	}

	if compositeID != nil {
		resourceID = compositeID.ResourceID
	}

	_, err := kibanaClient.KibanaSynthetics.Parameter.Delete(ctx, resourceID)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("Failed to delete parameter `%s`", resourceID), err.Error())
		return
	}
}
