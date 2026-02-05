package security_detection_rule

import (
	"context"
	"fmt"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils/customtypes"
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

type mockApiClient struct {
	serverVersion *version.Version
	serverFlavor  string
	enforceResult bool
}

func (m mockApiClient) EnforceMinVersion(ctx context.Context, minVersion *version.Version) (bool, v2Diag.Diagnostics) {
	supported := m.serverVersion.GreaterThanOrEqual(minVersion)
	return supported, nil
}

// NewMockApiClient creates a new mock API client with default values that support response actions
// This can be used in tests where you need to pass a client to functions like toUpdateProps
func NewMockApiClient() clients.MinVersionEnforceable {
	// Use version 8.16.0 by default to support response actions
	v, _ := version.NewVersion("8.16.0")

	return mockApiClient{
		serverVersion: v,
		serverFlavor:  "default",
		enforceResult: true,
	}
}

// NewMockApiClientWithVersion creates a mock API client with a specific version
// Use this when you need to test specific version behavior
func NewMockApiClientWithVersion(versionStr string) *mockApiClient {
	v, err := version.NewVersion(versionStr)
	if err != nil {
		panic(fmt.Sprintf("Invalid version in test: %s", versionStr))
	}
	return &mockApiClient{
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
		spaceId  string
		expected SecurityDetectionRuleData
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
				Index:          utils.Pointer([]string{"logs-*", "metrics-*"}),
				CreatedBy:      "test-user",
				UpdatedBy:      "test-user",
				Revision:       1,
				FalsePositives: []string{"Known false positive"},
				References:     []string{"https://example.com/test"},
				License:        utils.Pointer(kbapi.SecurityDetectionsAPIRuleLicense("MIT")),
				Note:           utils.Pointer(kbapi.SecurityDetectionsAPIInvestigationGuide("Investigation note")),
				Setup:          "Setup instructions",
			},
			expected: SecurityDetectionRuleData{
				Id:             types.StringValue("test-space/12345678-1234-1234-1234-123456789012"),
				SpaceId:        types.StringValue("test-space"),
				RuleId:         types.StringValue("test-rule-id"),
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
				Author:         utils.ListValueFrom(ctx, []string{"Test Author"}, types.StringType, path.Root("author"), &diags),
				Tags:           utils.ListValueFrom(ctx, []string{"test", "detection"}, types.StringType, path.Root("tags"), &diags),
				Index:          utils.ListValueFrom(ctx, []string{"logs-*", "metrics-*"}, types.StringType, path.Root("index"), &diags),
				CreatedBy:      types.StringValue("test-user"),
				UpdatedBy:      types.StringValue("test-user"),
				Revision:       types.Int64Value(1),
				FalsePositives: utils.ListValueFrom(ctx, []string{"Known false positive"}, types.StringType, path.Root("false_positives"), &diags),
				References:     utils.ListValueFrom(ctx, []string{"https://example.com/test"}, types.StringType, path.Root("references"), &diags),
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
			expected: SecurityDetectionRuleData{
				Id:          types.StringValue("default/87654321-4321-4321-4321-210987654321"),
				SpaceId:     types.StringValue("default"),
				RuleId:      types.StringValue("minimal-rule"),
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
			data := SecurityDetectionRuleData{
				SpaceId: types.StringValue(tt.spaceId),
			}

			diags := updateFromQueryRule(ctx, &tt.rule, &data)
			require.Empty(t, diags)

			// Compare key fields
			require.Equal(t, tt.expected.Id, data.Id)
			require.Equal(t, tt.expected.RuleId, data.RuleId)
			require.Equal(t, tt.expected.Name, data.Name)
			require.Equal(t, tt.expected.Type, data.Type)
			require.Equal(t, tt.expected.Query, data.Query)
			require.Equal(t, tt.expected.Language, data.Language)
			require.Equal(t, tt.expected.Enabled, data.Enabled)
			require.Equal(t, tt.expected.RiskScore, data.RiskScore)
			require.Equal(t, tt.expected.Severity, data.Severity)

			// Verify list fields have correct length
			require.Equal(t, len(tt.expected.Author.Elements()), len(data.Author.Elements()))
			require.Equal(t, len(tt.expected.Tags.Elements()), len(data.Tags.Elements()))
			require.Equal(t, len(tt.expected.Index.Elements()), len(data.Index.Elements()))
		})
	}
}

func TestToQueryRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name               string
		data               SecurityDetectionRuleData
		expectedName       string
		expectedType       string
		expectedQuery      string
		expectedRiskScore  int64
		expectedSeverity   string
		shouldHaveLanguage bool
		shouldHaveIndex    bool
		shouldHaveActions  bool
		shouldHaveRuleId   bool
		shouldError        bool
	}{
		{
			name: "complete query rule create",
			data: SecurityDetectionRuleData{
				Name:        types.StringValue("Test Create Rule"),
				Type:        types.StringValue("query"),
				Query:       types.StringValue("process.name:malicious"),
				Language:    types.StringValue("kuery"),
				RiskScore:   types.Int64Value(85),
				Severity:    types.StringValue("high"),
				Description: types.StringValue("Test rule description"),
				Index:       utils.ListValueFrom(ctx, []string{"winlogbeat-*"}, types.StringType, path.Root("index"), &diags),
				Author:      utils.ListValueFrom(ctx, []string{"Security Team"}, types.StringType, path.Root("author"), &diags),
				Enabled:     types.BoolValue(true),
				RuleId:      types.StringValue("custom-rule-id"),
			},
			expectedName:       "Test Create Rule",
			expectedType:       "query",
			expectedQuery:      "process.name:malicious",
			expectedRiskScore:  85,
			expectedSeverity:   "high",
			shouldHaveLanguage: true,
			shouldHaveIndex:    true,
			shouldHaveRuleId:   true,
		},
		{
			name: "minimal query rule create",
			data: SecurityDetectionRuleData{
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
			createProps, createDiags := toQueryRuleCreateProps(ctx, NewMockApiClient(), tt.data)

			if tt.shouldError {
				require.NotEmpty(t, createDiags)
				return
			}

			require.Empty(t, createDiags)

			// Extract the concrete type from the union
			queryRule, err := createProps.AsSecurityDetectionsAPIQueryRuleCreateProps()
			require.NoError(t, err)

			require.Equal(t, tt.expectedName, string(queryRule.Name))
			require.Equal(t, tt.expectedType, string(queryRule.Type))
			require.NotNil(t, queryRule.Query)
			require.Equal(t, tt.expectedQuery, string(*queryRule.Query))
			require.Equal(t, tt.expectedRiskScore, int64(queryRule.RiskScore))
			require.Equal(t, tt.expectedSeverity, string(queryRule.Severity))

			if tt.shouldHaveLanguage {
				require.NotNil(t, queryRule.Language)
			}

			if tt.shouldHaveIndex {
				require.NotNil(t, queryRule.Index)
				require.NotEmpty(t, *queryRule.Index)
			}

			if tt.shouldHaveRuleId {
				require.NotNil(t, queryRule.RuleId)
				require.Equal(t, "custom-rule-id", string(*queryRule.RuleId))
			}
		})
	}
}

func TestToEqlRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := SecurityDetectionRuleData{
		Name:            types.StringValue("EQL Test Rule"),
		Type:            types.StringValue("eql"),
		Query:           types.StringValue("process where process.name == \"cmd.exe\""),
		RiskScore:       types.Int64Value(60),
		Severity:        types.StringValue("medium"),
		Description:     types.StringValue("EQL rule description"),
		TiebreakerField: types.StringValue("@timestamp"),
	}

	createProps, createDiags := toEqlRuleCreateProps(ctx, NewMockApiClient(), data)
	require.Empty(t, createDiags)

	eqlRule, err := createProps.AsSecurityDetectionsAPIEqlRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "EQL Test Rule", string(eqlRule.Name))
	require.Equal(t, "eql", string(eqlRule.Type))
	require.Equal(t, "process where process.name == \"cmd.exe\"", string(eqlRule.Query))
	require.Equal(t, "eql", string(eqlRule.Language))
	require.Equal(t, int64(60), int64(eqlRule.RiskScore))
	require.Equal(t, "medium", string(eqlRule.Severity))

	require.NotNil(t, eqlRule.TiebreakerField)
	require.Equal(t, "@timestamp", string(*eqlRule.TiebreakerField))

	require.Empty(t, diags)
}

func TestToMachineLearningRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name               string
		data               SecurityDetectionRuleData
		expectedJobCount   int
		shouldHaveSingle   bool
		shouldHaveMultiple bool
	}{
		{
			name: "single ML job",
			data: SecurityDetectionRuleData{
				Name:                 types.StringValue("ML Test Rule"),
				Type:                 types.StringValue("machine_learning"),
				RiskScore:            types.Int64Value(70),
				Severity:             types.StringValue("high"),
				Description:          types.StringValue("ML rule description"),
				AnomalyThreshold:     types.Int64Value(50),
				MachineLearningJobId: utils.ListValueFrom(ctx, []string{"suspicious_activity"}, types.StringType, path.Root("machine_learning_job_id"), &diags),
			},
			expectedJobCount:   1,
			shouldHaveMultiple: true,
		},
		{
			name: "multiple ML jobs",
			data: SecurityDetectionRuleData{
				Name:                 types.StringValue("ML Multi Job Rule"),
				Type:                 types.StringValue("machine_learning"),
				RiskScore:            types.Int64Value(80),
				Severity:             types.StringValue("critical"),
				Description:          types.StringValue("ML multi job rule"),
				AnomalyThreshold:     types.Int64Value(75),
				MachineLearningJobId: utils.ListValueFrom(ctx, []string{"job1", "job2", "job3"}, types.StringType, path.Root("machine_learning_job_id"), &diags),
			},
			expectedJobCount:   3,
			shouldHaveMultiple: true,
		},
	}

	require.Empty(t, diags)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createProps, createDiags := tt.data.toMachineLearningRuleCreateProps(ctx, NewMockApiClient())
			require.Empty(t, createDiags)

			mlRule, err := createProps.AsSecurityDetectionsAPIMachineLearningRuleCreateProps()
			require.NoError(t, err)

			require.Equal(t, tt.data.Name.ValueString(), string(mlRule.Name))
			require.Equal(t, "machine_learning", string(mlRule.Type))
			require.Equal(t, tt.data.AnomalyThreshold.ValueInt64(), int64(mlRule.AnomalyThreshold))

			if tt.shouldHaveSingle {
				ingleJobId, err := mlRule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId0()
				require.NoError(t, err)
				require.Equal(t, "suspicious_activity", string(ingleJobId))
			}

			if tt.shouldHaveMultiple {
				multipleJobIds, err := mlRule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId1()
				require.NoError(t, err)
				require.Len(t, multipleJobIds, tt.expectedJobCount)
			}
		})
	}
}

func TestToEsqlRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := SecurityDetectionRuleData{
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
		Author:      utils.ListValueFrom(ctx, []string{"Security Team"}, types.StringType, path.Root("author"), &diags),
		Tags:        utils.ListValueFrom(ctx, []string{"esql", "test"}, types.StringType, path.Root("tags"), &diags),
	}

	require.Empty(t, diags)

	createProps, createDiags := data.toEsqlRuleCreateProps(ctx, NewMockApiClient())
	require.Empty(t, createDiags)

	esqlRule, err := createProps.AsSecurityDetectionsAPIEsqlRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "Test ESQL Rule", string(esqlRule.Name))
	require.Equal(t, "Test ESQL rule description", string(esqlRule.Description))
	require.Equal(t, "esql", string(esqlRule.Type))
	require.Equal(t, "FROM logs | WHERE user.name == \"suspicious_user\"", string(esqlRule.Query))
	require.Equal(t, "esql", string(esqlRule.Language))
	require.Equal(t, int64(85), int64(esqlRule.RiskScore))
	require.Equal(t, "high", string(esqlRule.Severity))
}

func TestToNewTermsRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := SecurityDetectionRuleData{
		Type:               types.StringValue("new_terms"),
		Name:               types.StringValue("Test New Terms Rule"),
		Description:        types.StringValue("Test new terms rule description"),
		Query:              types.StringValue("user.name:*"),
		Language:           types.StringValue("kuery"),
		NewTermsFields:     utils.ListValueFrom(ctx, []string{"user.name", "host.name"}, types.StringType, path.Root("new_terms_fields"), &diags),
		HistoryWindowStart: types.StringValue("now-7d"),
		RiskScore:          types.Int64Value(60),
		Severity:           types.StringValue("medium"),
		Enabled:            types.BoolValue(true),
		From:               types.StringValue("now-6m"),
		To:                 types.StringValue("now"),
		Interval:           types.StringValue("5m"),
		Index:              utils.ListValueFrom(ctx, []string{"logs-*"}, types.StringType, path.Root("index"), &diags),
	}

	require.Empty(t, diags)

	createProps, createDiags := data.toNewTermsRuleCreateProps(ctx, NewMockApiClient())
	require.Empty(t, createDiags)

	newTermsRule, err := createProps.AsSecurityDetectionsAPINewTermsRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "Test New Terms Rule", string(newTermsRule.Name))
	require.Equal(t, "Test new terms rule description", string(newTermsRule.Description))
	require.Equal(t, "new_terms", string(newTermsRule.Type))
	require.Equal(t, "user.name:*", string(newTermsRule.Query))
	require.Equal(t, "now-7d", string(newTermsRule.HistoryWindowStart))
	require.Equal(t, int64(60), int64(newTermsRule.RiskScore))
	require.Equal(t, "medium", string(newTermsRule.Severity))
	require.Len(t, newTermsRule.NewTermsFields, 2)
	require.Contains(t, newTermsRule.NewTermsFields, "user.name")
	require.Contains(t, newTermsRule.NewTermsFields, "host.name")
}

func TestToSavedQueryRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := SecurityDetectionRuleData{
		Type:        types.StringValue("saved_query"),
		Name:        types.StringValue("Test Saved Query Rule"),
		Description: types.StringValue("Test saved query rule description"),
		SavedId:     types.StringValue("my-saved-query-id"),
		RiskScore:   types.Int64Value(70),
		Severity:    types.StringValue("high"),
		Enabled:     types.BoolValue(true),
		From:        types.StringValue("now-30m"),
		To:          types.StringValue("now"),
		Interval:    types.StringValue("15m"),
		Index:       utils.ListValueFrom(ctx, []string{"auditbeat-*", "filebeat-*"}, types.StringType, path.Root("index"), &diags),
		Author:      utils.ListValueFrom(ctx, []string{"Security Team"}, types.StringType, path.Root("author"), &diags),
		Tags:        utils.ListValueFrom(ctx, []string{"saved-query", "detection"}, types.StringType, path.Root("tags"), &diags),
	}

	require.Empty(t, diags)

	createProps, createDiags := data.toSavedQueryRuleCreateProps(ctx, NewMockApiClient())
	require.Empty(t, createDiags)

	savedQueryRule, err := createProps.AsSecurityDetectionsAPISavedQueryRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "Test Saved Query Rule", string(savedQueryRule.Name))
	require.Equal(t, "Test saved query rule description", string(savedQueryRule.Description))
	require.Equal(t, "saved_query", string(savedQueryRule.Type))
	require.Equal(t, "my-saved-query-id", string(savedQueryRule.SavedId))
	require.Equal(t, int64(70), int64(savedQueryRule.RiskScore))
	require.Equal(t, "high", string(savedQueryRule.Severity))
}

func TestToThreatMatchRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := SecurityDetectionRuleData{
		Type:        types.StringValue("threat_match"),
		Name:        types.StringValue("Test Threat Match Rule"),
		Description: types.StringValue("Test threat match rule description"),
		Query:       types.StringValue("source.ip:*"),
		Language:    types.StringValue("kuery"),
		ThreatIndex: utils.ListValueFrom(ctx, []string{"threat-intel-*"}, types.StringType, path.Root("threat_index"), &diags),
		ThreatMapping: utils.ListValueFrom(ctx, []SecurityDetectionRuleTfDataItem{
			{
				Entries: utils.ListValueFrom(ctx, []SecurityDetectionRuleTfDataItemEntry{
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
		Index:     utils.ListValueFrom(ctx, []string{"logs-*"}, types.StringType, path.Root("index"), &diags),
	}

	require.Empty(t, diags)

	createProps, createDiags := data.toThreatMatchRuleCreateProps(ctx, NewMockApiClient())
	require.Empty(t, createDiags)

	threatMatchRule, err := createProps.AsSecurityDetectionsAPIThreatMatchRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "Test Threat Match Rule", string(threatMatchRule.Name))
	require.Equal(t, "Test threat match rule description", string(threatMatchRule.Description))
	require.Equal(t, "threat_match", string(threatMatchRule.Type))
	require.Equal(t, "source.ip:*", string(threatMatchRule.Query))
	require.Equal(t, int64(90), int64(threatMatchRule.RiskScore))
	require.Equal(t, "critical", string(threatMatchRule.Severity))
	require.Len(t, threatMatchRule.ThreatIndex, 1)
	require.Equal(t, "threat-intel-*", threatMatchRule.ThreatIndex[0])
	require.Len(t, threatMatchRule.ThreatMapping, 1)
}

func TestToThresholdRuleCreateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := SecurityDetectionRuleData{
		Type:        types.StringValue("threshold"),
		Name:        types.StringValue("Test Threshold Rule"),
		Description: types.StringValue("Test threshold rule description"),
		Query:       types.StringValue("event.action:login"),
		Language:    types.StringValue("kuery"),
		Threshold: utils.ObjectValueFrom(ctx, &ThresholdModel{
			Field:       utils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("threshold").AtName("field"), &diags),
			Value:       types.Int64Value(5),
			Cardinality: types.ListNull(getCardinalityType()),
		}, getThresholdType(), path.Root("threshold"), &diags),
		RiskScore: types.Int64Value(80),
		Severity:  types.StringValue("high"),
		Enabled:   types.BoolValue(true),
		From:      types.StringValue("now-1h"),
		To:        types.StringValue("now"),
		Interval:  types.StringValue("5m"),
		Index:     utils.ListValueFrom(ctx, []string{"auditbeat-*"}, types.StringType, path.Root("index"), &diags),
	}

	require.Empty(t, diags)

	createProps, createDiags := data.toThresholdRuleCreateProps(ctx, NewMockApiClient())
	require.Empty(t, createDiags)

	thresholdRule, err := createProps.AsSecurityDetectionsAPIThresholdRuleCreateProps()
	require.NoError(t, err)

	require.Equal(t, "Test Threshold Rule", string(thresholdRule.Name))
	require.Equal(t, "Test threshold rule description", string(thresholdRule.Description))
	require.Equal(t, "threshold", string(thresholdRule.Type))
	require.Equal(t, "event.action:login", string(thresholdRule.Query))
	require.Equal(t, int64(80), int64(thresholdRule.RiskScore))
	require.Equal(t, "high", string(thresholdRule.Severity))

	// Verify threshold configuration
	require.NotNil(t, thresholdRule.Threshold)
	require.Equal(t, int64(5), int64(thresholdRule.Threshold.Value))

	// Check single field
	singleField, err := thresholdRule.Threshold.Field.AsSecurityDetectionsAPIThresholdField0()
	require.NoError(t, err)
	require.Equal(t, "user.name", string(singleField))
}

func TestThresholdToApi(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name               string
		data               SecurityDetectionRuleData
		expectedValue      int64
		expectedFieldCount int
		hasCardinality     bool
	}{
		{
			name: "threshold with single field",
			data: SecurityDetectionRuleData{
				Threshold: utils.ObjectValueFrom(ctx, &ThresholdModel{
					Field:       utils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("threshold").AtName("field"), &diags),
					Value:       types.Int64Value(10),
					Cardinality: types.ListNull(getCardinalityType()),
				}, getThresholdType(), path.Root("threshold"), &diags),
			},
			expectedValue:      10,
			expectedFieldCount: 1,
		},
		{
			name: "threshold with multiple fields and cardinality",
			data: SecurityDetectionRuleData{
				Threshold: utils.ObjectValueFrom(ctx, &ThresholdModel{
					Field: utils.ListValueFrom(ctx, []string{"user.name", "source.ip"}, types.StringType, path.Root("threshold").AtName("field"), &diags),
					Value: types.Int64Value(5),
					Cardinality: utils.ListValueFrom(ctx, []CardinalityModel{
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
			threshold := tt.data.thresholdToApi(ctx, &diags)
			require.Empty(t, diags)
			require.NotNil(t, threshold)

			require.Equal(t, tt.expectedValue, int64(threshold.Value))

			// Check field count
			if singleField, err := threshold.Field.AsSecurityDetectionsAPIThresholdField0(); err == nil {
				require.Equal(t, 1, tt.expectedFieldCount)
				require.NotEmpty(t, string(singleField))
			} else if multipleFields, err := threshold.Field.AsSecurityDetectionsAPIThresholdField1(); err == nil {
				require.Equal(t, tt.expectedFieldCount, len(multipleFields))
			}

			if tt.hasCardinality {
				require.NotNil(t, threshold.Cardinality)
				require.NotEmpty(t, *threshold.Cardinality)
			}
		})
	}
}

func TestAlertSuppressionToApi(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name                     string
		data                     SecurityDetectionRuleData
		expectedGroupByCount     int
		hasDuration              bool
		hasMissingFieldsStrategy bool
	}{
		{
			name: "alert suppression with all fields",
			data: SecurityDetectionRuleData{
				AlertSuppression: utils.ObjectValueFrom(ctx, &AlertSuppressionModel{
					GroupBy:               utils.ListValueFrom(ctx, []string{"user.name", "source.ip"}, types.StringType, path.Root("alert_suppression").AtName("group_by"), &diags),
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
			data: SecurityDetectionRuleData{
				AlertSuppression: utils.ObjectValueFrom(ctx, &AlertSuppressionModel{
					GroupBy:               utils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("alert_suppression").AtName("group_by"), &diags),
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
			alertSuppression := tt.data.alertSuppressionToApi(ctx, &diags)
			require.Empty(t, diags)
			require.NotNil(t, alertSuppression)

			require.Equal(t, tt.expectedGroupByCount, len(alertSuppression.GroupBy))

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

func TestThreatMappingToApi(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := SecurityDetectionRuleData{
		ThreatMapping: utils.ListValueFrom(ctx, []SecurityDetectionRuleTfDataItem{
			{
				Entries: utils.ListValueFrom(ctx, []SecurityDetectionRuleTfDataItemEntry{
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

	threatMapping, threatMappingDiags := data.threatMappingToApi(ctx)
	require.Empty(t, threatMappingDiags)
	require.NotNil(t, threatMapping)
	require.Len(t, threatMapping, 1)

	mapping := threatMapping[0]
	require.Len(t, mapping.Entries, 2)

	require.Equal(t, "source.ip", string(mapping.Entries[0].Field))
	require.Equal(t, "mapping", string(mapping.Entries[0].Type))
	require.Equal(t, "threat.indicator.ip", string(mapping.Entries[0].Value))

	require.Equal(t, "user.name", string(mapping.Entries[1].Field))
	require.Equal(t, "mapping", string(mapping.Entries[1].Type))
	require.Equal(t, "threat.indicator.user.name", string(mapping.Entries[1].Value))
}

func TestActionsToApi(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Create params as a Map with string values
	paramsMap := utils.MapValueFrom(ctx, map[string]string{
		"message": "Alert triggered",
		"channel": "#security",
	}, types.StringType, path.Root("actions").AtListIndex(0).AtName("params"), &diags)

	data := SecurityDetectionRuleData{
		Actions: utils.ListValueFrom(ctx, []ActionModel{
			{
				ActionTypeId: types.StringValue(".slack"),
				Id:           types.StringValue("slack-action-1"),
				Params:       paramsMap,
				Group:        types.StringValue("default"),
				Uuid:         types.StringNull(),
				AlertsFilter: utils.MapValueFrom(ctx, map[string]attr.Value{
					"status":   types.StringValue("open"),
					"severity": types.StringValue("high"),
				}, types.StringType, path.Root("actions").AtListIndex(0).AtName("alerts_filter"), &diags),
				Frequency: utils.ObjectValueFrom(ctx, &ActionFrequencyModel{
					NotifyWhen: types.StringValue("onActionGroupChange"),
					Summary:    types.BoolValue(false),
					Throttle:   types.StringValue("1h"),
				}, getActionFrequencyType(), path.Root("actions").AtListIndex(0).AtName("frequency"), &diags),
			},
		}, getActionElementType(), path.Root("actions"), &diags),
	}

	require.Empty(t, diags)

	actions, actionsDiags := data.actionsToApi(ctx)
	require.Empty(t, actionsDiags)
	require.Len(t, actions, 1)

	action := actions[0]
	require.Equal(t, ".slack", action.ActionTypeId)
	require.Equal(t, "slack-action-1", string(action.Id))
	require.NotNil(t, action.Params)
	require.Contains(t, action.Params, "message")
	require.Equal(t, "Alert triggered", action.Params["message"])
	require.NotNil(t, action.Group)
	require.Equal(t, "default", string(*action.Group))
	require.NotNil(t, action.Frequency)
}

func TestFiltersToApi(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	filtersJSON := `[{"query": {"match": {"field": "value"}}}, {"range": {"timestamp": {"gte": "now-1h"}}}]`

	data := SecurityDetectionRuleData{
		Filters: jsontypes.NewNormalizedValue(filtersJSON),
	}

	// Test filters conversion
	filters, filtersDiags := data.filtersToApi(ctx)
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
			Group: utils.Pointer(kbapi.SecurityDetectionsAPIRuleActionGroup("default")),
			Uuid:  utils.Pointer(kbapi.SecurityDetectionsAPINonEmptyString("action-uuid-123")),
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
	require.Equal(t, ".email", action.ActionTypeId.ValueString())
	require.Equal(t, "email-action-1", action.Id.ValueString())
	require.Equal(t, "default", action.Group.ValueString())
	require.Equal(t, "action-uuid-123", action.Uuid.ValueString())
}

func TestUpdateFromRule_UnsupportedType(t *testing.T) {
	ctx := context.Background()
	data := &SecurityDetectionRuleData{}

	// Create a mock response that will fail to determine discriminator
	response := &kbapi.SecurityDetectionsAPIRuleResponse{}

	diags := data.updateFromRule(ctx, response)
	require.NotEmpty(t, diags)
	require.True(t, diags.HasError())
}

func TestUpdateFromRule(t *testing.T) {
	ctx := context.Background()
	testUUID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
	spaceId := "test-space"

	tests := []struct {
		name         string
		setupRule    func() *kbapi.SecurityDetectionsAPIRuleResponse
		expectError  bool
		errorMessage string
		validateData func(t *testing.T, data *SecurityDetectionRuleData)
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
			validateData: func(t *testing.T, data *SecurityDetectionRuleData) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceId, testUUID.String()), data.Id.ValueString())
				require.Equal(t, "test-query-rule", data.RuleId.ValueString())
				require.Equal(t, "Test Query Rule", data.Name.ValueString())
				require.Equal(t, "query", data.Type.ValueString())
				require.Equal(t, "user.name:test", data.Query.ValueString())
				require.Equal(t, "kuery", data.Language.ValueString())
				require.Equal(t, true, data.Enabled.ValueBool())
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
			validateData: func(t *testing.T, data *SecurityDetectionRuleData) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceId, testUUID.String()), data.Id.ValueString())
				require.Equal(t, "test-eql-rule", data.RuleId.ValueString())
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
			validateData: func(t *testing.T, data *SecurityDetectionRuleData) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceId, testUUID.String()), data.Id.ValueString())
				require.Equal(t, "test-esql-rule", data.RuleId.ValueString())
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
				mlJobId := kbapi.SecurityDetectionsAPIMachineLearningJobId{}
				err := mlJobId.FromSecurityDetectionsAPIMachineLearningJobId0("suspicious_activity")
				require.NoError(t, err)

				rule := kbapi.SecurityDetectionsAPIMachineLearningRule{
					Id:                   testUUID,
					RuleId:               "test-ml-rule",
					Name:                 "Test ML Rule",
					Type:                 "machine_learning",
					MachineLearningJobId: mlJobId,
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
			validateData: func(t *testing.T, data *SecurityDetectionRuleData) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceId, testUUID.String()), data.Id.ValueString())
				require.Equal(t, "test-ml-rule", data.RuleId.ValueString())
				require.Equal(t, "Test ML Rule", data.Name.ValueString())
				require.Equal(t, "machine_learning", data.Type.ValueString())
				require.Equal(t, int64(50), data.AnomalyThreshold.ValueInt64())
				require.Equal(t, int64(70), data.RiskScore.ValueInt64())
				require.Equal(t, "medium", data.Severity.ValueString())
				require.Len(t, data.MachineLearningJobId.Elements(), 1)
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
			validateData: func(t *testing.T, data *SecurityDetectionRuleData) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceId, testUUID.String()), data.Id.ValueString())
				require.Equal(t, "test-new-terms-rule", data.RuleId.ValueString())
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
			validateData: func(t *testing.T, data *SecurityDetectionRuleData) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceId, testUUID.String()), data.Id.ValueString())
				require.Equal(t, "test-saved-query-rule", data.RuleId.ValueString())
				require.Equal(t, "Test Saved Query Rule", data.Name.ValueString())
				require.Equal(t, "saved_query", data.Type.ValueString())
				require.Equal(t, "my-saved-query-id", data.SavedId.ValueString())
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
			validateData: func(t *testing.T, data *SecurityDetectionRuleData) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceId, testUUID.String()), data.Id.ValueString())
				require.Equal(t, "test-threat-match-rule", data.RuleId.ValueString())
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
			validateData: func(t *testing.T, data *SecurityDetectionRuleData) {
				require.Equal(t, fmt.Sprintf("%s/%s", spaceId, testUUID.String()), data.Id.ValueString())
				require.Equal(t, "test-threshold-rule", data.RuleId.ValueString())
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
			validateData: func(t *testing.T, data *SecurityDetectionRuleData) {
				// No validation needed for error case
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &SecurityDetectionRuleData{
				SpaceId: types.StringValue(spaceId),
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

func TestCompositeIdOperations(t *testing.T) {
	tests := []struct {
		name               string
		inputId            string
		expectedSpaceId    string
		expectedResourceId string
		shouldError        bool
	}{
		{
			name:               "valid composite id",
			inputId:            "my-space/12345678-1234-1234-1234-123456789012",
			expectedSpaceId:    "my-space",
			expectedResourceId: "12345678-1234-1234-1234-123456789012",
		},
		{
			name:        "invalid composite id format",
			inputId:     "invalid-format",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := SecurityDetectionRuleData{
				Id: types.StringValue(tt.inputId),
			}

			compId, diags := clients.CompositeIdFromStrFw(data.Id.ValueString())

			if tt.shouldError {
				require.NotEmpty(t, diags)
				return
			}

			require.Empty(t, diags)
			require.Equal(t, tt.expectedSpaceId, compId.ClusterId)
			require.Equal(t, tt.expectedResourceId, compId.ResourceId)
		})
	}
}

func TestResponseActionsToApi(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	tests := []struct {
		name        string
		data        SecurityDetectionRuleData
		actionType  string
		shouldError bool
	}{
		{
			name: "osquery response action",
			data: SecurityDetectionRuleData{
				ResponseActions: utils.ListValueFrom(ctx, []ResponseActionModel{
					{
						ActionTypeId: types.StringValue(".osquery"),
						Params: utils.ObjectValueFrom(ctx, &ResponseActionParamsModel{
							Query:        types.StringValue("SELECT * FROM processes"),
							Timeout:      types.Int64Value(300),
							EcsMapping:   types.MapNull(types.StringType),
							Queries:      types.ListNull(getOsqueryQueryElementType()),
							PackId:       types.StringNull(),
							SavedQueryId: types.StringNull(),
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
			data: SecurityDetectionRuleData{
				ResponseActions: utils.ListValueFrom(ctx, []ResponseActionModel{
					{
						ActionTypeId: types.StringValue(".endpoint"),
						Params: utils.ObjectValueFrom(ctx, &ResponseActionParamsModel{
							Command:      types.StringValue("isolate"),
							Comment:      types.StringValue("Isolating suspicious host"),
							Config:       types.ObjectNull(getEndpointProcessConfigType()),
							Query:        types.StringNull(),
							PackId:       types.StringNull(),
							SavedQueryId: types.StringNull(),
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
			data: SecurityDetectionRuleData{
				ResponseActions: utils.ListValueFrom(ctx, []ResponseActionModel{
					{
						ActionTypeId: types.StringValue(".unsupported"),
						Params: utils.ObjectValueFrom(ctx, &ResponseActionParamsModel{
							Query:        types.StringNull(),
							PackId:       types.StringNull(),
							SavedQueryId: types.StringNull(),
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
			responseActions, responseActionsDiags := tt.data.responseActionsToApi(ctx, NewMockApiClient())

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

func TestResponseActionsToApiVersionCheck(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Test data with response actions
	data := SecurityDetectionRuleData{
		ResponseActions: utils.ListValueFrom(ctx, []ResponseActionModel{
			{
				ActionTypeId: types.StringValue(".osquery"),
				Params: utils.ObjectValueFrom(ctx, &ResponseActionParamsModel{
					Query:        types.StringValue("SELECT * FROM processes"),
					Timeout:      types.Int64Value(300),
					EcsMapping:   types.MapNull(types.StringType),
					Queries:      types.ListNull(getOsqueryQueryElementType()),
					PackId:       types.StringNull(),
					SavedQueryId: types.StringNull(),
					Command:      types.StringNull(),
					Comment:      types.StringNull(),
					Config:       types.ObjectNull(getEndpointProcessConfigType()),
				}, getResponseActionParamsType(), path.Root("response_actions").AtListIndex(0).AtName("params"), &diags),
			},
		}, getResponseActionElementType(), path.Root("response_actions"), &diags),
	}

	require.Empty(t, diags)

	responseActions, responseActionsDiags := data.responseActionsToApi(ctx, NewMockApiClient())

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
			expected: utils.Pointer(kbapi.SecurityDetectionsAPIKqlQueryLanguage("kuery")),
		},
		{
			name:     "lucene language",
			language: types.StringValue("lucene"),
			expected: utils.Pointer(kbapi.SecurityDetectionsAPIKqlQueryLanguage("lucene")),
		},
		{
			name:     "unknown language defaults to kuery",
			language: types.StringValue("unknown"),
			expected: utils.Pointer(kbapi.SecurityDetectionsAPIKqlQueryLanguage("kuery")),
		},
		{
			name:     "null language returns nil",
			language: types.StringNull(),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := SecurityDetectionRuleData{
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

func TestExceptionsListToApi(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	data := SecurityDetectionRuleData{
		ExceptionsList: utils.ListValueFrom(ctx, []ExceptionsListModel{
			{
				Id:            types.StringValue("exception-1"),
				ListId:        types.StringValue("trusted-processes"),
				NamespaceType: types.StringValue("single"),
				Type:          types.StringValue("detection"),
			},
			{
				Id:            types.StringValue("exception-2"),
				ListId:        types.StringValue("allow-list"),
				NamespaceType: types.StringValue("agnostic"),
				Type:          types.StringValue("endpoint"),
			},
		}, getExceptionsListElementType(), path.Root("exceptions_list"), &diags),
	}

	require.Empty(t, diags)

	exceptionsList, exceptionsListDiags := data.exceptionsListToApi(ctx)
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
			require.Equal(t, tt.expectedFieldCount, len(thresholdModel.Field.Elements()))

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
		setupData   func() SecurityDetectionRuleData
	}{
		{
			name:     "query rule type",
			ruleType: "query",
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
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
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
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
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
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
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Type:                 types.StringValue("machine_learning"),
					Name:                 types.StringValue("Test ML Rule"),
					Description:          types.StringValue("Test description"),
					AnomalyThreshold:     types.Int64Value(50),
					MachineLearningJobId: utils.ListValueFrom(ctx, []string{"suspicious_activity"}, types.StringType, path.Root("machine_learning_job_id"), &diags),
					RiskScore:            types.Int64Value(75),
					Severity:             types.StringValue("medium"),
				}
			},
		},
		{
			name:     "new_terms rule type",
			ruleType: "new_terms",
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Type:               types.StringValue("new_terms"),
					Name:               types.StringValue("Test New Terms Rule"),
					Description:        types.StringValue("Test description"),
					Query:              types.StringValue("user.name:*"),
					NewTermsFields:     utils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("new_terms_fields"), &diags),
					HistoryWindowStart: types.StringValue("now-7d"),
					RiskScore:          types.Int64Value(75),
					Severity:           types.StringValue("medium"),
				}
			},
		},
		{
			name:     "saved_query rule type",
			ruleType: "saved_query",
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Type:        types.StringValue("saved_query"),
					Name:        types.StringValue("Test Saved Query Rule"),
					Description: types.StringValue("Test description"),
					SavedId:     types.StringValue("my-saved-query"),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
		{
			name:     "threat_match rule type",
			ruleType: "threat_match",
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Type:        types.StringValue("threat_match"),
					Name:        types.StringValue("Test Threat Match Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("source.ip:*"),
					ThreatIndex: utils.ListValueFrom(ctx, []string{"threat-intel-*"}, types.StringType, path.Root("threat_index"), &diags),
					ThreatMapping: utils.ListValueFrom(ctx, []SecurityDetectionRuleTfDataItem{
						{
							Entries: utils.ListValueFrom(ctx, []SecurityDetectionRuleTfDataItemEntry{
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
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Type:        types.StringValue("threshold"),
					Name:        types.StringValue("Test Threshold Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("event.action:login"),
					Threshold: utils.ObjectValueFrom(ctx, &ThresholdModel{
						Field:       utils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("threshold").AtName("field"), &diags),
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
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
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

			createProps, createDiags := data.toCreateProps(ctx, NewMockApiClient())

			if tt.shouldError {
				require.True(t, createDiags.HasError())
				require.Contains(t, createDiags.Errors()[0].Summary(), "Unsupported rule type")
				require.Contains(t, createDiags.Errors()[0].Detail(), tt.errorMsg)
				return
			}

			require.Empty(t, createDiags)

			// Verify that the create props can be converted to the expected rule type and check values
			switch tt.ruleType {
			case "query":
				queryRule, err := createProps.AsSecurityDetectionsAPIQueryRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Query Rule", string(queryRule.Name))
				require.Equal(t, "Test description", string(queryRule.Description))
				require.Equal(t, "query", string(queryRule.Type))
				require.Equal(t, "user.name:test", string(*queryRule.Query))
				require.Equal(t, "kuery", string(*queryRule.Language))
				require.Equal(t, int64(75), int64(queryRule.RiskScore))
				require.Equal(t, "medium", string(queryRule.Severity))
			case "eql":
				eqlRule, err := createProps.AsSecurityDetectionsAPIEqlRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test EQL Rule", string(eqlRule.Name))
				require.Equal(t, "Test description", string(eqlRule.Description))
				require.Equal(t, "eql", string(eqlRule.Type))
				require.Equal(t, "process where process.name == \"cmd.exe\"", string(eqlRule.Query))
				require.Equal(t, "eql", string(eqlRule.Language))
				require.Equal(t, int64(75), int64(eqlRule.RiskScore))
				require.Equal(t, "medium", string(eqlRule.Severity))
			case "esql":
				esqlRule, err := createProps.AsSecurityDetectionsAPIEsqlRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test ESQL Rule", string(esqlRule.Name))
				require.Equal(t, "Test description", string(esqlRule.Description))
				require.Equal(t, "esql", string(esqlRule.Type))
				require.Equal(t, "FROM logs | WHERE user.name == \"suspicious_user\"", string(esqlRule.Query))
				require.Equal(t, "esql", string(esqlRule.Language))
				require.Equal(t, int64(75), int64(esqlRule.RiskScore))
				require.Equal(t, "medium", string(esqlRule.Severity))
			case "machine_learning":
				mlRule, err := createProps.AsSecurityDetectionsAPIMachineLearningRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test ML Rule", string(mlRule.Name))
				require.Equal(t, "Test description", string(mlRule.Description))
				require.Equal(t, "machine_learning", string(mlRule.Type))
				require.Equal(t, int64(50), int64(mlRule.AnomalyThreshold))
				require.Equal(t, int64(75), int64(mlRule.RiskScore))
				require.Equal(t, "medium", string(mlRule.Severity))
				// Verify ML job ID is set correctly
				jobId, err := mlRule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId1()
				require.NoError(t, err)
				require.Equal(t, []string{"suspicious_activity"}, jobId)
			case "new_terms":
				newTermsRule, err := createProps.AsSecurityDetectionsAPINewTermsRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test New Terms Rule", string(newTermsRule.Name))
				require.Equal(t, "Test description", string(newTermsRule.Description))
				require.Equal(t, "new_terms", string(newTermsRule.Type))
				require.Equal(t, "user.name:*", string(newTermsRule.Query))
				require.Equal(t, "now-7d", string(newTermsRule.HistoryWindowStart))
				require.Equal(t, int64(75), int64(newTermsRule.RiskScore))
				require.Equal(t, "medium", string(newTermsRule.Severity))
				require.Len(t, newTermsRule.NewTermsFields, 1)
				require.Equal(t, "user.name", newTermsRule.NewTermsFields[0])
			case "saved_query":
				savedQueryRule, err := createProps.AsSecurityDetectionsAPISavedQueryRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Saved Query Rule", string(savedQueryRule.Name))
				require.Equal(t, "Test description", string(savedQueryRule.Description))
				require.Equal(t, "saved_query", string(savedQueryRule.Type))
				require.Equal(t, "my-saved-query", string(savedQueryRule.SavedId))
				require.Equal(t, int64(75), int64(savedQueryRule.RiskScore))
				require.Equal(t, "medium", string(savedQueryRule.Severity))
			case "threat_match":
				threatMatchRule, err := createProps.AsSecurityDetectionsAPIThreatMatchRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Threat Match Rule", string(threatMatchRule.Name))
				require.Equal(t, "Test description", string(threatMatchRule.Description))
				require.Equal(t, "threat_match", string(threatMatchRule.Type))
				require.Equal(t, "source.ip:*", string(threatMatchRule.Query))
				require.Equal(t, int64(75), int64(threatMatchRule.RiskScore))
				require.Equal(t, "medium", string(threatMatchRule.Severity))
				require.Len(t, threatMatchRule.ThreatIndex, 1)
				require.Equal(t, "threat-intel-*", threatMatchRule.ThreatIndex[0])
				require.Len(t, threatMatchRule.ThreatMapping, 1)
			case "threshold":
				thresholdRule, err := createProps.AsSecurityDetectionsAPIThresholdRuleCreateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Threshold Rule", string(thresholdRule.Name))
				require.Equal(t, "Test description", string(thresholdRule.Description))
				require.Equal(t, "threshold", string(thresholdRule.Type))
				require.Equal(t, "event.action:login", string(thresholdRule.Query))
				require.Equal(t, int64(75), int64(thresholdRule.RiskScore))
				require.Equal(t, "medium", string(thresholdRule.Severity))
				require.NotNil(t, thresholdRule.Threshold)
				require.Equal(t, int64(5), int64(thresholdRule.Threshold.Value))
				// Check single field
				singleField, err := thresholdRule.Threshold.Field.AsSecurityDetectionsAPIThresholdField0()
				require.NoError(t, err)
				require.Equal(t, "user.name", string(singleField))
			}
		})
	}
}

func TestToUpdateProps(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Create a valid composite ID for testing
	testUUID := uuid.New()
	testSpaceId := "test-space"
	validCompositeId := fmt.Sprintf("%s/%s", testSpaceId, testUUID.String())

	tests := []struct {
		name        string
		ruleType    string
		shouldError bool
		errorMsg    string
		setupData   func() SecurityDetectionRuleData
	}{
		{
			name:     "query rule type",
			ruleType: "query",
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Id:          types.StringValue(validCompositeId),
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
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Id:          types.StringValue(validCompositeId),
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
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Id:          types.StringValue(validCompositeId),
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
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Id:                   types.StringValue(validCompositeId),
					Type:                 types.StringValue("machine_learning"),
					Name:                 types.StringValue("Test ML Rule"),
					Description:          types.StringValue("Test description"),
					AnomalyThreshold:     types.Int64Value(50),
					MachineLearningJobId: utils.ListValueFrom(ctx, []string{"suspicious_activity"}, types.StringType, path.Root("machine_learning_job_id"), &diags),
					RiskScore:            types.Int64Value(75),
					Severity:             types.StringValue("medium"),
				}
			},
		},
		{
			name:     "new_terms rule type",
			ruleType: "new_terms",
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Id:                 types.StringValue(validCompositeId),
					Type:               types.StringValue("new_terms"),
					Name:               types.StringValue("Test New Terms Rule"),
					Description:        types.StringValue("Test description"),
					Query:              types.StringValue("user.name:*"),
					NewTermsFields:     utils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("new_terms_fields"), &diags),
					HistoryWindowStart: types.StringValue("now-7d"),
					RiskScore:          types.Int64Value(75),
					Severity:           types.StringValue("medium"),
				}
			},
		},
		{
			name:     "saved_query rule type",
			ruleType: "saved_query",
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Id:          types.StringValue(validCompositeId),
					Type:        types.StringValue("saved_query"),
					Name:        types.StringValue("Test Saved Query Rule"),
					Description: types.StringValue("Test description"),
					SavedId:     types.StringValue("my-saved-query"),
					RiskScore:   types.Int64Value(75),
					Severity:    types.StringValue("medium"),
				}
			},
		},
		{
			name:     "threat_match rule type",
			ruleType: "threat_match",
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Id:          types.StringValue(validCompositeId),
					Type:        types.StringValue("threat_match"),
					Name:        types.StringValue("Test Threat Match Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("source.ip:*"),
					ThreatIndex: utils.ListValueFrom(ctx, []string{"threat-intel-*"}, types.StringType, path.Root("threat_index"), &diags),
					ThreatMapping: utils.ListValueFrom(ctx, []SecurityDetectionRuleTfDataItem{
						{
							Entries: utils.ListValueFrom(ctx, []SecurityDetectionRuleTfDataItemEntry{
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
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Id:          types.StringValue(validCompositeId),
					Type:        types.StringValue("threshold"),
					Name:        types.StringValue("Test Threshold Rule"),
					Description: types.StringValue("Test description"),
					Query:       types.StringValue("event.action:login"),
					Threshold: utils.ObjectValueFrom(ctx, &ThresholdModel{
						Field:       utils.ListValueFrom(ctx, []string{"user.name"}, types.StringType, path.Root("threshold").AtName("field"), &diags),
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
			setupData: func() SecurityDetectionRuleData {
				return SecurityDetectionRuleData{
					Id:          types.StringValue(validCompositeId),
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

			updateProps, updateDiags := data.toUpdateProps(ctx, NewMockApiClient())

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
				require.Equal(t, "Test Query Rule", string(queryRule.Name))
				require.Equal(t, "Test description", string(queryRule.Description))
				require.Equal(t, "user.name:test", string(*queryRule.Query))
				require.Equal(t, "kuery", string(*queryRule.Language))
				require.Equal(t, int64(75), int64(queryRule.RiskScore))
				require.Equal(t, "medium", string(queryRule.Severity))
			case "eql":
				eqlRule, err := updateProps.AsSecurityDetectionsAPIEqlRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test EQL Rule", string(eqlRule.Name))
				require.Equal(t, "Test description", string(eqlRule.Description))
				require.Equal(t, "process where process.name == \"cmd.exe\"", string(eqlRule.Query))
				require.Equal(t, int64(75), int64(eqlRule.RiskScore))
				require.Equal(t, "medium", string(eqlRule.Severity))
			case "esql":
				esqlRule, err := updateProps.AsSecurityDetectionsAPIEsqlRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test ESQL Rule", string(esqlRule.Name))
				require.Equal(t, "Test description", string(esqlRule.Description))
				require.Equal(t, "FROM logs | WHERE user.name == \"suspicious_user\"", string(esqlRule.Query))
				require.Equal(t, int64(75), int64(esqlRule.RiskScore))
				require.Equal(t, "medium", string(esqlRule.Severity))
			case "machine_learning":
				mlRule, err := updateProps.AsSecurityDetectionsAPIMachineLearningRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test ML Rule", string(mlRule.Name))
				require.Equal(t, "Test description", string(mlRule.Description))
				require.Equal(t, int64(50), int64(mlRule.AnomalyThreshold))
				require.Equal(t, int64(75), int64(mlRule.RiskScore))
				require.Equal(t, "medium", string(mlRule.Severity))
				// Verify ML job ID is set correctly
				jobId, err := mlRule.MachineLearningJobId.AsSecurityDetectionsAPIMachineLearningJobId1()
				require.NoError(t, err)
				require.Equal(t, []string{"suspicious_activity"}, jobId)
			case "new_terms":
				newTermsRule, err := updateProps.AsSecurityDetectionsAPINewTermsRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test New Terms Rule", string(newTermsRule.Name))
				require.Equal(t, "Test description", string(newTermsRule.Description))
				require.Equal(t, "user.name:*", string(newTermsRule.Query))
				require.Equal(t, "now-7d", string(newTermsRule.HistoryWindowStart))
				require.Equal(t, int64(75), int64(newTermsRule.RiskScore))
				require.Equal(t, "medium", string(newTermsRule.Severity))
				require.Len(t, newTermsRule.NewTermsFields, 1)
				require.Equal(t, "user.name", newTermsRule.NewTermsFields[0])
			case "saved_query":
				savedQueryRule, err := updateProps.AsSecurityDetectionsAPISavedQueryRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Saved Query Rule", string(savedQueryRule.Name))
				require.Equal(t, "Test description", string(savedQueryRule.Description))
				require.Equal(t, "my-saved-query", string(savedQueryRule.SavedId))
				require.Equal(t, int64(75), int64(savedQueryRule.RiskScore))
				require.Equal(t, "medium", string(savedQueryRule.Severity))
			case "threat_match":
				threatMatchRule, err := updateProps.AsSecurityDetectionsAPIThreatMatchRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Threat Match Rule", string(threatMatchRule.Name))
				require.Equal(t, "Test description", string(threatMatchRule.Description))
				require.Equal(t, "source.ip:*", string(threatMatchRule.Query))
				require.Equal(t, int64(75), int64(threatMatchRule.RiskScore))
				require.Equal(t, "medium", string(threatMatchRule.Severity))
				require.Len(t, threatMatchRule.ThreatIndex, 1)
				require.Equal(t, "threat-intel-*", threatMatchRule.ThreatIndex[0])
				require.Len(t, threatMatchRule.ThreatMapping, 1)
			case "threshold":
				thresholdRule, err := updateProps.AsSecurityDetectionsAPIThresholdRuleUpdateProps()
				require.NoError(t, err)
				require.Equal(t, "Test Threshold Rule", string(thresholdRule.Name))
				require.Equal(t, "Test description", string(thresholdRule.Description))
				require.Equal(t, "event.action:login", string(thresholdRule.Query))
				require.Equal(t, int64(75), int64(thresholdRule.RiskScore))
				require.Equal(t, "medium", string(thresholdRule.Severity))
				require.NotNil(t, thresholdRule.Threshold)
				require.Equal(t, int64(5), int64(thresholdRule.Threshold.Value))
				// Check single field
				singleField, err := thresholdRule.Threshold.Field.AsSecurityDetectionsAPIThresholdField0()
				require.NoError(t, err)
				require.Equal(t, "user.name", string(singleField))
			}
		})
	}
}

func TestParseDurationToApi(t *testing.T) {
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
			result, diags := parseDurationToApi(tt.duration)

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
