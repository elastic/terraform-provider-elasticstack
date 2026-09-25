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

package clients

import (
	"fmt"
	"strings"

	fwdiags "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type CompositeID struct {
	ClusterID  string
	ResourceID string
}

const ServerlessFlavor = "serverless"

// CompositeIDFromStr parses an ID as <cluster_uuid>/<resource_identifier>. Only the first "/"
// separates cluster from resource, so resource_identifier may contain further slashes (for example
// ML calendar events "<calendar_id>/<event_id>" after the cluster segment).
//
// For backward compatibility, an ID with an empty cluster segment and a non-empty resource
// segment (for example "/<synthetics_monitor_id>" from legacy [CompositeID.String] formatting) is
// accepted; an empty resource segment (including a trailing slash after the cluster) is rejected.
func CompositeIDFromStr(id string) (*CompositeID, fwdiags.Diagnostics) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return nil, fwdiags.Diagnostics{
			fwdiags.NewErrorDiagnostic(
				"Wrong resource ID.",
				"Resource ID must have following format: <cluster_uuid>/<resource identifier>",
			),
		}
	}
	if parts[0] == "" {
		return &CompositeID{
			ClusterID:  "",
			ResourceID: parts[1],
		}, nil
	}
	return &CompositeID{
		ClusterID:  parts[0],
		ResourceID: parts[1],
	}, nil
}

func (c *CompositeID) String() string {
	return fmt.Sprintf("%s/%s", c.ClusterID, c.ResourceID)
}

// CompositeIDValue builds a Terraform types.String from a cluster/space ID and
// resource ID, in the same format CompositeIDFromStr expects.
func CompositeIDValue(clusterID, resourceID string) types.String {
	return types.StringValue((&CompositeID{ClusterID: clusterID, ResourceID: resourceID}).String())
}
