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

package calendar

import (
	"testing"

	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	fwtypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPutCalendarRequestFromTFModel(t *testing.T) {
	desc := "hello"
	m := TFModel{
		CalendarID:  fwtypes.StringValue("cal-1"),
		Description: fwtypes.StringValue(desc),
	}
	req := newPutCalendarRequestFromTFModel(m)
	require.NotNil(t, req.Description)
	assert.Equal(t, desc, *req.Description)
}

func TestApplyTypedCalendarToTFModel(t *testing.T) {
	d := "from API"
	c := estypes.Calendar{
		CalendarId:  "cal-1",
		Description: &d,
	}
	var m TFModel
	m.ID = fwtypes.StringValue("cluster/cal-1")
	applyTypedCalendarToTFModel(&m, &c)
	assert.Equal(t, "cal-1", m.CalendarID.ValueString())
	assert.Equal(t, d, m.Description.ValueString())
	assert.Equal(t, "cluster/cal-1", m.ID.ValueString(), "ID should be left to caller/read envelope")
}

func TestApplyTypedCalendarToTFModel_nilDescription(t *testing.T) {
	c := estypes.Calendar{CalendarId: "cal-1"}
	var m TFModel
	applyTypedCalendarToTFModel(&m, &c)
	assert.Empty(t, m.Description.ValueString())
}
