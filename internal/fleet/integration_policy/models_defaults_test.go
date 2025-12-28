package integration_policy

import (
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/elastic/terraform-provider-elasticstack/generated/kbapi"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/integration_kafka.json
var kafkaIntegrationJSON []byte

//go:embed testdata/integration_gcp_vertexai.json
var gcpVertexAIIntegrationJSON []byte

func TestApiVarsDefaults(t *testing.T) {
	tests := []struct {
		name           string
		vars           apiVars
		expectedJSON   string
		expectError    bool
		expectedIsNull bool
	}{
		{
			name:           "nil vars returns empty object",
			vars:           nil,
			expectedJSON:   "{}",
			expectedIsNull: false,
		},
		{
			name:           "empty vars returns empty object",
			vars:           apiVars{},
			expectedJSON:   "{}",
			expectedIsNull: false,
		},
		{
			name: "single var with default",
			vars: apiVars{
				{
					Name:    "hosts",
					Default: []interface{}{"http://127.0.0.1:8778"},
				},
			},
			expectedJSON: `{"hosts":["http://127.0.0.1:8778"]}`,
		},
		{
			name: "single var without default is omitted",
			vars: apiVars{
				{
					Name:    "username",
					Default: nil,
				},
			},
			expectedJSON: "{}",
		},
		{
			name: "multiple vars with mixed defaults",
			vars: apiVars{
				{
					Name:    "hosts",
					Default: []interface{}{"localhost:9092"},
				},
				{
					Name:    "period",
					Default: "10s",
				},
				{
					Name:    "username",
					Default: nil,
				},
				{
					Name:    "ssl.verification_mode",
					Default: "none",
				},
			},
			expectedJSON: `{"hosts":["localhost:9092"],"period":"10s","ssl.verification_mode":"none"}`,
		},
		{
			name: "var with complex default value",
			vars: apiVars{
				{
					Name: "headers",
					Default: `# headers:
#   Cookie: abcdef=123456
#   My-Custom-Header: my-custom-value
`,
				},
			},
			expectedJSON: `{"headers":"# headers:\n#   Cookie: abcdef=123456\n#   My-Custom-Header: my-custom-value\n"}`,
		},
		{
			name: "var with boolean default",
			vars: apiVars{
				{
					Name:    "preserve_original_event",
					Default: false,
				},
			},
			expectedJSON: `{"preserve_original_event":false}`,
		},
		{
			name: "var with string array default",
			vars: apiVars{
				{
					Name:    "tags",
					Default: []interface{}{"kafka-log"},
				},
			},
			expectedJSON: `{"tags":["kafka-log"]}`,
		},
		{
			name: "var with multi-element array default",
			vars: apiVars{
				{
					Name: "paths",
					Default: []interface{}{
						"/logs/controller.log*",
						"/logs/server.log*",
						"/logs/state-change.log*",
					},
				},
			},
			expectedJSON: `{"paths":["/logs/controller.log*","/logs/server.log*","/logs/state-change.log*"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.vars.defaults()

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
			} else {
				require.False(t, diags.HasError(), "Expected no error but got: %v", diags)

				if tt.expectedIsNull {
					assert.True(t, result.IsNull(), "Expected null result")
				} else {
					assert.False(t, result.IsNull(), "Expected non-null result")

					// Normalize JSON for comparison
					var expectedMap map[string]interface{}
					err := json.Unmarshal([]byte(tt.expectedJSON), &expectedMap)
					require.NoError(t, err, "Failed to unmarshal expected JSON")

					var actualMap map[string]interface{}
					err = json.Unmarshal([]byte(result.ValueString()), &actualMap)
					require.NoError(t, err, "Failed to unmarshal actual JSON")

					assert.Equal(t, expectedMap, actualMap, "JSON content mismatch")
				}
			}
		})
	}
}

func TestApiPolicyTemplateDefaults(t *testing.T) {
	tests := []struct {
		name         string
		templates    apiPolicyTemplates
		expectedKeys []string
		expectError  bool
	}{
		{
			name:         "nil templates returns empty map",
			templates:    nil,
			expectedKeys: []string{},
		},
		{
			name:         "empty templates returns empty map",
			templates:    apiPolicyTemplates{},
			expectedKeys: []string{},
		},
		{
			name: "template with no inputs returns empty map",
			templates: apiPolicyTemplates{
				{
					Inputs: []apiPolicyTemplateInput{},
				},
			},
			expectedKeys: []string{},
		},
		{
			name: "template with single input",
			templates: apiPolicyTemplates{
				{
					Name: "kafka",
					Inputs: []apiPolicyTemplateInput{
						{
							Type: "jolokia/metrics",
							Vars: apiVars{
								{
									Name:    "hosts",
									Default: []interface{}{"http://127.0.0.1:8778"},
								},
							},
						},
					},
				},
			},
			expectedKeys: []string{"kafka-jolokia/metrics"},
		},
		{
			name: "template with multiple input types",
			templates: apiPolicyTemplates{
				{
					Name: "kafka",
					Inputs: []apiPolicyTemplateInput{
						{
							Type: "jolokia/metrics",
							Vars: apiVars{
								{
									Name:    "hosts",
									Default: []interface{}{"http://127.0.0.1:8778"},
								},
							},
						},
						{
							Type: "logfile",
							Vars: apiVars{},
						},
						{
							Type: "kafka/metrics",
							Vars: apiVars{
								{
									Name:    "hosts",
									Default: []interface{}{"localhost:9092"},
								},
								{
									Name:    "period",
									Default: "10s",
								},
							},
						},
					},
				},
			},
			expectedKeys: []string{"kafka-jolokia/metrics", "kafka-logfile", "kafka-kafka/metrics"},
		},
		{
			name: "multiple templates",
			templates: apiPolicyTemplates{
				{
					Name: "kafka",
					Inputs: []apiPolicyTemplateInput{
						{
							Type: "jolokia/metrics",
							Vars: apiVars{},
						},
					},
				},
				{
					Name: "nginx",
					Inputs: []apiPolicyTemplateInput{
						{
							Type: "access",
							Vars: apiVars{},
						},
					},
				},
			},
			expectedKeys: []string{"kafka-jolokia/metrics", "nginx-access"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.templates.defaults()

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
			} else {
				require.False(t, diags.HasError(), "Expected no error but got: %v", diags)
				assert.Len(t, result, len(tt.expectedKeys), "Unexpected number of input types")

				for _, key := range tt.expectedKeys {
					assert.Contains(t, result, key, "Expected key %s not found", key)
					assert.False(t, result[key].IsNull(), "Expected non-null value for key %s", key)
				}
			}
		})
	}
}

func TestApiDatastreamsDefaults(t *testing.T) {
	tests := []struct {
		name                 string
		datastreams          apiDatastreams
		expectedInputKeys    []string
		expectedStreamCounts map[string]int
		expectError          bool
	}{
		{
			name:                 "nil datastreams returns empty map",
			datastreams:          nil,
			expectedInputKeys:    []string{},
			expectedStreamCounts: map[string]int{},
		},
		{
			name:                 "empty datastreams returns empty map",
			datastreams:          apiDatastreams{},
			expectedInputKeys:    []string{},
			expectedStreamCounts: map[string]int{},
		},
		{
			name: "single datastream with single stream",
			datastreams: apiDatastreams{
				{
					Type:    "metrics",
					Dataset: "kafka.broker",
					Streams: []apiDatastreamStream{
						{
							Input:   "kafka/metrics",
							Enabled: true,
							Vars: apiVars{
								{
									Name:    "jolokia_hosts",
									Default: []interface{}{"localhost:8778"},
								},
							},
						},
					},
				},
			},
			expectedInputKeys: []string{"kafka/metrics"},
			expectedStreamCounts: map[string]int{
				"kafka/metrics": 1,
			},
		},
		{
			name: "multiple datastreams with different inputs",
			datastreams: apiDatastreams{
				{
					Type:    "metrics",
					Dataset: "kafka.consumer",
					Streams: []apiDatastreamStream{
						{
							Input:   "jolokia/metrics",
							Enabled: false,
							Vars: apiVars{
								{
									Name:    "period",
									Default: "60s",
								},
							},
						},
					},
				},
				{
					Type:    "logs",
					Dataset: "kafka.log",
					Streams: []apiDatastreamStream{
						{
							Input:   "logfile",
							Enabled: true,
							Vars: apiVars{
								{
									Name:    "kafka_home",
									Default: "/opt/kafka*",
								},
							},
						},
					},
				},
			},
			expectedInputKeys: []string{"jolokia/metrics", "logfile"},
			expectedStreamCounts: map[string]int{
				"jolokia/metrics": 1,
				"logfile":         1,
			},
		},
		{
			name: "multiple streams for same input",
			datastreams: apiDatastreams{
				{
					Type:    "metrics",
					Dataset: "kafka.broker",
					Streams: []apiDatastreamStream{
						{
							Input:   "kafka/metrics",
							Enabled: true,
							Vars:    apiVars{},
						},
					},
				},
				{
					Type:    "metrics",
					Dataset: "kafka.partition",
					Streams: []apiDatastreamStream{
						{
							Input:   "kafka/metrics",
							Enabled: true,
							Vars:    apiVars{},
						},
					},
				},
				{
					Type:    "metrics",
					Dataset: "kafka.consumergroup",
					Streams: []apiDatastreamStream{
						{
							Input:   "kafka/metrics",
							Enabled: true,
							Vars:    apiVars{},
						},
					},
				},
			},
			expectedInputKeys: []string{"kafka/metrics"},
			expectedStreamCounts: map[string]int{
				"kafka/metrics": 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := tt.datastreams.defaults()

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
			} else {
				require.False(t, diags.HasError(), "Expected no error but got: %v", diags)
				assert.Len(t, result, len(tt.expectedInputKeys), "Unexpected number of input types")

				for _, inputKey := range tt.expectedInputKeys {
					assert.Contains(t, result, inputKey, "Expected input key %s not found", inputKey)

					streams := result[inputKey]
					expectedCount := tt.expectedStreamCounts[inputKey]
					assert.Len(t, streams, expectedCount, "Unexpected stream count for input %s", inputKey)
				}
			}
		})
	}
}

func TestApiDatastreamsDefaults_StreamProperties(t *testing.T) {
	datastreams := apiDatastreams{
		{
			Type:    "metrics",
			Dataset: "kafka.consumer",
			Streams: []apiDatastreamStream{
				{
					Input:   "jolokia/metrics",
					Enabled: false,
					Vars: apiVars{
						{
							Name:    "jolokia_hosts",
							Default: []interface{}{"localhost:8774"},
						},
						{
							Name:    "period",
							Default: "60s",
						},
					},
				},
			},
		},
	}

	result, diags := datastreams.defaults()
	require.False(t, diags.HasError(), "Expected no error but got: %v", diags)
	require.Contains(t, result, "jolokia/metrics", "Expected jolokia/metrics input")

	streams := result["jolokia/metrics"]
	require.Contains(t, streams, "kafka.consumer", "Expected kafka.consumer stream")

	stream := streams["kafka.consumer"]
	assert.Equal(t, types.BoolValue(false), stream.Enabled, "Stream enabled mismatch")
	assert.False(t, stream.Vars.IsNull(), "Expected non-null vars")

	// Verify vars content
	var varsMap map[string]interface{}
	err := json.Unmarshal([]byte(stream.Vars.ValueString()), &varsMap)
	require.NoError(t, err, "Failed to unmarshal vars")

	assert.Equal(t, []interface{}{"localhost:8774"}, varsMap["jolokia_hosts"])
	assert.Equal(t, "60s", varsMap["period"])
}

func TestPackageInfoToDefaults(t *testing.T) {
	tests := []struct {
		name              string
		pkg               *kbapi.PackageInfo
		expectedInputKeys []string
		expectError       bool
	}{
		{
			name:              "nil package returns empty map",
			pkg:               nil,
			expectedInputKeys: []string{},
		},
		{
			name:              "package with no policy templates or datastreams",
			pkg:               &kbapi.PackageInfo{},
			expectedInputKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, diags := packageInfoToDefaults(tt.pkg)

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
			} else {
				require.False(t, diags.HasError(), "Expected no error but got: %v", diags)
				assert.Len(t, result, len(tt.expectedInputKeys), "Unexpected number of input types")

				for _, key := range tt.expectedInputKeys {
					assert.Contains(t, result, key, "Expected input key %s not found", key)
				}
			}
		})
	}
}

func TestPackageInfoToDefaults_Kafka(t *testing.T) {
	// Load the actual Kafka package JSON
	var wrapper struct {
		Item kbapi.PackageInfo `json:"item"`
	}
	err := json.Unmarshal(kafkaIntegrationJSON, &wrapper)
	require.NoError(t, err, "Failed to unmarshal Kafka integration JSON")

	pkg := &wrapper.Item

	// Test with the full Kafka package
	result, diags := packageInfoToDefaults(pkg)
	require.False(t, diags.HasError(), "Expected no error but got: %v", diags)

	// Verify expected input types exist
	expectedInputs := []string{"kafka-jolokia/metrics", "kafka-logfile", "kafka-kafka/metrics"}
	for _, inputType := range expectedInputs {
		assert.Contains(t, result, inputType, "Expected input type %s not found", inputType)
	}

	// Verify jolokia/metrics input has expected vars
	jolokiaInput := result["kafka-jolokia/metrics"]
	assert.False(t, jolokiaInput.Vars.IsNull(), "Expected non-null vars for jolokia/metrics")

	var jolokiaVars map[string]interface{}
	err = json.Unmarshal([]byte(jolokiaInput.Vars.ValueString()), &jolokiaVars)
	require.NoError(t, err, "Failed to unmarshal jolokia/metrics vars")

	// Check some specific defaults from the Kafka package
	assert.Contains(t, jolokiaVars, "hosts", "Expected 'hosts' var")
	assert.Contains(t, jolokiaVars, "metrics_path", "Expected 'metrics_path' var")
	assert.Contains(t, jolokiaVars, "http_method", "Expected 'http_method' var")
	assert.Contains(t, jolokiaVars, "ssl.verification_mode", "Expected 'ssl.verification_mode' var")

	assert.Equal(t, []interface{}{"http://127.0.0.1:8778"}, jolokiaVars["hosts"])
	assert.Equal(t, "/jolokia", jolokiaVars["metrics_path"])
	assert.Equal(t, "GET", jolokiaVars["http_method"])
	assert.Equal(t, "none", jolokiaVars["ssl.verification_mode"])

	// Verify kafka/metrics input has expected vars
	kafkaInput := result["kafka-kafka/metrics"]
	assert.False(t, kafkaInput.Vars.IsNull(), "Expected non-null vars for kafka-kafka/metrics")

	var kafkaVars map[string]interface{}
	err = json.Unmarshal([]byte(kafkaInput.Vars.ValueString()), &kafkaVars)
	require.NoError(t, err, "Failed to unmarshal kafka-kafka/metrics vars")
	assert.Contains(t, kafkaVars, "hosts", "Expected 'hosts' var")
	assert.Contains(t, kafkaVars, "period", "Expected 'period' var")

	assert.Equal(t, []interface{}{"localhost:9092"}, kafkaVars["hosts"])
	assert.Equal(t, "10s", kafkaVars["period"])

	// Verify streams are populated correctly
	assert.NotNil(t, jolokiaInput.Streams, "Expected streams for jolokia/metrics")
	assert.NotEmpty(t, jolokiaInput.Streams, "Expected non-empty streams for jolokia/metrics")

	// Check specific stream - kafka.consumer
	consumerStream, ok := jolokiaInput.Streams["kafka.consumer"]
	require.True(t, ok, "Expected kafka.consumer stream")
	assert.Equal(t, types.BoolValue(false), consumerStream.Enabled, "kafka.consumer should be disabled by default")

	var consumerVars map[string]interface{}
	err = json.Unmarshal([]byte(consumerStream.Vars.ValueString()), &consumerVars)
	require.NoError(t, err, "Failed to unmarshal kafka.consumer vars")

	assert.Contains(t, consumerVars, "jolokia_hosts", "Expected 'jolokia_hosts' var in stream")
	assert.Contains(t, consumerVars, "period", "Expected 'period' var in stream")

	assert.Equal(t, []interface{}{"localhost:8774"}, consumerVars["jolokia_hosts"])
	assert.Equal(t, "60s", consumerVars["period"])

	// Verify logfile input and log stream
	logfileInput := result["kafka-logfile"]
	assert.NotNil(t, logfileInput.Streams, "Expected streams for logfile")

	logStream, ok := logfileInput.Streams["kafka.log"]
	require.True(t, ok, "Expected kafka.log stream")
	assert.Equal(t, types.BoolValue(true), logStream.Enabled, "kafka.log should be enabled by default")

	var logVars map[string]interface{}
	err = json.Unmarshal([]byte(logStream.Vars.ValueString()), &logVars)
	require.NoError(t, err, "Failed to unmarshal kafka.log vars")

	assert.Contains(t, logVars, "kafka_home", "Expected 'kafka_home' var")
	assert.Contains(t, logVars, "paths", "Expected 'paths' var")
	assert.Contains(t, logVars, "tags", "Expected 'tags' var")
	assert.Contains(t, logVars, "preserve_original_event", "Expected 'preserve_original_event' var")

	assert.Equal(t, "/opt/kafka*", logVars["kafka_home"])
	assert.Equal(t, []interface{}{
		"/logs/controller.log*",
		"/logs/server.log*",
		"/logs/state-change.log*",
		"/logs/kafka-*.log*",
	}, logVars["paths"])
	assert.Equal(t, []interface{}{"kafka-log"}, logVars["tags"])
	assert.Equal(t, false, logVars["preserve_original_event"])

	// Verify kafka/metrics streams
	kafkaStreams := kafkaInput.Streams
	assert.Contains(t, kafkaStreams, "kafka.broker", "Expected kafka.broker stream")
	assert.Contains(t, kafkaStreams, "kafka.partition", "Expected kafka.partition stream")
	assert.Contains(t, kafkaStreams, "kafka.consumergroup", "Expected kafka.consumergroup stream")

	// Verify all streams have expected enabled state
	assert.Equal(t, types.BoolValue(true), kafkaStreams["kafka.broker"].Enabled)
	assert.Equal(t, types.BoolValue(true), kafkaStreams["kafka.partition"].Enabled)
	assert.Equal(t, types.BoolValue(true), kafkaStreams["kafka.consumergroup"].Enabled)
}

func TestPolicyTemplateAndDataStreamsFromPackageInfo(t *testing.T) {
	tests := []struct {
		name                   string
		pkg                    *kbapi.PackageInfo
		expectedPolicyTemplate bool
		expectedDataStreams    bool
		expectError            bool
	}{
		{
			name:                   "nil package returns nil",
			pkg:                    nil,
			expectedPolicyTemplate: false,
			expectedDataStreams:    false,
		},
		{
			name:                   "empty package returns nil",
			pkg:                    &kbapi.PackageInfo{},
			expectedPolicyTemplate: false,
			expectedDataStreams:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policyTemplate, datastreams, diags := policyTemplateAndDataStreamsFromPackageInfo(tt.pkg)

			if tt.expectError {
				require.True(t, diags.HasError(), "Expected error but got none")
			} else {
				require.False(t, diags.HasError(), "Expected no error but got: %v", diags)

				if tt.expectedPolicyTemplate {
					assert.NotNil(t, policyTemplate, "Expected non-nil policy template")
				} else {
					assert.Nil(t, policyTemplate, "Expected nil policy template")
				}

				if tt.expectedDataStreams {
					assert.NotNil(t, datastreams, "Expected non-nil datastreams")
					assert.NotEmpty(t, datastreams, "Expected non-empty datastreams")
				}
			}
		})
	}
}

func TestPolicyTemplateAndDataStreamsFromPackageInfo_Kafka(t *testing.T) {
	// Load the actual Kafka package JSON
	var wrapper struct {
		Item kbapi.PackageInfo `json:"item"`
	}
	err := json.Unmarshal(kafkaIntegrationJSON, &wrapper)
	require.NoError(t, err, "Failed to unmarshal Kafka integration JSON")

	pkg := &wrapper.Item

	policyTemplates, datastreams, diags := policyTemplateAndDataStreamsFromPackageInfo(pkg)
	require.False(t, diags.HasError(), "Expected no error but got: %v", diags)

	// Verify policy template was extracted
	require.Len(t, policyTemplates, 1, "Expected 1 policy template")
	policyTemplate := policyTemplates[0]
	assert.Len(t, policyTemplate.Inputs, 3, "Expected 3 input types in policy template")

	// Verify input types
	inputTypes := make([]string, 0, len(policyTemplate.Inputs))
	for _, input := range policyTemplate.Inputs {
		inputTypes = append(inputTypes, input.Type)
	}
	assert.Contains(t, inputTypes, "jolokia/metrics")
	assert.Contains(t, inputTypes, "logfile")
	assert.Contains(t, inputTypes, "kafka/metrics")

	// Verify datastreams were extracted
	require.NotNil(t, datastreams, "Expected non-nil datastreams")
	assert.Len(t, datastreams, 13, "Expected 13 datastreams in Kafka package")

	// Verify some specific datastreams
	datasetNames := make([]string, 0, len(datastreams))
	for _, ds := range datastreams {
		datasetNames = append(datasetNames, ds.Dataset)
	}

	expectedDatasets := []string{
		"kafka.broker",
		"kafka.consumer",
		"kafka.consumergroup",
		"kafka.controller",
		"kafka.jvm",
		"kafka.log",
		"kafka.log_manager",
		"kafka.network",
		"kafka.partition",
		"kafka.producer",
		"kafka.raft",
		"kafka.replica_manager",
		"kafka.topic",
	}

	for _, expected := range expectedDatasets {
		assert.Contains(t, datasetNames, expected, "Expected dataset %s not found", expected)
	}

	// Verify a specific datastream has correct structure
	var logDatastream *apiDatastream
	for i := range datastreams {
		if datastreams[i].Dataset == "kafka.log" {
			logDatastream = &datastreams[i]
			break
		}
	}

	require.NotNil(t, logDatastream, "Expected to find kafka.log datastream")
	assert.Equal(t, "logs", logDatastream.Type)
	assert.Len(t, logDatastream.Streams, 1, "Expected 1 stream in kafka.log datastream")
	assert.Equal(t, "logfile", logDatastream.Streams[0].Input)
	assert.True(t, logDatastream.Streams[0].Enabled)
}

func TestPolicyTemplateAndDataStreamsFromPackageInfo_GCP_VertexAI(t *testing.T) {
	// Load the actual Kafka package JSON
	var wrapper struct {
		Item kbapi.PackageInfo `json:"item"`
	}
	err := json.Unmarshal(gcpVertexAIIntegrationJSON, &wrapper)
	require.NoError(t, err, "Failed to unmarshal GCP Vertex AI integration JSON")

	pkg := &wrapper.Item

	policyTemplates, datastreams, diags := policyTemplateAndDataStreamsFromPackageInfo(pkg)
	require.False(t, diags.HasError(), "Expected no error but got: %v", diags)

	// Verify policy template was extracted
	require.Len(t, policyTemplates, 2, "Expected 2 policy templates")

	// Template 1: Metrics
	metricsTemplate := policyTemplates[0]
	assert.Equal(t, "GCP Vertex AI Metrics", metricsTemplate.Name)
	assert.Len(t, metricsTemplate.Inputs, 1, "Expected 1 input type in metrics policy template")
	assert.Equal(t, "gcp/metrics", metricsTemplate.Inputs[0].Type)

	// Template 2: Logs
	logsTemplate := policyTemplates[1]
	assert.Equal(t, "GCP Vertex AI  Logs", logsTemplate.Name)
	assert.Len(t, logsTemplate.Inputs, 2, "Expected 2 input types in logs policy template")

	logInputTypes := make([]string, 0, len(logsTemplate.Inputs))
	for _, input := range logsTemplate.Inputs {
		logInputTypes = append(logInputTypes, input.Type)
	}
	assert.Contains(t, logInputTypes, "gcp/metrics")
	assert.Contains(t, logInputTypes, "gcp-pubsub")

	// Verify datastreams were extracted
	require.NotNil(t, datastreams, "Expected non-nil datastreams")
	assert.Len(t, datastreams, 3, "Expected 3 datastreams in GCP Vertex AI package")

	// Verify some specific datastreams
	datasetNames := make([]string, 0, len(datastreams))
	for _, ds := range datastreams {
		datasetNames = append(datasetNames, ds.Dataset)
	}

	expectedDatasets := []string{
		"gcp_vertexai.auditlogs",
		"gcp_vertexai.metrics",
		"gcp_vertexai.prompt_response_logs",
	}

	for _, expected := range expectedDatasets {
		assert.Contains(t, datasetNames, expected, "Expected dataset %s not found", expected)
	}

	// Verify a specific datastream has correct structure
	var auditDatastream *apiDatastream
	for i := range datastreams {
		if datastreams[i].Dataset == "gcp_vertexai.auditlogs" {
			auditDatastream = &datastreams[i]
			break
		}
	}

	require.NotNil(t, auditDatastream, "Expected to find gcp_vertexai.auditlogs datastream")
	assert.Equal(t, "logs", auditDatastream.Type)
	assert.Len(t, auditDatastream.Streams, 1, "Expected 1 stream in gcp_vertexai.auditlogs datastream")
	assert.Equal(t, "gcp-pubsub", auditDatastream.Streams[0].Input)
	assert.True(t, auditDatastream.Streams[0].Enabled)
}

func TestInputDefaultsModel_Integration(t *testing.T) {
	// This test ensures that the defaults model structure correctly represents
	// the data needed for integration policies

	defaults := map[string]inputDefaultsModel{
		"kafka/metrics": {
			Vars: jsontypes.NewNormalizedValue(`{"hosts":["localhost:9092"],"period":"10s"}`),
			Streams: map[string]inputDefaultsStreamModel{
				"kafka.broker": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{"jolokia_hosts":["localhost:8778"]}`),
				},
				"kafka.partition": {
					Enabled: types.BoolValue(true),
					Vars:    jsontypes.NewNormalizedValue(`{}`),
				},
			},
		},
	}

	// Verify structure
	assert.Contains(t, defaults, "kafka/metrics")

	kafkaDefaults := defaults["kafka/metrics"]
	assert.False(t, kafkaDefaults.Vars.IsNull())
	assert.Len(t, kafkaDefaults.Streams, 2)

	brokerStream := kafkaDefaults.Streams["kafka.broker"]
	assert.Equal(t, types.BoolValue(true), brokerStream.Enabled)
	assert.False(t, brokerStream.Vars.IsNull())
}
