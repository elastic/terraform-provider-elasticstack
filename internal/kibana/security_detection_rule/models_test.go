package securitydetectionrule

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/typeutils"
	"github.com/google/uuid"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v2Diag "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/require"
)

type mockAPIClient struct {
	serverVersion *version.Version
	serverFlavor  string
	enforceResult bool
}

func (m mockAPIClient) EnforceMinVersion(_ context.Context, minVersion *version.Version) (bool, v2Diag.Diagnostics) {
	supported := m.serverVersion.GreaterThanOrEqual(minVersion)
	return supported, nil
}

// NewMockAPIClient creates a new mock API client with default values that support response actions
// This can be used in tests where you need to pass a client to functions like toUpdateProps
func NewMockAPIClient() clients.MinVersionEnforceable {
	// Use version 8.16.0 by default to support response actions
	v, _ := version.NewVersion("8.16.0")

	return mockAPIClient{
		serverVersion: v,
		serverFlavor:  "default",
		enforceResult: true,
	}
}

// NewMockAPIClientWithVersion creates a mock API client with a specific version
// Use this when you need to test specific version behavior
func NewMockAPIClientWithVersion(versionStr string) *mockAPIClient {
	v, err := version.NewVersion(versionStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid version in test: %s", versionStr))
	}
	return &mockAPIClient{
		serverVersion: v,
		serverFlavor:  "default",
		enforceResult: true,
	}
}
func TestUpdateFromQueryRule(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name     string
		rule     kbapi.SecurityDetectionsAPIQueryRule
		spaceId  string //nolint:revive // struct field name matches API
		expected Data
	}{
		{
			name:    "complete query rule",
			spaceId: "test-space",
			rule: kbapi.SecurityDetectionsAPIQueryRule{
				Id:             uuid.MustParse("12345678-1234-1234-1234-123456789012"),
				RuleId:         "test-rule-id",
				Name:           "Test Query Rule",
				Type:           "query",
				Query:          "user.name:test",
				Language:       "kuery",
				Enabled:        true,
				From:           "now-6m",
				To:             "now",
				Interval:       "5m",
				Description:    "Test description",
				RiskScore:      75,
				Severity:       "medium",
				MaxSignals:     100,
				Version:        1,
				Author:         []string{"Test Author"},
				Tags:           []string{"test", "detection"},
				Index:          schemautil.Pointer([]string{"logs-*", "metrics-*"}),
				CreatedBy:      "test-user",
				UpdatedBy:      "test-user",
				Revision:       1,
				FalsePositives: []string{"Known false positive"},
				References:     []string{"https://example.com/test"},
				License:        schemautil.Pointer(kbapi.SecurityDetectionsAPIRuleLicense("MIT")),
				Note:           schemautil.Pointer(kbapi.SecurityDetectionsAPIInvestigationGuide("Investigation note")),
				Setup:          "Setup instructions",
			},
			expected: Data{
				ID:             types.StringValue("test-space/12345678-1234-1234-1234-123456789012"),
				SpaceID:        types.StringValue("test-space"),
				RuleID:         types.StringValue("test-rule-id"),
				Name:           types.StringValue("Test Query Rule"),
				Type:           types.StringValue("query"),
				Query:          types.StringValue("user.name:test"),
				Language:       types.StringValue("kuery"),
				Enabled:        types.BoolValue(true),
				From:           types.StringValue("now-6m"),
				To:             types.StringValue("now"),
				Interval:       types.StringValue("5m"),
				Description:    types.StringValue("Test description"),
				RiskScore:      types.Int64Value(75),
				Severity:       types.StringValue("medium"),
				MaxSignals:     types.Int64Value(100),
				Version:        types.Int64Value(1),
				Author:         typeutils.ListValueFrom(ctx, []string{"Test Author"}, types.StringType, path.Root("author"), &diags),
				Tags:           typeutils.ListValueFrom(ctx, []string{"test", "detection"}, types.StringType, path.Root("tags"), &diags),
				Index:          typeutils.ListValueFrom(ctx, []string{"logs-*", "metrics-*"}, types.StringType, path.Root("index"), &diags),
				CreatedBy:      types.StringValue("test-user"),
				UpdatedBy:      types.StringValue("test-user"),
				Revision:       types.Int64Value(1),
				FalsePositives: typeutils.ListValueFrom(ctx, []string{"Known false positive"}, types.StringType, path.Root("false_positives"), &diags),
				References:     typeutils.ListValueFrom(ctx, []string{"https://example.com/test"}, types.StringType, path.Root("references"), &diags),
				License:        types.StringValue("MIT"),
				Note:           types.StringValue("Investigation note"),
				Setup:          types.StringValue("Setup instructions"),
			},
		},
		{
			name:    "minimal query rule",
			spaceId: "default",
			rule: kbapi.SecurityDetectionsAPIQueryRule{
				Id:          uuid.MustParse("87654321-4321-4321-4321-210987654321"),
				RuleId:      "minimal-rule",
				Name:        "Minimal Rule",
				Type:        "query",
				Query:       "*",
				Language:    "kuery",
				Enabled:     false,
				From:        "now-1h",
				To:          "now",
				Interval:    "1m",
				Description: "Minimal test",
				RiskScore:   1,
				Severity:    "low",
				MaxSignals:  50,
				Version:     1,
				CreatedBy:   "system",
				UpdatedBy:   "system",
				Revision:    1,
			},
			expected: Data{
				ID:          types.StringValue("default/87654321-4321-4321-4321-210987654321"),
				SpaceID:     types.StringValue("default"),
				RuleID:      types.StringValue("minimal-rule"),
				Name:        types.StringValue("Minimal Rule"),
				Type:        types.StringValue("query"),
				Query:       types.StringValue("*"),
				Language:    types.StringValue("kuery"),
				Enabled:     types.BoolValue(false),
				From:        types.StringValue("now-1h"),
				To:          types.StringValue("now"),
				Interval:    types.StringValue("1m"),
				Description: types.StringValue("Minimal test"),
				RiskScore:   types.Int64Value(1),
				Severity:    types.StringValue("low"),
				MaxSignals:  types.Int64Value(50),
				Version:     types.Int64Value(1),
				CreatedBy:   types.StringValue("system"),
				UpdatedBy:   types.StringValue("system"),
				Revision:    types.Int64Value(1),
				Author:      types.ListValueMust(types.StringType, []attr.Value{}),
				Tags:        types.ListValueMust(types.StringType, []attr.Value{}),
				Index:       types.ListValueMust(types.StringType, []attr.Value{}),
			},
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Data{
				SpaceID: types.StringValue(tt.spaceId),
			}

			diags := updateFromQueryRule(ctx, &tt.rule, &data)
			require.Empty(t, diags)

			// Compare key fields
			require.Equal(t, tt.expected.ID, data.ID)
			require.Equal(t, tt.expected.RuleID, data.RuleID)
			require.Equal(t, tt.expected.Name, data.Name)
			require.Equal(t, tt.expected.Type, data.Type)
			require.Equal(t, tt.expected.Query, data.Query)
			require.Equal(t, tt.expected.Language, data.Language)
			require.Equal(t, tt.expected.Enabled, data.Enabled)
			require.Equal(t, tt.expected.RiskScore, data.RiskScore)
			require.Equal(t, tt.expected.Severity, data.Severity)

			// Verify list fields have correct length
			require.Len(t, data.Author.Elements(), len(tt.expected.Author.Elements()))
			require.Len(t, data.Tags.Elements(), len(tt.expected.Tags.Elements()))
			require.Len(t, data.Index.Elements(), len(tt.expected.Index.Elements()))
		})
	}
}

func TestToQueryRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name               string
		data               Data
		expectedName       string
		expectedType       string
		expectedQuery      string
		expectedRiskScore  int64
		expectedSeverity   string
		shouldHaveLanguage bool
		shouldHaveIndex    bool
		shouldHaveActions  bool
		shouldHaveRuleID   bool
		shouldError        bool
	}{
		{
			name: "complete query rule create",
			data: Data{
				Name:        types.StringValue("Test Create Rule"),
				Type:        types.StringValue("query"),
				Query:       types.StringValue("process.name:malicious"),
				Language:    types.StringValue("kuery"),
				RiskScore:   types.Int64Value(85),
				Severity:    types.StringValue("high"),
				Description: types.StringValue("Test rule description"),
				Index:       typeutils.ListValueFrom(ctx, []string{"winlogbeat-*"}, types.StringType, path.Root("index"), &diags),
				Author:      typeutils.ListValueFrom(ctx, []string{"Security Team"}, types.StringType, path.Root("author"), &diags),
				Enabled:     types.BoolValue(true),
				RuleID:      types.StringValue("custom-rule-id"),
			},
			expectedName:       "Test Create Rule",
			expectedType:       "query",
			expectedQuery:      "process.name:malicious",
			expectedRiskScore:  85,
			expectedSeverity:   "high",
			shouldHaveLanguage: true,
			shouldHaveIndex:    true,
			shouldHaveRuleID:   true,
		},
		{
			name: "minimal query rule create",
			data: Data{
				Name:        types.StringValue("Minimal Rule"),
				Type:        types.StringValue("query"),
				Query:       types.StringValue("*"),
				RiskScore:   types.Int64Value(1),
				Severity:    types.StringValue("low"),
				Description: types.StringValue("Minimal description"),
			},
			expectedName:      "Minimal Rule",
			expectedType:      "query",
			expectedQuery:     "*",
			expectedRiskScore: 1,
			expectedSeverity:  "low",
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createProps, createDiags := toQueryRuleCreateProps(ctx, NewMockAPIClient(), tt.data)

			if tt.shouldError {
				require.NotEmpty(t, createDiags)
				return
			}

			require.Empty(t, createDiags)

			// Extract the concrete type from the union
			queryRule, err := createProps.AsSecurityDetectionsAPIQueryRuleCreateProps()
			require.NoError(t, err)

			require.Equal(t, tt.expectedName, queryRule.Name)
			require.Equal(t, tt.expectedType, string(queryRule.Type))
			require.NotNil(t, queryRule.Query)
			require.Equal(t, tt.expectedQuery, *queryRule.Query)
			require.Equal(t, tt.expectedRiskScore, int64(queryRule.RiskScore))
			require.Equal(t, tt.expectedSeverity, string(queryRule.Severity))

			if tt.shouldHaveLanguage {
				require.NotNil(t, queryRule.Language)
			}

			if tt.shouldHaveIndex {
				require.NotNil(t, queryRule.Index)
				require.NotEmpty(t, *queryRule.Index)
			}

			if tt.shouldHaveRuleID {
				require.NotNil(t, queryRule.RuleId)
				require.Equal(t, "custom-rule-id", *queryRule.RuleId)
			}
		})
	}
}

func TestToEqlRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := Data{
		Name:            types.StringValue("EQL Test Rule"),
		Type:            types.StringValue("eql"),
		Query:           types.StringValue("process where process.name == \"cmd.exe\""),
		RiskScore:       types.Int64Value(60),
		Severity:        types.StringValue("medium"),
		Description:     types.StringValue("EQL rule description"),
		TiebreakerField: types.StringValue("@timestamp"),
	}

	createProps, createDiags := toEqlRuleCreateProps(ctx, NewMockAPIClient(), data)
	require.Empty(t, createDiags)

	eqlRule, err := createProps.AsSecurityDetectionsAPIEqlRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "EQL Test Rule", eqlRule.Name)
	require.Equal(t, "eql", string(eqlRule.Type))
	require.Equal(t, "process where process.name == \"cmd.exe\"", eqlRule.Query)
	require.Equal(t, "eql", string(eqlRule.Language))
	require.Equal(t, int64(60), int64(eqlRule.RiskScore))
	require.Equal(t, "medium", string(eqlRule.Severity))

	require.NotNil(t, eqlRule.TiebreakerField)
	require.Equal(t, "@timestamp", *eqlRule.TiebreakerField)

	require.Empty(t, diags)
}

func TestToMachineLearningRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name               string
		data               Data
		expectedJobCount   int
		shouldHaveSingle   bool
		shouldHaveMultiple bool
	}{
		{
			name: "single ML job",
			data: Data{
				Name:                 types.StringValue("ML Test Rule"),
				Type:                 types.StringValue("machine_learning"),
				RiskScore:            types.Int64Value(70),
				Severity:             types.StringValue("high"),
				Description:          types.StringValue("ML rule description"),
				AnomalyThreshold:     types.Int64Value(50),
				MachineLearningJobID: typeutils.ListValueFrom(ctx, []string{"suspicious_activity"}, types.StringType, path.Root("machine_learning_job_id"), &diags),
			},
			expectedJobCount:   1,
			shouldHaveMultiple: true,
		},
		{
			name: "multiple ML jobs",
			data: Data{
				Name:                 types.StringValue("ML Multi Job Rule"),
				Type:                 types.StringValue("machine_learning"),
				RiskScore:            types.Int64Value(80),
				Severity:             types.StringValue("critical"),
				Description:          types.StringValue("ML multi job rule"),
				AnomalyThreshold:     types.Int64Value(75),
				MachineLearningJobID: typeutils.ListValueFrom(ctx, []string{"job1", "job2", "job3"}, types.StringType, path.Root("machine_learning_job_id"), &diags),
			},
			expectedJobCount:   3,
			shouldHaveMultiple: true,
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createProps, createDiags := tt.data.toMachineLearningRuleCreateProps(ctx, NewMockAPIClient())
			require.Empty(t, createDiags)

			mlRule, err := createProps.AsSecurityDetectionsAPIMachineLearningRuleCreateProps()
			require.NoError(t, err)

			require.Equal(t, tt.data.Name.ValueString(), mlRule.Name)
			require.Equal(t, "machine_learning", string(mlRule.Type))
			require.Equal(t, tt.data.AnomalyThreshold.ValueInt64(), int64(mlRule.AnomalyThreshold))

			if tt.shouldHaveSingle {
				singleJobID, err := mlRule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId0()
				require.NoError(t, err)
				require.Equal(t, "suspicious_activity", singleJobID)
			}

			if tt.shouldHaveMultiple {
				multipleJobIDs, err := mlRule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId1()
				require.NoError(t, err)
				require.Len(t, multipleJobIDs, tt.expectedJobCount)
			}
		})
	}
}

func TestToEsqlRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := Data{
		Type:        types.StringValue("esql"),
		Name:        types.StringValue("Test ESQL Rule"),
		Description: types.StringValue("Test ESQL rule description"),
		Query:       types.StringValue("FROM logs | WHERE user.name == \"suspicious_user\""),
		RiskScore:   types.Int64Value(85),
		Severity:    types.StringValue("high"),
		Enabled:     types.BoolValue(true),
		From:        types.StringValue("now-1h"),
		To:          types.StringValue("now"),
		Interval:    types.StringValue("10m"),
		Author:      typeutils.ListValueFrom(ctx, []string{"Security Team"}, types.StringType, path.Root("author"), &diags),
		Tags:        typeutils.ListValueFrom(ctx, []string{"esql", "test"}, types.StringType, path.Root("tags"), &diags),
	}

	require.Empty(t, diags)

	createProps, createDiags := data.toEsqlRuleCreateProps(ctx, NewMockAPIClient())
	require.Empty(t, createDiags)

	esqlRule, err := createProps.AsSecurityDetectionsAPIEsqlRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "Test ESQL Rule", esqlRule.Name)
	require.Equal(t, "Test ESQL rule description", esqlRule.Description)
	require.Equal(t, "esql", string(esqlRule.Type))
	require.Equal(t, "FROM logs | WHERE user.name == \"suspicious_user\"", esqlRule.Query)
	require.Equal(t, "esql", string(esqlRule.Language))
	require.Equal(t, int64(85), int64(esqlRule.RiskScore))
	require.Equal(t, "high", string(esqlRule.Severity))
}

func TestToNewTermsRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := Data{
		Type:               types.StringValue("new_terms"),
		Name:               types.StringValue("Test New Terms Rule"),
		Description:        types.StringValue("Test new terms rule description"),
		Query:              types.StringValue("user.name:*"),
		Language:           types.StringValue("kuery"),
		NewTermsFields:     typeutils.ListValueFrom(ctx, []string{"user.name", "host.name"}, types.StringType, path.Root("new_terms_fields"), &diags),
		HistoryWindowStart: types.StringValue("now-7d"),
		RiskScore:          types.Int64Value(60),
		Severity:           types.StringValue("medium"),
		Enabled:            types.BoolValue(true),
		From:               types.StringValue("now-6m"),
		To:                 types.StringValue("now"),
		Interval:           types.StringValue("5m"),
		Index:              typeutils.ListValueFrom(ctx, []string{"logs-*"}, types.StringType, path.Root("index"), &diags),
	}

	require.Empty(t, diags)

	createProps, createDiags := data.toNewTermsRuleCreateProps(ctx, NewMockAPIClient())
	require.Empty(t, createDiags)

	newTermsRule, err := createProps.AsSecurityDetectionsAPINewTermsRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "Test New Terms Rule", newTermsRule.Name)
	require.Equal(t, "Test new terms rule description", newTermsRule.Description)
	require.Equal(t, "new_terms", string(newTermsRule.Type))
	require.Equal(t, "user.name:*", newTermsRule.Query)
	require.Equal(t, "now-7d", newTermsRule.HistoryWindowStart)
	require.Equal(t, int64(60), int64(newTermsRule.RiskScore))
	require.Equal(t, "medium", string(newTermsRule.Severity))
	require.Len(t, newTermsRule.NewTermsFields, 2)
	require.Contains(t, newTermsRule.NewTermsFields, "user.name")
	require.Contains(t, newTermsRule.NewTermsFields, "host.name")
}

func TestToSavedQueryRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := Data{
		Type:        types.StringValue("saved_query"),
		Name:        types.StringValue("Test Saved Query Rule"),
		Description: types.StringValue("Test saved query rule description"),
		SavedID:     types.StringValue("my-saved-query-id"),
		RiskScore:   types.Int64Value(70),
		Severity:    types.StringValue("high"),
		Enabled:     types.BoolValue(true),
		From:        types.StringValue("now-30m"),
		To:          types.StringValue("now"),
		Interval:    types.StringValue("15m"),
		Index:       typeutils.ListValueFrom(ctx, []string{"auditbeat-*", "filebeat-*"}, types.StringType, path.Root("index"), &diags),
		Author:      typeutils.ListValueFrom(ctx, []string{"Security Team"}, types.StringType, path.Root("author"), &diags),
		Tags:        typeutils.ListValueFrom(ctx, []string{"saved-query", "detection"}, types.StringType, path.Root("tags"), &diags),
	}

	require.Empty(t, diags)

	createProps, createDiags := data.toSavedQueryRuleCreateProps(ctx, NewMockAPIClient())
	require.Empty(t, createDiags)

	savedQueryRule, err := createProps.AsSecurityDetectionsAPISavedQueryRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "Test Saved Query Rule", savedQueryRule.Name)
	require.Equal(t, "Test saved query rule description", savedQueryRule.Description)
	require.Equal(t, "saved_query", string(savedQueryRule.Type))
	require.Equal(t, "my-saved-query-id", savedQueryRule.SavedId)
	require.Equal(t, int64(70), int64(savedQueryRule.RiskScore))
	require.Equal(t, "high", string(savedQueryRule.Severity))
}

func TestToThreatMatchRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := Data{
		Type:        types.StringValue("threat_match"),
		Name:        types.StringValue("Test Threat Match Rule"),
		Description: types.StringValue("Test threat match rule description"),
		Query:       types.StringValue("source.ip:*"),
		Language:    types.StringValue("kuery"),
		ThreatIndex: typeutils.ListValueFrom(ctx, []string{"threat-intel-*"}, types.StringType, path.Root("threat_index"), &diags),
		ThreatMapping: typeutils.ListValueFrom(ctx, []TfDataItem{
			{
				Entries: typeutils.ListValueFrom(ctx, []TfDataItemEntry{
					{
						Field: types.StringValue("source.ip"),
						Type:  types.StringValue("mapping"),
						Value: types.StringValue("threat.indicator.ip"),
					},
				}, getThreatMappingEntryElementType(), path.Root("threat_mapping").AtListIndex(0).AtName("entries"), &diags),
			},
		}, getThreatMappingElementType(), path.Root("threat_mapping"), &diags),
		RiskScore: types.Int64Value(90),
		Severity:  types.StringValue("critical"),
		Enabled:   types.BoolValue(true),
		From:      types.StringValue("now-1h"),
		To:        types.StringValue("now"),
		Interval:  types.StringValue("5m"),
		Index:     typeutils.ListValueFrom(ctx, []string{"logs-*"}, types.StringType, path.Root("index"), &diags),
	}

	require.Empty(t, diags)

	createProps, createDiags := data.toThreatMatchRuleCreateProps(ctx, NewMockAPIClient())
	require.Empty(t, createDiags)

	threatMatchRule, err := createProps.AsSecurityDetectionsAPIThreatMatchRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "Test Threat Match Rule", threatMatchRule.Name)
	require.Equal(t, "Test threat match rule description", threatMatchRule.Description)
	require.Equal(t, "threat_match", string(threatMatchRule.Type))
	require.Equal(t, "source.ip:*", threatMatchRule.Query)
	require.Equal(t, int64(90), int64(threatMatchRule.RiskScore))
	require.Equal(t, "critical", string(threatMatchRule.Severity))
	require.Len(t, threatMatchRule.ThreatIndex, 1)
	require.Equal(t, "threat-intel-*", threatMatchRule.ThreatIndex[0])
	require.Len(t, threatMatchRule.ThreatMapping, 1)
}

func TestToThresholdRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := Data{
		Type:        types.StringValue("threshold"),
		Name:        types.StringValue("Test Threshold Rule"),
		Description: types.StringValue("Test threshold rule description"),
		Query:       types.StringValue("event.action:login"),
		Language:    types.StringValue("kuery"),
		Threshold: typeutils.ObjectValueFrom(ctx, &ThresholdModel{
			Field:       typeutils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("threshold").AtName("field"), &diags),
			Value:       types.Int64Value(5),
			Cardinality: types.ListNull(getCardinalityType()),
		}, getThresholdType(), path.Root("threshold"), &diags),
		RiskScore: types.Int64Value(80),
		Severity:  types.StringValue("high"),
		Enabled:   types.BoolValue(true),
		From:      types.StringValue("now-1h"),
		To:        types.StringValue("now"),
		Interval:  types.StringValue("5m"),
		Index:     typeutils.ListValueFrom(ctx, []string{"auditbeat-*"}, types.StringType, path.Root("index"), &diags),
	}

	require.Empty(t, diags)

	createProps, createDiags := data.toThresholdRuleCreateProps(ctx, NewMockAPIClient())
	require.Empty(t, createDiags)

	thresholdRule, err := createProps.AsSecurityDetectionsAPIThresholdRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "Test Threshold Rule", thresholdRule.Name)
	require.Equal(t, "Test threshold rule description", thresholdRule.Description)
	require.Equal(t, "threshold", string(thresholdRule.Type))
	require.Equal(t, "event.action:login", thresholdRule.Query)
	require.Equal(t, int64(80), int64(thresholdRule.RiskScore))
	require.Equal(t, "high", string(thresholdRule.Severity))

	// Verify threshold configuration
	require.NotNil(t, thresholdRule.Threshold)
	require.Equal(t, int64(5), int64(thresholdRule.Threshold.Value))

	// Check single field
	singleField, err := thresholdRule.Threshold.Field.AsSecurityDetectionsAPIThresholdField0()
	require.NoError(t, err)
	require.Equal(t, "user.name", singleField)
}

func TestThresholdToAPI(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name               string
		data               Data
		expectedValue      int64
		expectedFieldCount int
		hasCardinality     bool
	}{
		{
			name: "threshold with single field",
			data: Data{
				Threshold: typeutils.ObjectValueFrom(ctx, &ThresholdModel{
					Field:       typeutils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("threshold").AtName("field"), &diags),
					Value:       types.Int64Value(10),
					Cardinality: types.ListNull(getCardinalityType()),
				}, getThresholdType(), path.Root("threshold"), &diags),
			},
			expectedValue:      10,
			expectedFieldCount: 1,
		},
		{
			name: "threshold with multiple fields and cardinality",
			data: Data{
				Threshold: typeutils.ObjectValueFrom(ctx, &ThresholdModel{
					Field: typeutils.ListValueFrom(ctx, []string{"user.name", "source.ip"}, types.StringType, path.Root("threshold").AtName("field"), &diags),
					Value: types.Int64Value(5),
					Cardinality: typeutils.ListValueFrom(ctx, []CardinalityModel{
						{
							Field: types.StringValue("destination.ip"),
							Value: types.Int64Value(2),
						},
					}, getCardinalityType(), path.Root("threshold").AtName("cardinality"), &diags),
				}, getThresholdType(), path.Root("threshold"), &diags),
			},
			expectedValue:      5,
			expectedFieldCount: 2,
			hasCardinality:     true,
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			threshold := tt.data.thresholdToAPI(ctx, &diags)
			require.Empty(t, diags)
			require.NotNil(t, threshold)

			require.Equal(t, tt.expectedValue, int64(threshold.Value))

			// Check field count
			if singleField, err := threshold.Field.AsSecurityDetectionsAPIThresholdField0(); err == nil {
				require.Equal(t, 1, tt.expectedFieldCount)
				require.NotEmpty(t, singleField)
			} else if multipleFields, err := threshold.Field.AsSecurityDetectionsAPIThresholdField1(); err == nil {
				require.Len(t, multipleFields, tt.expectedFieldCount)
			}

			if tt.hasCardinality {
				require.NotNil(t, threshold.Cardinality)
				require.NotEmpty(t, *threshold.Cardinality)
			}
		})
	}
}

func TestAlertSuppressionToAPI(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name                     string
		data                     Data
		expectedGroupByCount     int
		hasDuration              bool
		hasMissingFieldsStrategy bool
	}{
		{
			name: "alert suppression with all fields",
			data: Data{
				AlertSuppression: typeutils.ObjectValueFrom(ctx, &AlertSuppressionModel{
					GroupBy:               typeutils.ListValueFrom(ctx, []string{"user.name", "source.ip"}, types.StringType, path.Root("alert_suppression").AtName("group_by"), &diags),
					Duration:              customtypes.NewDurationValue("10m"),
					MissingFieldsStrategy: types.StringValue("suppress"),
				}, getAlertSuppressionType(), path.Root("alert_suppression"), &diags),
			},
			expectedGroupByCount:     2,
			hasDuration:              true,
			hasMissingFieldsStrategy: true,
		},
		{
			name: "alert suppression minimal",
			data: Data{
				AlertSuppression: typeutils.ObjectValueFrom(ctx, &AlertSuppressionModel{
					GroupBy:               typeutils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("alert_suppression").AtName("group_by"), &diags),
					Duration:              customtypes.NewDurationNull(),
					MissingFieldsStrategy: types.StringNull(),
				}, getAlertSuppressionType(), path.Root("alert_suppression"), &diags),
			},
			expectedGroupByCount: 1,
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alertSuppression := tt.data.alertSuppressionToAPI(ctx, &diags)
			require.Empty(t, diags)
			require.NotNil(t, alertSuppression)

			require.Len(t, alertSuppression.GroupBy, tt.expectedGroupByCount)

			if tt.hasDuration {
				require.NotNil(t, alertSuppression.Duration)
				require.Equal(t, 10, alertSuppression.Duration.Value)
				require.Equal(t, "m", string(alertSuppression.Duration.Unit))
			}

			if tt.hasMissingFieldsStrategy {
				require.NotNil(t, alertSuppression.MissingFieldsStrategy)
				require.Equal(t, "suppress", string(*alertSuppression.MissingFieldsStrategy))
			}
		})
	}
}

func TestThreatMappingToAPI(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := Data{
		ThreatMapping: typeutils.ListValueFrom(ctx, []TfDataItem{
			{
				Entries: typeutils.ListValueFrom(ctx, []TfDataItemEntry{
					{
						Field: types.StringValue("source.ip"),
						Type:  types.StringValue("mapping"),
						Value: types.StringValue("threat.indicator.ip"),
					},
					{
						Field: types.StringValue("user.name"),
						Type:  types.StringValue("mapping"),
						Value: types.StringValue("threat.indicator.user.name"),
					},
				}, getThreatMappingEntryElementType(), path.Root("threat_mapping").AtListIndex(0).AtName("entries"), &diags),
			},
		}, getThreatMappingElementType(), path.Root("threat_mapping"), &diags),
	}

	require.Empty(t, diags)

	threatMapping, threatMappingDiags := data.threatMappingToAPI(ctx)
	require.Empty(t, threatMappingDiags)
	require.NotNil(t, threatMapping)
	require.Len(t, threatMapping, 1)

	mapping := threatMapping[0]
	require.Len(t, mapping.Entries, 2)

	require.Equal(t, "source.ip", mapping.Entries[0].Field)
	require.Equal(t, "mapping", string(mapping.Entries[0].Type))
	require.Equal(t, "threat.indicator.ip", mapping.Entries[0].Value)

	require.Equal(t, "user.name", mapping.Entries[1].Field)
	require.Equal(t, "mapping", string(mapping.Entries[1].Type))
	require.Equal(t, "threat.indicator.user.name", mapping.Entries[1].Value)
}

func TestActionsToAPI(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := Data{
		Actions: typeutils.ListValueFrom(ctx, []ActionModel{
			{
				ActionTypeID: types.StringValue(".slack"),
				ID:           types.StringValue("slack-action-1"),
				Params: typeutils.MapValueFrom(ctx, map[string]attr.Value{
					"message": types.StringValue("Alert triggered"),
					"channel": types.StringValue("#security"),
				}, types.StringType, path.Root("actions").AtListIndex(0).AtName("params"), &diags),
				Group: types.StringValue("default"),
				UUID:  types.StringNull(),
				AlertsFilter: typeutils.MapValueFrom(ctx, map[string]attr.Value{
					"status":   types.StringValue("open"),
					"severity": types.StringValue("high"),
				}, types.StringType, path.Root("actions").AtListIndex(0).AtName("alerts_filter"), &diags),
				Frequency: typeutils.ObjectValueFrom(ctx, &ActionFrequencyModel{
					NotifyWhen: types.StringValue("onActionGroupChange"),
					Summary:    types.BoolValue(false),
					Throttle:   types.StringValue("1h"),
				}, getActionFrequencyType(), path.Root("actions").AtListIndex(0).AtName("frequency"), &diags),
			},
		}, getActionElementType(), path.Root("actions"), &diags),
	}

	require.Empty(t, diags)

	actions, actionsDiags := data.actionsToAPI(ctx)
	require.Empty(t, actionsDiags)
	require.Len(t, actions, 1)

	action := actions[0]
	require.Equal(t, ".slack", action.ActionTypeId)
	require.Equal(t, "slack-action-1", action.Id)
	require.NotNil(t, action.Params)
	require.Contains(t, action.Params, "message")
	require.Equal(t, "Alert triggered", action.Params["message"])
	require.NotNil(t, action.Group)
	require.Equal(t, "default", *action.Group)
	require.NotNil(t, action.Frequency)
}

func TestFiltersToAPI(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	filtersJSON := `[{"query": {"match": {"field": "value"}}}, {"range": {"timestamp": {"gte": "now-1h"}}}]`

	data := Data{
		Filters: jsontypes.NewNormalizedValue(filtersJSON),
	}

	// Test filters conversion
	filters, filtersDiags := data.filtersToAPI(ctx)
	require.Empty(t, filtersDiags)
	require.NotNil(t, filters)
	require.Len(t, *filters, 2)

	require.Empty(t, diags)
}

func TestConvertActionsToModel(t *testing.T) {
	ctx := context.Background()

	apiActions := []kbapi.SecurityDetectionsAPIRuleAction{
		{
			ActionTypeId: ".email",
			Id:           "email-action-1",
			Params: kbapi.SecurityDetectionsAPIRuleActionParams{
				"to":      []string{"admin@example.com"},
				"subject": "Security Alert",
				"message": "Alert details here",
			},
			Group: schemautil.Pointer(kbapi.SecurityDetectionsAPIRuleActionGroup("default")),
			Uuid:  schemautil.Pointer(kbapi.SecurityDetectionsAPINonEmptyString("action-uuid-123")),
		},
	}

	actionsList, diags := convertActionsToModel(ctx, apiActions)
	require.Empty(t, diags)
	require.False(t, actionsList.IsNull())

	var actions []ActionModel
	elemDiags := actionsList.ElementsAs(ctx, &actions, false)
	require.Empty(t, elemDiags)
	require.Len(t, actions, 1)

	action := actions[0]
	require.Equal(t, ".email", action.ActionTypeID.ValueString())
	require.Equal(t, "email-action-1", action.ID.ValueString())
	require.Equal(t, "default", action.Group.ValueString())
	require.Equal(t, "action-uuid-123", action.UUID.ValueString())
}

func TestUpdateFromRule_UnsupportedType(t *testing.T) {
	ctx := context.Background()
	data := &Data{}

	// Create a mock response that will fail to determine discriminator
	response := &kbapi.SecurityDetectionsAPIRuleResponse{}

	diags := data.updateFromRule(ctx, response)
	require.NotEmpty(t, diags)
	require.True(t, diags.HasError())
}

func TestUpdateFromRule(t *testing.T) {
	ctx := context.Background()
	testUUID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	spaceID := "test-space"

	tests := []struct {
		name         string
		setupRule    func() *kbapi.SecurityDetectionsAPIRuleResponse
		expectError  bool
		errorMessage string
		validateData func(t *testing.T, data *Data)
	}{
		{
			name: "query rule type",
			setupRule: func() *kbapi.SecurityDetectionsAPIRuleResponse {
				rule := kbapi.SecurityDetectionsAPIQueryRule{
					Id:          testUUID,
					RuleId:      "test-query-rule",
					Name:        "Test Query Rule",
					Type:        "query",
					Query:       "user.name:test",
					Language:    "kuery",
					Enabled:     true,
					RiskScore:   75,
					Severity:    "medium",
					Version:     1,
					Description: "Test query rule description",
					From:        "now-6m",
					To:          "now",
					Interval:    "5m",
					CreatedBy:   "test-user",
					UpdatedBy:   "test-user",
					Revision:    1,
				}
				response := &kbapi.SecurityDetectionsAPIRuleResponse{}
				err := response.FromSecurityDetectionsAPIQueryRule(rule)
				require.NoError(t, err)
				return response
			},
			validateData: func(t *testing.T, data *Data) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceID, testUUID.String()), data.ID.ValueString())
				require.Equal(t, "test-query-rule", data.RuleID.ValueString())
				require.Equal(t, "Test Query Rule", data.Name.ValueString())
				require.Equal(t, "query", data.Type.ValueString())
				require.Equal(t, "user.name:test", data.Query.ValueString())
				require.Equal(t, "kuery", data.Language.ValueString())
				require.True(t, data.Enabled.ValueBool())
				require.Equal(t, int64(75), data.RiskScore.ValueInt64())
				require.Equal(t, "medium", data.Severity.ValueString())
			},
		},
		{
			name: "eql rule type",
			setupRule: func() *kbapi.SecurityDetectionsAPIRuleResponse {
				rule := kbapi.SecurityDetectionsAPIEqlRule{
					Id:          testUUID,
					RuleId:      "test-eql-rule",
					Name:        "Test EQL Rule",
					Type:        "eql",
					Query:       "process where process.name == \"cmd.exe\"",
					Language:    "eql",
					Enabled:     true,
					RiskScore:   80,
					Severity:    "high",
					Version:     1,
					Description: "Test EQL rule description",
					From:        "now-6m",
					To:          "now",
					Interval:    "5m",
					CreatedBy:   "test-user",
					UpdatedBy:   "test-user",
					Revision:    1,
				}
				response := &kbapi.SecurityDetectionsAPIRuleResponse{}
				err := response.FromSecurityDetectionsAPIEqlRule(rule)
				require.NoError(t, err)
				return response
			},
			validateData: func(t *testing.T, data *Data) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceID, testUUID.String()), data.ID.ValueString())
				require.Equal(t, "test-eql-rule", data.RuleID.ValueString())
				require.Equal(t, "Test EQL Rule", data.Name.ValueString())
				require.Equal(t, "eql", data.Type.ValueString())
				require.Equal(t, "process where process.name == \"cmd.exe\"", data.Query.ValueString())
				require.Equal(t, "eql", data.Language.ValueString())
				require.Equal(t, int64(80), data.RiskScore.ValueInt64())
				require.Equal(t, "high", data.Severity.ValueString())
			},
		},
		{
			name: "esql rule type",
			setupRule: func() *kbapi.SecurityDetectionsAPIRuleResponse {
				rule := kbapi.SecurityDetectionsAPIEsqlRule{
					Id:          testUUID,
					RuleId:      "test-esql-rule",
					Name:        "Test ESQL Rule",
					Type:        "esql",
					Query:       "FROM logs | WHERE user.name == \"suspicious_user\"",
					Language:    "esql",
					Enabled:     true,
					RiskScore:   85,
					Severity:    "high",
					Version:     1,
					Description: "Test ESQL rule description",
					From:        "now-6m",
					To:          "now",
					Interval:    "5m",
					CreatedBy:   "test-user",
					UpdatedBy:   "test-user",
					Revision:    1,
				}
				response := &kbapi.SecurityDetectionsAPIRuleResponse{}
				err := response.FromSecurityDetectionsAPIEsqlRule(rule)
				require.NoError(t, err)
				return response
			},
			validateData: func(t *testing.T, data *Data) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceID, testUUID.String()), data.ID.ValueString())
				require.Equal(t, "test-esql-rule", data.RuleID.ValueString())
				require.Equal(t, "Test ESQL Rule", data.Name.ValueString())
				require.Equal(t, "esql", data.Type.ValueString())
				require.Equal(t, "FROM logs | WHERE user.name == \"suspicious_user\"", data.Query.ValueString())
				require.Equal(t, "esql", data.Language.ValueString())
				require.Equal(t, int64(85), data.RiskScore.ValueInt64())
				require.Equal(t, "high", data.Severity.ValueString())
			},
		},
		{
			name: "machine_learning rule type",
			setupRule: func() *kbapi.SecurityDetectionsAPIRuleResponse {
				mlJobID := kbapi.SecurityDetectionsAPIMachineLearningJobId{}
				err := mlJobID.FromSecurityDetectionsAPIMachineLearningJobId0("suspicious_activity")
				require.NoError(t, err)

				rule := kbapi.SecurityDetectionsAPIMachineLearningRule{
					Id:                   testUUID,
					RuleId:               "test-ml-rule",
					Name:                 "Test ML Rule",
					Type:                 "machine_learning",
					MachineLearningJobId: mlJobID,
					AnomalyThreshold:     50,
					Enabled:              true,
					RiskScore:            70,
					Severity:             "medium",
					Version:              1,
					Description:          "Test ML rule description",
					From:                 "now-6m",
					To:                   "now",
					Interval:             "5m",
					CreatedBy:            "test-user",
					UpdatedBy:            "test-user",
					Revision:             1,
				}
				response := &kbapi.SecurityDetectionsAPIRuleResponse{}
				err = response.FromSecurityDetectionsAPIMachineLearningRule(rule)
				require.NoError(t, err)
				return response
			},
			validateData: func(t *testing.T, data *Data) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceID, testUUID.String()), data.ID.ValueString())
				require.Equal(t, "test-ml-rule", data.RuleID.ValueString())
				require.Equal(t, "Test ML Rule", data.Name.ValueString())
				require.Equal(t, "machine_learning", data.Type.ValueString())
				require.Equal(t, int64(50), data.AnomalyThreshold.ValueInt64())
				require.Equal(t, int64(70), data.RiskScore.ValueInt64())
				require.Equal(t, "medium", data.Severity.ValueString())
				require.Len(t, data.MachineLearningJobID.Elements(), 1)
			},
		},
		{
			name: "new_terms rule type",
			setupRule: func() *kbapi.SecurityDetectionsAPIRuleResponse {
				rule := kbapi.SecurityDetectionsAPINewTermsRule{
					Id:                 testUUID,
					RuleId:             "test-new-terms-rule",
					Name:               "Test New Terms Rule",
					Type:               "new_terms",
					Query:              "user.name:*",
					Language:           "kuery",
					NewTermsFields:     []string{"user.name", "host.name"},
					HistoryWindowStart: "now-7d",
					Enabled:            true,
					RiskScore:          60,
					Severity:           "medium",
					Version:            1,
					Description:        "Test new terms rule description",
					From:               "now-6m",
					To:                 "now",
					Interval:           "5m",
					CreatedBy:          "test-user",
					UpdatedBy:          "test-user",
					Revision:           1,
				}
				response := &kbapi.SecurityDetectionsAPIRuleResponse{}
				err := response.FromSecurityDetectionsAPINewTermsRule(rule)
				require.NoError(t, err)
				return response
			},
			validateData: func(t *testing.T, data *Data) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceID, testUUID.String()), data.ID.ValueString())
				require.Equal(t, "test-new-terms-rule", data.RuleID.ValueString())
				require.Equal(t, "Test New Terms Rule", data.Name.ValueString())
				require.Equal(t, "new_terms", data.Type.ValueString())
				require.Equal(t, "user.name:*", data.Query.ValueString())
				require.Equal(t, "now-7d", data.HistoryWindowStart.ValueString())
				require.Equal(t, int64(60), data.RiskScore.ValueInt64())
				require.Equal(t, "medium", data.Severity.ValueString())
				require.Len(t, data.NewTermsFields.Elements(), 2)
			},
		},
		{
			name: "saved_query rule type",
			setupRule: func() *kbapi.SecurityDetectionsAPIRuleResponse {
				rule := kbapi.SecurityDetectionsAPISavedQueryRule{
					Id:          testUUID,
					RuleId:      "test-saved-query-rule",
					Name:        "Test Saved Query Rule",
					Type:        "saved_query",
					SavedId:     "my-saved-query-id",
					Enabled:     true,
					RiskScore:   65,
					Severity:    "medium",
					Version:     1,
					Description: "Test saved query rule description",
					From:        "now-6m",
					To:          "now",
					Interval:    "5m",
					CreatedBy:   "test-user",
					UpdatedBy:   "test-user",
					Revision:    1,
				}
				response := &kbapi.SecurityDetectionsAPIRuleResponse{}
				err := response.FromSecurityDetectionsAPISavedQueryRule(rule)
				require.NoError(t, err)
				return response
			},
			validateData: func(t *testing.T, data *Data) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceID, testUUID.String()), data.ID.ValueString())
				require.Equal(t, "test-saved-query-rule", data.RuleID.ValueString())
				require.Equal(t, "Test Saved Query Rule", data.Name.ValueString())
				require.Equal(t, "saved_query", data.Type.ValueString())
				require.Equal(t, "my-saved-query-id", data.SavedID.ValueString())
				require.Equal(t, int64(65), data.RiskScore.ValueInt64())
				require.Equal(t, "medium", data.Severity.ValueString())
			},
		},
		{
			name: "threat_match rule type",
			setupRule: func() *kbapi.SecurityDetectionsAPIRuleResponse {
				rule := kbapi.SecurityDetectionsAPIThreatMatchRule{
					Id:          testUUID,
					RuleId:      "test-threat-match-rule",
					Name:        "Test Threat Match Rule",
					Type:        "threat_match",
					Query:       "source.ip:*",
					Language:    "kuery",
					ThreatIndex: []string{"threat-intel-*"},
					ThreatMapping: kbapi.SecurityDetectionsAPIThreatMapping{
						{
							Entries: []kbapi.SecurityDetectionsAPIThreatMappingEntry{
								{
									Field: "source.ip",
									Type:  "mapping",
									Value: "threat.indicator.ip",
								},
							},
						},
					},
					Enabled:     true,
					RiskScore:   90,
					Severity:    "critical",
					Version:     1,
					Description: "Test threat match rule description",
					From:        "now-6m",
					To:          "now",
					Interval:    "5m",
					CreatedBy:   "test-user",
					UpdatedBy:   "test-user",
					Revision:    1,
				}
				response := &kbapi.SecurityDetectionsAPIRuleResponse{}
				err := response.FromSecurityDetectionsAPIThreatMatchRule(rule)
				require.NoError(t, err)
				return response
			},
			validateData: func(t *testing.T, data *Data) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceID, testUUID.String()), data.ID.ValueString())
				require.Equal(t, "test-threat-match-rule", data.RuleID.ValueString())
				require.Equal(t, "Test Threat Match Rule", data.Name.ValueString())
				require.Equal(t, "threat_match", data.Type.ValueString())
				require.Equal(t, "source.ip:*", data.Query.ValueString())
				require.Equal(t, int64(90), data.RiskScore.ValueInt64())
				require.Equal(t, "critical", data.Severity.ValueString())
				require.Len(t, data.ThreatIndex.Elements(), 1)
				require.Len(t, data.ThreatMapping.Elements(), 1)
			},
		},
		{
			name: "threshold rule type",
			setupRule: func() *kbapi.SecurityDetectionsAPIRuleResponse {
				thresholdField := kbapi.SecurityDetectionsAPIThresholdField{}
				err := thresholdField.FromSecurityDetectionsAPIThresholdField0("user.name")
				require.NoError(t, err)

				rule := kbapi.SecurityDetectionsAPIThresholdRule{
					Id:       testUUID,
					RuleId:   "test-threshold-rule",
					Name:     "Test Threshold Rule",
					Type:     "threshold",
					Query:    "event.action:login",
					Language: "kuery",
					Threshold: kbapi.SecurityDetectionsAPIThreshold{
						Field: thresholdField,
						Value: 5,
					},
					Enabled:     true,
					RiskScore:   75,
					Severity:    "high",
					Version:     1,
					Description: "Test threshold rule description",
					From:        "now-6m",
					To:          "now",
					Interval:    "5m",
					CreatedBy:   "test-user",
					UpdatedBy:   "test-user",
					Revision:    1,
				}
				response := &kbapi.SecurityDetectionsAPIRuleResponse{}
				err = response.FromSecurityDetectionsAPIThresholdRule(rule)
				require.NoError(t, err)
				return response
			},
			validateData: func(t *testing.T, data *Data) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceID, testUUID.String()), data.ID.ValueString())
				require.Equal(t, "test-threshold-rule", data.RuleID.ValueString())
				require.Equal(t, "Test Threshold Rule", data.Name.ValueString())
				require.Equal(t, "threshold", data.Type.ValueString())
				require.Equal(t, "event.action:login", data.Query.ValueString())
				require.Equal(t, int64(75), data.RiskScore.ValueInt64())
				require.Equal(t, "high", data.Severity.ValueString())
				require.False(t, data.Threshold.IsNull())
			},
		},
		{
			name: "discriminator error",
			setupRule: func() *kbapi.SecurityDetectionsAPIRuleResponse {
				// Create an empty response that will fail discriminator check
				return &kbapi.SecurityDetectionsAPIRuleResponse{}
			},
			expectError:  true,
			errorMessage: "Error determining rule processor",
			validateData: func(_ *testing.T, _ *Data) {
				// No validation needed for error case
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &Data{
				SpaceID: types.StringValue(spaceID),
			}

			response := tt.setupRule()
			diags := data.updateFromRule(ctx, response)

			if tt.expectError {
				require.True(t, diags.HasError())
				require.Contains(t, diags.Errors()[0].Summary(), tt.errorMessage)
			} else {
				require.Empty(t, diags)
				tt.validateData(t, data)
			}
		})
	}
}

func TestCompositeIDOperations(t *testing.T) {
	tests := []struct {
		name               string
		inputID            string
		expectedSpaceID    string
		expectedResourceID string
		shouldError        bool
	}{
		{
			name:               "valid composite id",
			inputID:            "my-space/12345678-1234-1234-1234-123456789012",
			expectedSpaceID:    "my-space",
			expectedResourceID: "12345678-1234-1234-1234-123456789012",
		},
		{
			name:        "invalid composite id format",
			inputID:     "invalid-format",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Data{
				ID: types.StringValue(tt.inputID),
			}

			compID, diags := clients.CompositeIDFromStrFw(data.ID.ValueString())

			if tt.shouldError {
				require.NotEmpty(t, diags)
				return
			}

			require.Empty(t, diags)
			require.Equal(t, tt.expectedSpaceID, compID.ClusterID)
			require.Equal(t, tt.expectedResourceID, compID.ResourceID)
		})
	}
}

func TestResponseActionsToAPI(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name        string
		data        Data
		actionType  string
		shouldError bool
	}{
		{
			name: "osquery response action",
			data: Data{
				ResponseActions: typeutils.ListValueFrom(ctx, []ResponseActionModel{
					{
						ActionTypeID: types.StringValue(".osquery"),
						Params: typeutils.ObjectValueFrom(ctx, &ResponseActionParamsModel{
							Query:        types.StringValue("SELECT * FROM processes"),
							Timeout:      types.Int64Value(300),
							EcsMapping:   types.MapNull(types.StringType),
							Queries:      types.ListNull(getOsqueryQueryElementType()),
							PackID:       types.StringNull(),
							SavedQueryID: types.StringNull(),
							Command:      types.StringNull(),
							Comment:      types.StringNull(),
							Config:       types.ObjectNull(getEndpointProcessConfigType()),
						}, getResponseActionParamsType(), path.Root("response_actions").AtListIndex(0).AtName("params"), &diags),
					},
				}, getResponseActionElementType(), path.Root("response_actions"), &diags),
			},
			actionType: ".osquery",
		},
		{
			name: "endpoint response action - isolate",
			data: Data{
				ResponseActions: typeutils.ListValueFrom(ctx, []ResponseActionModel{
					{
						ActionTypeID: types.StringValue(".endpoint"),
						Params: typeutils.ObjectValueFrom(ctx, &ResponseActionParamsModel{
							Command:      types.StringValue("isolate"),
							Comment:      types.StringValue("Isolating suspicious host"),
							Config:       types.ObjectNull(getEndpointProcessConfigType()),
							Query:        types.StringNull(),
							PackID:       types.StringNull(),
							SavedQueryID: types.StringNull(),
							Timeout:      types.Int64Null(),
							EcsMapping:   types.MapNull(types.StringType),
							Queries:      types.ListNull(getOsqueryQueryElementType()),
						}, getResponseActionParamsType(), path.Root("response_actions").AtListIndex(0).AtName("params"), &diags),
					},
				}, getResponseActionElementType(), path.Root("response_actions"), &diags),
			},
			actionType: ".endpoint",
		},
		{
			name: "unsupported response action type",
			data: Data{
				ResponseActions: typeutils.ListValueFrom(ctx, []ResponseActionModel{
					{
						ActionTypeID: types.StringValue(".unsupported"),
						Params: typeutils.ObjectValueFrom(ctx, &ResponseActionParamsModel{
							Query:        types.StringNull(),
							PackID:       types.StringNull(),
							SavedQueryID: types.StringNull(),
							Timeout:      types.Int64Null(),
							EcsMapping:   types.MapNull(types.StringType),
							Queries:      types.ListNull(getOsqueryQueryElementType()),
							Command:      types.StringValue("unknown"),
							Comment:      types.StringNull(),
							Config:       types.ObjectNull(getEndpointProcessConfigType()),
						}, getResponseActionParamsType(), path.Root("response_actions").AtListIndex(0).AtName("params"), &diags),
					},
				}, getResponseActionElementType(), path.Root("response_actions"), &diags),
			},
			actionType:  ".unsupported",
			shouldError: true,
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responseActions, responseActionsDiags := tt.data.responseActionsToAPI(ctx, NewMockAPIClient())

			if tt.shouldError {
				require.NotEmpty(t, responseActionsDiags)
				return
			}

			require.Empty(t, responseActionsDiags)
			require.Len(t, responseActions, 1)

			// Verify the action type by checking discriminator
			_, err := responseActions[0].ValueByDiscriminator()
			require.NoError(t, err)
		})
	}
}

func TestResponseActionsToAPIVersionCheck(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Test data with response actions
	data := Data{
		ResponseActions: typeutils.ListValueFrom(ctx, []ResponseActionModel{
			{
				ActionTypeID: types.StringValue(".osquery"),
				Params: typeutils.ObjectValueFrom(ctx, &ResponseActionParamsModel{
					Query:        types.StringValue("SELECT * FROM processes"),
					Timeout:      types.Int64Value(300),
					EcsMapping:   types.MapNull(types.StringType),
					Queries:      types.ListNull(getOsqueryQueryElementType()),
					PackID:       types.StringNull(),
					SavedQueryID: types.StringNull(),
					Command:      types.StringNull(),
					Comment:      types.StringNull(),
					Config:       types.ObjectNull(getEndpointProcessConfigType()),
				}, getResponseActionParamsType(), path.Root("response_actions").AtListIndex(0).AtName("params"), &diags),
			},
		}, getResponseActionElementType(), path.Root("response_actions"), &diags),
	}

	require.Empty(t, diags)

	responseActions, responseActionsDiags := data.responseActionsToAPI(ctx, NewMockAPIClient())

	// Should work with the test client and return response actions
	require.Empty(t, responseActionsDiags)
	require.Len(t, responseActions, 1)

	// Verify the action type
	actionValue, err := responseActions[0].ValueByDiscriminator()
	require.NoError(t, err)

	// Verify it's an osquery action
	osqueryAction, ok := actionValue.(kbapi.SecurityDetectionsAPIOsqueryResponseAction)
	require.True(t, ok, "Expected osquery action")
	require.Equal(t, kbapi.SecurityDetectionsAPIOsqueryResponseActionActionTypeId(".osquery"), osqueryAction.ActionTypeId)
}

func TestKQLQueryLanguage(t *testing.T) {
	tests := []struct {
		name     string
		language types.String
		expected *kbapi.SecurityDetectionsAPIKqlQueryLanguage
	}{
		{
			name:     "kuery language",
			language: types.StringValue("kuery"),
			expected: schemautil.Pointer(kbapi.SecurityDetectionsAPIKqlQueryLanguage("kuery")),
		},
		{
			name:     "lucene language",
			language: types.StringValue("lucene"),
			expected: schemautil.Pointer(kbapi.SecurityDetectionsAPIKqlQueryLanguage("lucene")),
		},
		{
			name:     "unknown language defaults to kuery",
			language: types.StringValue("unknown"),
			expected: schemautil.Pointer(kbapi.SecurityDetectionsAPIKqlQueryLanguage("kuery")),
		},
		{
			name:     "null language returns nil",
			language: types.StringNull(),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := Data{
				Language: tt.language,
			}

			result := data.getKQLQueryLanguage()

			if tt.expected == nil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
				require.Equal(t, *tt.expected, *result)
			}
		})
	}
}

func TestExceptionsListToAPI(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := Data{
		ExceptionsList: typeutils.ListValueFrom(ctx, []ExceptionsListModel{
			{
				ID:            types.StringValue("exception-1"),
				ListID:        types.StringValue("trusted-processes"),
				NamespaceType: types.StringValue("single"),
				Type:          types.StringValue("detection"),
			},
			{
				ID:            types.StringValue("exception-2"),
				ListID:        types.StringValue("allow-list"),
				NamespaceType: types.StringValue("agnostic"),
				Type:          types.StringValue("endpoint"),
			},
		}, getExceptionsListElementType(), path.Root("exceptions_list"), &diags),
	}

	require.Empty(t, diags)

	exceptionsList, exceptionsListDiags := data.exceptionsListToAPI(ctx)
	require.Empty(t, exceptionsListDiags)
	require.Len(t, exceptionsList, 2)

	require.Equal(t, "exception-1", exceptionsList[0].Id)
	require.Equal(t, "trusted-processes", exceptionsList[0].ListId)
	require.Equal(t, "single", string(exceptionsList[0].NamespaceType))
	require.Equal(t, "detection", string(exceptionsList[0].Type))

	require.Equal(t, "exception-2", exceptionsList[1].Id)
	require.Equal(t, "allow-list", exceptionsList[1].ListId)
	require.Equal(t, "agnostic", string(exceptionsList[1].NamespaceType))
	require.Equal(t, "endpoint", string(exceptionsList[1].Type))
}

func TestConvertThresholdToModel(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		apiThreshold       kbapi.SecurityDetectionsAPIThreshold
		expectedValue      int64
		expectedFieldCount int
		hasCardinality     bool
	}{
		{
			name: "threshold with single field",
			apiThreshold: func() kbapi.SecurityDetectionsAPIThreshold {
				threshold := kbapi.SecurityDetectionsAPIThreshold{
					Value: 5,
				}
				err := threshold.Field.FromSecurityDetectionsAPIThresholdField0("user.name")
				require.NoError(t, err)
				return threshold
			}(),
			expectedValue:      5,
			expectedFieldCount: 1,
		},
		{
			name: "threshold with multiple fields and cardinality",
			apiThreshold: func() kbapi.SecurityDetectionsAPIThreshold {
				threshold := kbapi.SecurityDetectionsAPIThreshold{
					Value: 10,
					Cardinality: &kbapi.SecurityDetectionsAPIThresholdCardinality{
						{Field: "source.ip", Value: 3},
					},
				}
				err := threshold.Field.FromSecurityDetectionsAPIThresholdField1([]string{"user.name", "process.name"})
				require.NoError(t, err)
				return threshold
			}(),
			expectedValue:      10,
			expectedFieldCount: 2,
			hasCardinality:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thresholdObj, diags := convertThresholdToModel(ctx, tt.apiThreshold)
			require.Empty(t, diags)
			require.False(t, thresholdObj.IsNull())

			var thresholdModel ThresholdModel
			objDiags := thresholdObj.As(ctx, &thresholdModel, basetypes.ObjectAsOptions{})
			require.Empty(t, objDiags)

			require.Equal(t, tt.expectedValue, thresholdModel.Value.ValueInt64())
			require.Len(t, thresholdModel.Field.Elements(), tt.expectedFieldCount)

			if tt.hasCardinality {
				require.False(t, thresholdModel.Cardinality.IsNull())
				require.NotEmpty(t, thresholdModel.Cardinality.Elements())
			}
		})
	}
}

func TestToCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name        string
		ruleType    string
		shouldError bool
		errorMsg    string
		setupData   func() Data
	}{
		{
			name:     "query rule type",
			ruleType: "query",
			setupData: func() Data {
				return Data{
					Type:        types.StringValue("query"),
					Name:        types.StringValue("Test Query Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("user.name:test"),
					Language:    types.StringValue("kuery"),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
		{
			name:     "eql rule type",
			ruleType: "eql",
			setupData: func() Data {
				return Data{
					Type:        types.StringValue("eql"),
					Name:        types.StringValue("Test EQL Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("process where process.name == \"cmd.exe\""),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
		{
			name:     "esql rule type",
			ruleType: "esql",
			setupData: func() Data {
				return Data{
					Type:        types.StringValue("esql"),
					Name:        types.StringValue("Test ESQL Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("FROM logs | WHERE user.name == \"suspicious_user\""),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
		{
			name:     "machine_learning rule type",
			ruleType: "machine_learning",
			setupData: func() Data {
				return Data{
					Type:                 types.StringValue("machine_learning"),
					Name:                 types.StringValue("Test ML Rule"),
					Description:          types.StringValue("Test description"),
					AnomalyThreshold:     types.Int64Value(50),
					MachineLearningJobID: typeutils.ListValueFrom(ctx, []string{"suspicious_activity"}, types.StringType, path.Root("machine_learning_job_id"), &diags),
					RiskScore:            types.Int64Value(75),
					Severity:             types.StringValue("medium"),
				}
			},
		},
		{
			name:     "new_terms rule type",
			ruleType: "new_terms",
			setupData: func() Data {
				return Data{
					Type:               types.StringValue("new_terms"),
					Name:               types.StringValue("Test New Terms Rule"),
					Description:        types.StringValue("Test description"),
					Query:              types.StringValue("user.name:*"),
					NewTermsFields:     typeutils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("new_terms_fields"), &diags),
					HistoryWindowStart: types.StringValue("now-7d"),
					RiskScore:          types.Int64Value(75),
					Severity:           types.StringValue("medium"),
				}
			},
		},
		{
			name:     "saved_query rule type",
			ruleType: "saved_query",
			setupData: func() Data {
				return Data{
					Type:        types.StringValue("saved_query"),
					Name:        types.StringValue("Test Saved Query Rule"),
					Description: types.StringValue("Test description"),
					SavedID:     types.StringValue("my-saved-query"),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
		{
			name:     "threat_match rule type",
			ruleType: "threat_match",
			setupData: func() Data {
				return Data{
					Type:        types.StringValue("threat_match"),
					Name:        types.StringValue("Test Threat Match Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("source.ip:*"),
					ThreatIndex: typeutils.ListValueFrom(ctx, []string{"threat-intel-*"}, types.StringType, path.Root("threat_index"), &diags),
					ThreatMapping: typeutils.ListValueFrom(ctx, []TfDataItem{
						{
							Entries: typeutils.ListValueFrom(ctx, []TfDataItemEntry{
								{
									Field: types.StringValue("source.ip"),
									Type:  types.StringValue("mapping"),
									Value: types.StringValue("threat.indicator.ip"),
								},
							}, getThreatMappingEntryElementType(), path.Root("threat_mapping").AtListIndex(0).AtName("entries"), &diags),
						},
					}, getThreatMappingElementType(), path.Root("threat_mapping"), &diags),
					RiskScore: types.Int64Value(75),
					Severity:  types.StringValue("medium"),
				}
			},
		},
		{
			name:     "threshold rule type",
			ruleType: "threshold",
			setupData: func() Data {
				return Data{
					Type:        types.StringValue("threshold"),
					Name:        types.StringValue("Test Threshold Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("event.action:login"),
					Threshold: typeutils.ObjectValueFrom(ctx, &ThresholdModel{
						Field:       typeutils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("threshold").AtName("field"), &diags),
						Value:       types.Int64Value(5),
						Cardinality: types.ListNull(getCardinalityType()),
					}, getThresholdType(), path.Root("threshold"), &diags),
					RiskScore: types.Int64Value(75),
					Severity:  types.StringValue("medium"),
				}
			},
		},
		{
			name:        "unsupported rule type",
			ruleType:    "unsupported_type",
			shouldError: true,
			errorMsg:    "Rule type 'unsupported_type' is not supported",
			setupData: func() Data {
				return Data{
					Type:        types.StringValue("unsupported_type"),
					Name:        types.StringValue("Test Unsupported Rule"),
					Description: types.StringValue("Test description"),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
	}

	require.Empty(t, diags) // Check for any setup errors

	const (
		ruleTypeQuery           = "query"
		ruleTypeEQL             = "eql"
		ruleTypeESQL            = "esql"
		ruleTypeMachineLearning = "machine_learning"
		ruleTypeNewTerms        = "new_terms"
		ruleTypeSavedQuery      = "saved_query"
		ruleTypeThreatMatch     = "threat_match"
		ruleTypeThreshold       = "threshold"
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.setupData()

			createProps, createDiags := data.toCreateProps(ctx, NewMockAPIClient())

			if tt.shouldError {
				require.True(t, createDiags.HasError())
				require.Contains(t, createDiags.Errors()[0].Summary(), "Unsupported rule type")
				require.Contains(t, createDiags.Errors()[0].Detail(), tt.errorMsg)
				return
			}

			require.Empty(t, createDiags)

			// Verify that the create props can be converted to the expected rule type and check values
			switch tt.ruleType {
			case ruleTypeQuery:
				queryRule, err := createProps.AsSecurityDetectionsAPIQueryRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Query Rule", queryRule.Name)
				require.Equal(t, "Test description", queryRule.Description)
				require.Equal(t, ruleTypeQuery, string(queryRule.Type))
				require.Equal(t, "user.name:test", *queryRule.Query)
				require.Equal(t, "kuery", string(*queryRule.Language))
				require.Equal(t, int64(75), int64(queryRule.RiskScore))
				require.Equal(t, "medium", string(queryRule.Severity))
			case ruleTypeEQL:
				eqlRule, err := createProps.AsSecurityDetectionsAPIEqlRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test EQL Rule", eqlRule.Name)
				require.Equal(t, "Test description", eqlRule.Description)
				require.Equal(t, ruleTypeEQL, string(eqlRule.Type))
				require.Equal(t, "process where process.name == \"cmd.exe\"", eqlRule.Query)
				require.Equal(t, "eql", string(eqlRule.Language))
				require.Equal(t, int64(75), int64(eqlRule.RiskScore))
				require.Equal(t, "medium", string(eqlRule.Severity))
			case ruleTypeESQL:
				esqlRule, err := createProps.AsSecurityDetectionsAPIEsqlRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test ESQL Rule", esqlRule.Name)
				require.Equal(t, "Test description", esqlRule.Description)
				require.Equal(t, ruleTypeESQL, string(esqlRule.Type))
				require.Equal(t, "FROM logs | WHERE user.name == \"suspicious_user\"", esqlRule.Query)
				require.Equal(t, "esql", string(esqlRule.Language))
				require.Equal(t, int64(75), int64(esqlRule.RiskScore))
				require.Equal(t, "medium", string(esqlRule.Severity))
			case ruleTypeMachineLearning:
				mlRule, err := createProps.AsSecurityDetectionsAPIMachineLearningRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test ML Rule", mlRule.Name)
				require.Equal(t, "Test description", mlRule.Description)
				require.Equal(t, ruleTypeMachineLearning, string(mlRule.Type))
				require.Equal(t, int64(50), int64(mlRule.AnomalyThreshold))
				require.Equal(t, int64(75), int64(mlRule.RiskScore))
				require.Equal(t, "medium", string(mlRule.Severity))
				// Verify ML job ID is set correctly
				jobID, err := mlRule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId1()
				require.NoError(t, err)
				require.Equal(t, []string{"suspicious_activity"}, jobID)
			case ruleTypeNewTerms:
				newTermsRule, err := createProps.AsSecurityDetectionsAPINewTermsRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test New Terms Rule", newTermsRule.Name)
				require.Equal(t, "Test description", newTermsRule.Description)
				require.Equal(t, ruleTypeNewTerms, string(newTermsRule.Type))
				require.Equal(t, "user.name:*", newTermsRule.Query)
				require.Equal(t, "now-7d", newTermsRule.HistoryWindowStart)
				require.Equal(t, int64(75), int64(newTermsRule.RiskScore))
				require.Equal(t, "medium", string(newTermsRule.Severity))
				require.Len(t, newTermsRule.NewTermsFields, 1)
				require.Equal(t, "user.name", newTermsRule.NewTermsFields[0])
			case ruleTypeSavedQuery:
				savedQueryRule, err := createProps.AsSecurityDetectionsAPISavedQueryRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Saved Query Rule", savedQueryRule.Name)
				require.Equal(t, "Test description", savedQueryRule.Description)
				require.Equal(t, ruleTypeSavedQuery, string(savedQueryRule.Type))
				require.Equal(t, "my-saved-query", savedQueryRule.SavedId)
				require.Equal(t, int64(75), int64(savedQueryRule.RiskScore))
				require.Equal(t, "medium", string(savedQueryRule.Severity))
			case ruleTypeThreatMatch:
				threatMatchRule, err := createProps.AsSecurityDetectionsAPIThreatMatchRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Threat Match Rule", threatMatchRule.Name)
				require.Equal(t, "Test description", threatMatchRule.Description)
				require.Equal(t, ruleTypeThreatMatch, string(threatMatchRule.Type))
				require.Equal(t, "source.ip:*", threatMatchRule.Query)
				require.Equal(t, int64(75), int64(threatMatchRule.RiskScore))
				require.Equal(t, "medium", string(threatMatchRule.Severity))
				require.Len(t, threatMatchRule.ThreatIndex, 1)
				require.Equal(t, "threat-intel-*", threatMatchRule.ThreatIndex[0])
				require.Len(t, threatMatchRule.ThreatMapping, 1)
			case ruleTypeThreshold:
				thresholdRule, err := createProps.AsSecurityDetectionsAPIThresholdRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Threshold Rule", thresholdRule.Name)
				require.Equal(t, "Test description", thresholdRule.Description)
				require.Equal(t, ruleTypeThreshold, string(thresholdRule.Type))
				require.Equal(t, "event.action:login", thresholdRule.Query)
				require.Equal(t, int64(75), int64(thresholdRule.RiskScore))
				require.Equal(t, "medium", string(thresholdRule.Severity))
				require.NotNil(t, thresholdRule.Threshold)
				require.Equal(t, int64(5), int64(thresholdRule.Threshold.Value))
				// Check single field
				singleField, err := thresholdRule.Threshold.Field.AsSecurityDetectionsAPIThresholdField0()
				require.NoError(t, err)
				require.Equal(t, "user.name", singleField)
			}
		})
	}
}

func TestToUpdateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Create a valid composite ID for testing
	testUUID := uuid.New()
	testSpaceID := "test-space"
	validCompositeID := fmt.Sprintf("%s/%s", testSpaceID, testUUID.String())

	tests := []struct {
		name        string
		ruleType    string
		shouldError bool
		errorMsg    string
		setupData   func() Data
	}{
		{
			name:     "query rule type",
			ruleType: "query",
			setupData: func() Data {
				return Data{
					ID:          types.StringValue(validCompositeID),
					Type:        types.StringValue("query"),
					Name:        types.StringValue("Test Query Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("user.name:test"),
					Language:    types.StringValue("kuery"),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
		{
			name:     "eql rule type",
			ruleType: "eql",
			setupData: func() Data {
				return Data{
					ID:          types.StringValue(validCompositeID),
					Type:        types.StringValue("eql"),
					Name:        types.StringValue("Test EQL Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("process where process.name == \"cmd.exe\""),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
		{
			name:     "esql rule type",
			ruleType: "esql",
			setupData: func() Data {
				return Data{
					ID:          types.StringValue(validCompositeID),
					Type:        types.StringValue("esql"),
					Name:        types.StringValue("Test ESQL Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("FROM logs | WHERE user.name == \"suspicious_user\""),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
		{
			name:     "machine_learning rule type",
			ruleType: "machine_learning",
			setupData: func() Data {
				return Data{
					ID:                   types.StringValue(validCompositeID),
					Type:                 types.StringValue("machine_learning"),
					Name:                 types.StringValue("Test ML Rule"),
					Description:          types.StringValue("Test description"),
					AnomalyThreshold:     types.Int64Value(50),
					MachineLearningJobID: typeutils.ListValueFrom(ctx, []string{"suspicious_activity"}, types.StringType, path.Root("machine_learning_job_id"), &diags),
					RiskScore:            types.Int64Value(75),
					Severity:             types.StringValue("medium"),
				}
			},
		},
		{
			name:     "new_terms rule type",
			ruleType: "new_terms",
			setupData: func() Data {
				return Data{
					ID:                 types.StringValue(validCompositeID),
					Type:               types.StringValue("new_terms"),
					Name:               types.StringValue("Test New Terms Rule"),
					Description:        types.StringValue("Test description"),
					Query:              types.StringValue("user.name:*"),
					NewTermsFields:     typeutils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("new_terms_fields"), &diags),
					HistoryWindowStart: types.StringValue("now-7d"),
					RiskScore:          types.Int64Value(75),
					Severity:           types.StringValue("medium"),
				}
			},
		},
		{
			name:     "saved_query rule type",
			ruleType: "saved_query",
			setupData: func() Data {
				return Data{
					ID:          types.StringValue(validCompositeID),
					Type:        types.StringValue("saved_query"),
					Name:        types.StringValue("Test Saved Query Rule"),
					Description: types.StringValue("Test description"),
					SavedID:     types.StringValue("my-saved-query"),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
		{
			name:     "threat_match rule type",
			ruleType: "threat_match",
			setupData: func() Data {
				return Data{
					ID:          types.StringValue(validCompositeID),
					Type:        types.StringValue("threat_match"),
					Name:        types.StringValue("Test Threat Match Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("source.ip:*"),
					ThreatIndex: typeutils.ListValueFrom(ctx, []string{"threat-intel-*"}, types.StringType, path.Root("threat_index"), &diags),
					ThreatMapping: typeutils.ListValueFrom(ctx, []TfDataItem{
						{
							Entries: typeutils.ListValueFrom(ctx, []TfDataItemEntry{
								{
									Field: types.StringValue("source.ip"),
									Type:  types.StringValue("mapping"),
									Value: types.StringValue("threat.indicator.ip"),
								},
							}, getThreatMappingEntryElementType(), path.Root("threat_mapping").AtListIndex(0).AtName("entries"), &diags),
						},
					}, getThreatMappingElementType(), path.Root("threat_mapping"), &diags),
					RiskScore: types.Int64Value(75),
					Severity:  types.StringValue("medium"),
				}
			},
		},
		{
			name:     "threshold rule type",
			ruleType: "threshold",
			setupData: func() Data {
				return Data{
					ID:          types.StringValue(validCompositeID),
					Type:        types.StringValue("threshold"),
					Name:        types.StringValue("Test Threshold Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("event.action:login"),
					Threshold: typeutils.ObjectValueFrom(ctx, &ThresholdModel{
						Field:       typeutils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("threshold").AtName("field"), &diags),
						Value:       types.Int64Value(5),
						Cardinality: types.ListNull(getCardinalityType()),
					}, getThresholdType(), path.Root("threshold"), &diags),
					RiskScore: types.Int64Value(75),
					Severity:  types.StringValue("medium"),
				}
			},
		},
		{
			name:        "unsupported rule type",
			ruleType:    "unsupported_type",
			shouldError: true,
			errorMsg:    "Rule type 'unsupported_type' is not supported",
			setupData: func() Data {
				return Data{
					ID:          types.StringValue(validCompositeID),
					Type:        types.StringValue("unsupported_type"),
					Name:        types.StringValue("Test Unsupported Rule"),
					Description: types.StringValue("Test description"),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
	}

	require.Empty(t, diags) // Check for any setup errors

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.setupData()

			updateProps, updateDiags := data.toUpdateProps(ctx, NewMockAPIClient())

			if tt.shouldError {
				require.True(t, updateDiags.HasError())
				require.Contains(t, updateDiags.Errors()[0].Summary(), "Unsupported rule type")
				require.Contains(t, updateDiags.Errors()[0].Detail(), tt.errorMsg)
				return
			}

			require.Empty(t, updateDiags)

			// Verify that the update props can be converted to the expected rule type and check values
			switch tt.ruleType {
			case "query":
				queryRule, err := updateProps.AsSecurityDetectionsAPIQueryRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Query Rule", queryRule.Name)
				require.Equal(t, "Test description", queryRule.Description)
				require.Equal(t, "user.name:test", *queryRule.Query)
				require.Equal(t, "kuery", string(*queryRule.Language))
				require.Equal(t, int64(75), int64(queryRule.RiskScore))
				require.Equal(t, "medium", string(queryRule.Severity))
			case "eql":
				eqlRule, err := updateProps.AsSecurityDetectionsAPIEqlRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test EQL Rule", eqlRule.Name)
				require.Equal(t, "Test description", eqlRule.Description)
				require.Equal(t, "process where process.name == \"cmd.exe\"", eqlRule.Query)
				require.Equal(t, int64(75), int64(eqlRule.RiskScore))
				require.Equal(t, "medium", string(eqlRule.Severity))
			case "esql":
				esqlRule, err := updateProps.AsSecurityDetectionsAPIEsqlRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test ESQL Rule", esqlRule.Name)
				require.Equal(t, "Test description", esqlRule.Description)
				require.Equal(t, "FROM logs | WHERE user.name == \"suspicious_user\"", esqlRule.Query)
				require.Equal(t, int64(75), int64(esqlRule.RiskScore))
				require.Equal(t, "medium", string(esqlRule.Severity))
			case "machine_learning":
				mlRule, err := updateProps.AsSecurityDetectionsAPIMachineLearningRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test ML Rule", mlRule.Name)
				require.Equal(t, "Test description", mlRule.Description)
				require.Equal(t, int64(50), int64(mlRule.AnomalyThreshold))
				require.Equal(t, int64(75), int64(mlRule.RiskScore))
				require.Equal(t, "medium", string(mlRule.Severity))
				// Verify ML job ID is set correctly
				jobID, err := mlRule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId1()
				require.NoError(t, err)
				require.Equal(t, []string{"suspicious_activity"}, jobID)
			case "new_terms":
				newTermsRule, err := updateProps.AsSecurityDetectionsAPINewTermsRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test New Terms Rule", newTermsRule.Name)
				require.Equal(t, "Test description", newTermsRule.Description)
				require.Equal(t, "user.name:*", newTermsRule.Query)
				require.Equal(t, "now-7d", newTermsRule.HistoryWindowStart)
				require.Equal(t, int64(75), int64(newTermsRule.RiskScore))
				require.Equal(t, "medium", string(newTermsRule.Severity))
				require.Len(t, newTermsRule.NewTermsFields, 1)
				require.Equal(t, "user.name", newTermsRule.NewTermsFields[0])
			case "saved_query":
				savedQueryRule, err := updateProps.AsSecurityDetectionsAPISavedQueryRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Saved Query Rule", savedQueryRule.Name)
				require.Equal(t, "Test description", savedQueryRule.Description)
				require.Equal(t, "my-saved-query", savedQueryRule.SavedId)
				require.Equal(t, int64(75), int64(savedQueryRule.RiskScore))
				require.Equal(t, "medium", string(savedQueryRule.Severity))
			case "threat_match":
				threatMatchRule, err := updateProps.AsSecurityDetectionsAPIThreatMatchRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Threat Match Rule", threatMatchRule.Name)
				require.Equal(t, "Test description", threatMatchRule.Description)
				require.Equal(t, "source.ip:*", threatMatchRule.Query)
				require.Equal(t, int64(75), int64(threatMatchRule.RiskScore))
				require.Equal(t, "medium", string(threatMatchRule.Severity))
				require.Len(t, threatMatchRule.ThreatIndex, 1)
				require.Equal(t, "threat-intel-*", threatMatchRule.ThreatIndex[0])
				require.Len(t, threatMatchRule.ThreatMapping, 1)
			case "threshold":
				thresholdRule, err := updateProps.AsSecurityDetectionsAPIThresholdRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Threshold Rule", thresholdRule.Name)
				require.Equal(t, "Test description", thresholdRule.Description)
				require.Equal(t, "event.action:login", thresholdRule.Query)
				require.Equal(t, int64(75), int64(thresholdRule.RiskScore))
				require.Equal(t, "medium", string(thresholdRule.Severity))
				require.NotNil(t, thresholdRule.Threshold)
				require.Equal(t, int64(5), int64(thresholdRule.Threshold.Value))
				// Check single field
				singleField, err := thresholdRule.Threshold.Field.AsSecurityDetectionsAPIThresholdField0()
				require.NoError(t, err)
				require.Equal(t, "user.name", singleField)
			}
		})
	}
}

func TestParseDurationToAPI(t *testing.T) {
	tests := []struct {
		name         string
		duration     customtypes.Duration
		expectedVal  int
		expectedUnit kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnit
		expectError  bool
	}{
		{
			name:         "valid seconds",
			duration:     customtypes.NewDurationValue("30s"),
			expectedVal:  30,
			expectedUnit: kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitS,
			expectError:  false,
		},
		{
			name:         "valid minutes",
			duration:     customtypes.NewDurationValue("5m"),
			expectedVal:  5,
			expectedUnit: kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitM,
			expectError:  false,
		},
		{
			name:         "valid hours",
			duration:     customtypes.NewDurationValue("2h"),
			expectedVal:  2,
			expectedUnit: kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitH,
			expectError:  false,
		},
		{
			name:         "valid days converted to hours",
			duration:     customtypes.NewDurationValue("1d"),
			expectedVal:  24,
			expectedUnit: kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitH,
			expectError:  false,
		},
		{
			name:         "multiple days converted to hours",
			duration:     customtypes.NewDurationValue("3d"),
			expectedVal:  72,
			expectedUnit: kbapi.SecurityDetectionsAPIAlertSuppressionDurationUnitH,
			expectError:  false,
		},
		{
			name:        "invalid format - no unit",
			duration:    customtypes.NewDurationValue("30"),
			expectError: true,
		},
		{
			name:        "invalid format - non-numeric value",
			duration:    customtypes.NewDurationValue("ABCs"),
			expectError: true,
		},
		{
			name:        "invalid format - unsupported unit",
			duration:    customtypes.NewDurationValue("30w"),
			expectError: true,
		},
		{
			name:        "invalid format - empty string",
			duration:    customtypes.NewDurationValue(""),
			expectError: true,
		},
		{
			name:        "null duration",
			duration:    customtypes.NewDurationNull(),
			expectError: true,
		},
		{
			name:        "unknown duration",
			duration:    customtypes.NewDurationUnknown(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := parseDurationToAPI(tt.duration)

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
				return
			}

			require.False(t, diags.HasError(), "Unexpected error: %v", diags)
			require.Equal(t, tt.expectedVal, result.Value)
			require.Equal(t, tt.expectedUnit, result.Unit)
		})
	}
}
