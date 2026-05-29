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

package connector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// REQ-010 runtime telemetry attributes must appear on the data source schema.
var req010RuntimeTelemetryAttrs = []string{
	"status",
	"last_seen",
	"last_synced",
	"last_sync_status",
	"last_indexed_document_count",
	"last_deleted_document_count",
	"last_sync_scheduled_at",
	"last_sync_error",
	"last_access_control_sync_status",
	"last_access_control_sync_error",
	"last_access_control_sync_scheduled_at",
	"last_incremental_sync_scheduled_at",
	"error",
	"filtering",
	"custom_scheduling",
	"configuration",
	"sync_cursor",
	"sync_now",
}

func TestDataSourceSchemaFactory_containsREQ010RuntimeTelemetryAttributes(t *testing.T) {
	t.Parallel()

	schema := dataSourceSchemaFactory(context.Background())
	attrs := schema.GetAttributes()

	for _, name := range req010RuntimeTelemetryAttrs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, ok := attrs[name]
			require.True(t, ok, "data source schema missing REQ-010 attribute %q", name)
		})
	}
}
