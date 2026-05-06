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

package typeutils_test

import (
	"testing"
	"time"

	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestFormatStrictDateTime(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, 3, 15, 10, 30, 45, 123000000, time.UTC)
	got := typeutils.FormatStrictDateTime(ts)
	require.Equal(t, "2024-03-15T10:30:45.123Z", got)
}

func TestTimeToStringValue(t *testing.T) {
	t.Parallel()

	ts := time.Date(2024, 3, 15, 10, 30, 45, 123000000, time.UTC)
	got := typeutils.TimeToStringValue(ts)
	require.Equal(t, types.StringValue("2024-03-15T10:30:45.123Z"), got)
}
