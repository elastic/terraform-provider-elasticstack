package kbapi

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

/*
TODO: monitor update
TODO: different input params for monitors
*/

var (
	namespaces = []string{"", "default", "testacc"}
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

func (s *KBAPITestSuite) TestKibanaSyntheticsMonitorAPI() {

	for _, n := range namespaces {
		testUuid := uuid.New().String()
		space := n
		syntheticsAPI := s.API.KibanaSynthetics

		s.Run(fmt.Sprintf("TestKibanaSyntheticsMonitorAPI - %s", n), func() {
			testWithPolicy(s.T(), s.client, space, func(policyId string) {
				locationConfig := PrivateLocationConfig{
					Label:         fmt.Sprintf("TestKibanaSyntheticsMonitorAdd %s", testUuid),
					AgentPolicyId: policyId,
				}
				location, err := syntheticsAPI.PrivateLocation.Create(locationConfig, space)
				assert.NoError(s.T(), err)
				defer func(id string) {
					syntheticsAPI.PrivateLocation.Delete(id, space)
				}(location.Id)

				config := SyntheticsMonitorConfig{
					Name:             fmt.Sprintf("test synthetics monitor %s", testUuid),
					Type:             Http,
					PrivateLocations: []string{location.Label},
				}
				fields := HTTPMonitorFields{
					Url: "http://localhost:5601",
				}
				monitor, err := syntheticsAPI.Monitor.Add(config, fields, space)
				assert.NoError(s.T(), err)
				assert.NotNil(s.T(), monitor)

				get, err := syntheticsAPI.Monitor.Get(monitor.Id, space)
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), monitor, get)

				get, err = syntheticsAPI.Monitor.Get(monitor.ConfigId, space)
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), monitor, get)

				deleted, err := syntheticsAPI.Monitor.Delete(space, monitor.ConfigId)
				assert.NoError(s.T(), err)
				for _, d := range deleted {
					assert.True(s.T(), d.Deleted)
				}

				deleted, err = syntheticsAPI.Monitor.Delete(space, monitor.Id)
				assert.NoError(s.T(), err)
				for _, d := range deleted {
					assert.False(s.T(), d.Deleted)
				}
			})
		})
	}
}

func (s *KBAPITestSuite) TestKibanaSyntheticsPrivateLocationAPI() {

	for _, n := range namespaces {
		testUuid := uuid.New().String()
		space := n
		pAPI := s.API.KibanaSynthetics.PrivateLocation

		s.Run(fmt.Sprintf("TestKibanaSyntheticsPrivateLocationAPI - %s", n), func() {
			testWithPolicy(s.T(), s.client, space, func(policyId string) {

				cfg := PrivateLocationConfig{
					Label:         fmt.Sprintf("TestKibanaSyntheticsPrivateLocationAPI %s", testUuid),
					AgentPolicyId: policyId,
					Tags:          []string{"a", "b"},
					Geo: Geo{
						Lat: 12.12,
						Lon: -42.42,
					},
				}
				created, err := pAPI.Create(cfg, space)
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), created.Label, cfg.Label)
				assert.Equal(s.T(), created.AgentPolicyId, cfg.AgentPolicyId)

				get, err := pAPI.Get(created.Id, space)
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), created, get)

				get, err = pAPI.Get(created.Label, space)
				assert.NoError(s.T(), err)
				assert.Equal(s.T(), created, get)

				err = pAPI.Delete(created.Id, space)
				assert.NoError(s.T(), err)

				_, err = pAPI.Get(created.Id, space)
				assert.Error(s.T(), err)
			})
		})
	}
}
