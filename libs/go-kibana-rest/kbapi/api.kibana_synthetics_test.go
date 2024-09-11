package kbapi

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var (
	spaces = []string{"", "default", "testacc"}
)

func testWithPolicy(t *testing.T, client *resty.Client, namespace string, f func(policyId string)) {

	policyName := uuid.New().String()
	path := namespaceBasesPath(namespace, "/api/fleet", "/agent_policies")

	if namespace == "" {
		namespace = "default"
	}

	policyResponse, err := client.R().SetBody(map[string]interface{}{
		"name":               fmt.Sprintf("Test synthetics monitor policy %s", policyName),
		"description":        "test policy for synthetics API",
		"namespace":          namespace,
		"monitoring_enabled": []string{"logs", "metrics"},
	}).Post(path)
	assert.NoError(t, err)

	policy := struct {
		Item struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"item"`
	}{}

	err = json.Unmarshal(policyResponse.Body(), &policy)
	assert.NoError(t, err)
	defer func(policyId string) {
		client.R().SetBody(map[string]interface{}{
			"agentPolicyId": policyId,
		}).Post(fmt.Sprintf("%s/delete", path))
	}(policy.Item.Id)

	f(policy.Item.Id)
}

// TODO: test update method when set an optional parameter to `null`

func (s *KBAPITestSuite) TestKibanaSyntheticsMonitorAPI() {

	type TestConfig struct {
		config SyntheticsMonitorConfig
		fields MonitorFields
	}

	ctx := context.Background()

	for _, n := range spaces {
		testUuid := uuid.New().String()
		space := n
		syntheticsAPI := s.API.KibanaSynthetics

		testWithPolicy(s.T(), s.client, space, func(policyId string) {
			locationConfig := PrivateLocationConfig{
				Label:         fmt.Sprintf("TestKibanaSyntheticsMonitorAdd %s", testUuid),
				AgentPolicyId: policyId,
			}
			location, err := syntheticsAPI.PrivateLocation.Create(ctx, locationConfig)
			assert.NoError(s.T(), err)
			defer func(id string) {
				syntheticsAPI.PrivateLocation.Delete(ctx, id)
			}(location.Id)

			f := new(bool)
			*f = false
			t := new(bool)
			*t = true

			testCases := []struct {
				name   string
				input  TestConfig
				update TestConfig
			}{
				{
					name: "bare minimum http monitor",
					input: TestConfig{
						config: SyntheticsMonitorConfig{
							Name:             fmt.Sprintf("test synthetics http monitor %s", testUuid),
							PrivateLocations: []string{location.Label},
						},
						fields: HTTPMonitorFields{
							Url: "http://localhost:5601",
						},
					},
					update: TestConfig{
						config: SyntheticsMonitorConfig{},
						fields: HTTPMonitorFields{
							Url: "http://localhost:9200",
						},
					},
				},
				{
					name: "bare minimum tcp monitor",
					input: TestConfig{
						config: SyntheticsMonitorConfig{
							Name:             fmt.Sprintf("test synthetics tcp monitor %s", testUuid),
							PrivateLocations: []string{location.Label},
						},
						fields: TCPMonitorFields{
							Host: "localhost:5601",
						},
					},
					update: TestConfig{
						config: SyntheticsMonitorConfig{},
						fields: TCPMonitorFields{
							Host: "localhost:9200",
						},
					},
				},
				{
					name: "all fields http monitor",
					input: TestConfig{
						config: SyntheticsMonitorConfig{
							Name:             fmt.Sprintf("test all fields http monitor %s", testUuid),
							Schedule:         Every10Minutes,
							PrivateLocations: []string{location.Label},
							Enabled:          f,
							Tags:             []string{"aaa", "bbb"},
							Alert: &MonitorAlertConfig{
								Status: &SyntheticsStatusConfig{Enabled: t},
								Tls:    &SyntheticsStatusConfig{Enabled: f},
							},
							APMServiceName: "APMServiceName",
							TimeoutSeconds: 42,
							Namespace:      space,
							Params: map[string]interface{}{
								"param1": "some-params",
								"my_url": "http://localhost:8080",
							},
							RetestOnFailure: f,
						},
						fields: HTTPMonitorFields{
							Url:                   "http://localhost:5601",
							SslSupportedProtocols: []string{"TLSv1.0", "TLSv1.1", "TLSv1.2"},
							SslVerificationMode:   "full",
							MaxRedirects:          "2",
							Mode:                  ModeAny,
							Ipv4:                  t,
							Ipv6:                  f,
							Username:              "test-user-name",
							Password:              "test-password",
							ProxyHeader: map[string]interface{}{
								"User-Agent": "test",
							},
							ProxyUrl: "http://localhost",
							Response: map[string]interface{}{
								"include_body":           "always",
								"include_body_max_bytes": "1024",
							},
							Check: map[string]interface{}{
								"request": map[string]interface{}{
									"method": "POST",
									"headers": map[string]interface{}{
										"Content-Type": "application/x-www-form-urlencoded",
									},
									"body": "name=first&email=someemail%40someemailprovider.com",
								},
								"response": map[string]interface{}{
									"status": []int{200, 201},
									"body": map[string]interface{}{
										"positive": []string{"foo", "bar"},
									},
								},
							},
						},
					},
					update: TestConfig{
						config: SyntheticsMonitorConfig{
							Name:     fmt.Sprintf("update all fields http monitor %s", testUuid),
							Schedule: Every30Minutes,
						},
						fields: HTTPMonitorFields{
							Url:  "http://localhost:9200",
							Mode: ModeAll,
						},
					},
				},
				{
					name: "all fields tcp monitor",
					input: TestConfig{
						config: SyntheticsMonitorConfig{
							Name:             fmt.Sprintf("test all fields tcp monitor %s", testUuid),
							Schedule:         Every10Minutes,
							PrivateLocations: []string{location.Label},
							Enabled:          f,
							Tags:             []string{"aaa", "bbb"},
							Alert: &MonitorAlertConfig{
								Status: &SyntheticsStatusConfig{Enabled: t},
								Tls:    &SyntheticsStatusConfig{Enabled: f},
							},
							APMServiceName: "APMServiceName",
							TimeoutSeconds: 42,
							Namespace:      space,
							Params: map[string]interface{}{
								"param1": "some-params",
								"my_url": "http://localhost:8080",
							},
							RetestOnFailure: f,
						},
						fields: TCPMonitorFields{
							Host:                  "localhost:5601",
							SslSupportedProtocols: []string{"TLSv1.0", "TLSv1.1", "TLSv1.2"},
							SslVerificationMode:   "full",
							ProxyUseLocalResolver: t,
							ProxyUrl:              "http://localhost",
							CheckSend:             "Hello World",
							CheckReceive:          "Hello",
						},
					},
					update: TestConfig{
						config: SyntheticsMonitorConfig{
							Name:     fmt.Sprintf("update all fields tcp monitor %s", testUuid),
							Schedule: Every30Minutes,
						},
						fields: TCPMonitorFields{
							Host:                  "localhost:9200",
							ProxyUrl:              "http://127.0.0.1",
							ProxyUseLocalResolver: f,
						},
					},
				},
			}

			for _, tc := range testCases {
				s.Run(fmt.Sprintf("TestKibanaSyntheticsMonitorAPI ns [%s] - %s", n, tc.name), func() {
					config := tc.input.config
					fields := tc.input.fields

					monitor, err := syntheticsAPI.Monitor.Add(ctx, config, fields, space)
					assert.NoError(s.T(), err)
					assert.NotNil(s.T(), monitor)
					updateDueToKibanaAPIDiff(monitor)

					get, err := syntheticsAPI.Monitor.Get(ctx, monitor.Id, space)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), monitor, get)

					get, err = syntheticsAPI.Monitor.Get(ctx, monitor.ConfigId, space)
					assert.NoError(s.T(), err)
					assert.Equal(s.T(), monitor, get)

					update, err := syntheticsAPI.Monitor.Update(ctx, monitor.Id, tc.update.config, tc.update.fields, space)
					assert.NoError(s.T(), err)
					assert.NotNil(s.T(), update)
					updateDueToKibanaAPIDiff(update)

					get, err = syntheticsAPI.Monitor.Get(ctx, monitor.ConfigId, space)
					assert.NoError(s.T(), err)
					get.CreatedAt = time.Time{} // update response doesn't have created_at field
					assert.Equal(s.T(), update, get)

					deleted, err := syntheticsAPI.Monitor.Delete(ctx, space, monitor.ConfigId)
					assert.NoError(s.T(), err)
					for _, d := range deleted {
						assert.True(s.T(), d.Deleted)
					}

					deleted, err = syntheticsAPI.Monitor.Delete(ctx, space, monitor.Id)
					assert.NoError(s.T(), err)
					for _, d := range deleted {
						assert.False(s.T(), d.Deleted)
					}
					_, err = syntheticsAPI.Monitor.Get(ctx, monitor.Id, space)
					assert.Error(s.T(), err)
					assert.IsType(s.T(), APIError{}, err)
					assert.Equal(s.T(), 404, err.(APIError).Code)
				})
			}
		})
	}
}

// see https://github.com/elastic/kibana/issues/189906
func updateDueToKibanaAPIDiff(m *SyntheticsMonitor) {
	m.Params = nil
	m.Username = ""
	m.Password = ""
	m.ProxyHeaders = nil
	m.CheckResponseBodyPositive = nil
	m.CheckRequestBody = nil
	m.CheckRequestHeaders = nil
	m.CheckSend = ""
	m.CheckReceive = ""
}

func (s *KBAPITestSuite) TestKibanaSyntheticsPrivateLocationAPI() {

	ctx := context.Background()

	for _, n := range spaces {
		testUuid := uuid.New().String()
		namespace := n
		pAPI := s.API.KibanaSynthetics.PrivateLocation

		s.Run(fmt.Sprintf("TestKibanaSyntheticsPrivateLocationAPI - %s", n), func() {
			testWithPolicy(s.T(), s.client, namespace, func(policyId string) {

				cfg := PrivateLocationConfig{
					Label:         fmt.Sprintf("TestKibanaSyntheticsPrivateLocationAPI-%s", testUuid),
					AgentPolicyId: policyId,
					Tags:          []string{"a", "b"},
					Geo: &SyntheticGeoConfig{
						Lat: 12.12,
						Lon: -42.42,
					},
				}
				created, err := pAPI.Create(ctx, cfg)

				assert.NoError(s.T(), err)
				assert.Equal(s.T(), created.Label, cfg.Label)
				assert.Equal(s.T(), created.AgentPolicyId, cfg.AgentPolicyId)

				get, err := pAPI.Get(ctx, created.Id)
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), created, get)

				get, err = pAPI.Get(ctx, created.Label)
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), created, get)

				err = pAPI.Delete(ctx, created.Id)
				assert.NoError(s.T(), err)

				_, err = pAPI.Get(ctx, created.Id)
				assert.Error(s.T(), err)
			})
		})
	}
}

func (s *KBAPITestSuite) TestKibanaSyntheticsPrivateLocationNotFound() {
	testUuid := uuid.New().String()
	pAPI := s.API.KibanaSynthetics.PrivateLocation

	ids := []string{"", "not-found", testUuid}
	ctx := context.Background()

	for _, id := range ids {
		s.Run(fmt.Sprintf("TestKibanaSyntheticsPrivateLocationNotFound - %s", id), func() {
			_, err := pAPI.Get(ctx, id)
			assert.Error(s.T(), err)
			assert.IsType(s.T(), APIError{}, err)
			assert.Equal(s.T(), 404, err.(APIError).Code)
		})
	}
}
