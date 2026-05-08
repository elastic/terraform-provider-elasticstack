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

package alertingrule

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/require"
)

func TestPersistArtifactsChecksum(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	guidePath := filepath.Join(tempDir, "guide.md")
	content := []byte("hello artifacts\n")
	require.NoError(t, os.WriteFile(guidePath, content, 0o600))

	igObj, diags := types.ObjectValueFrom(ctx, getInvestigationGuideAttrTypes(), investigationGuideModel{
		Content:     types.StringNull(),
		ContentPath: types.StringValue(guidePath),
		Checksum:    types.StringUnknown(),
	})
	require.False(t, diags.HasError())

	artifactsObj, diags := types.ObjectValueFrom(ctx, getArtifactsAttrTypes(), artifactsModel{
		Dashboards:         types.ListNull(types.ObjectType{AttrTypes: getDashboardsAttrTypes()}),
		InvestigationGuide: igObj,
	})
	require.False(t, diags.HasError())

	model := alertingRuleModel{Artifacts: artifactsObj}
	diags = persistArtifactsChecksum(ctx, &model)
	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)

	var am artifactsModel
	diags = model.Artifacts.As(ctx, &am, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	var igm investigationGuideModel
	diags = am.InvestigationGuide.As(ctx, &igm, basetypes.ObjectAsOptions{})
	require.False(t, diags.HasError())

	sum := sha256.Sum256(content)
	require.Equal(t, hex.EncodeToString(sum[:]), igm.Checksum.ValueString())
}
