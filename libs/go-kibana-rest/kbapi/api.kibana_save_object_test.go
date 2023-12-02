package kbapi

import (
	"encoding/json"

	"github.com/stretchr/testify/assert"
)

func (s *KBAPITestSuite) TestKibanaSaveObject() {

	// Create new index pattern
	dataJSON := `{"attributes": {"title": "test-pattern-*"}}`
	data := make(map[string]interface{})
	err := json.Unmarshal([]byte(dataJSON), &data)
	if err != nil {
		panic(err)
	}
	resp, err := s.API.KibanaSavedObject.Create(data, "index-pattern", "test", true, "default")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	assert.Equal(s.T(), "test", resp["id"])
	assert.Equal(s.T(), "test-pattern-*", resp["attributes"].(map[string]interface{})["title"])

	// Create new index pattern in space
	dataJSON = `{"attributes": {"title": "test-pattern-*"}}`
	data = make(map[string]interface{})
	err = json.Unmarshal([]byte(dataJSON), &data)
	if err != nil {
		panic(err)
	}
	resp, err = s.API.KibanaSavedObject.Create(data, "index-pattern", "test2", true, "testacc")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	assert.Equal(s.T(), "test2", resp["id"])
	assert.Equal(s.T(), "test-pattern-*", resp["attributes"].(map[string]interface{})["title"])

	// Get index pattern
	resp, err = s.API.KibanaSavedObject.Get("index-pattern", "test", "default")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	assert.Equal(s.T(), "test", resp["id"])

	// Get index pattern from space
	resp, err = s.API.KibanaSavedObject.Get("index-pattern", "test2", "testacc")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	assert.Equal(s.T(), "test2", resp["id"])

	// Search index pattern
	parameters := &OptionalFindParameters{
		Search:       "test*",
		SearchFields: []string{"title"},
		Fields:       []string{"id"},
	}
	resp, err = s.API.KibanaSavedObject.Find("index-pattern", "default", parameters)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	dataRes := resp["saved_objects"].([]interface{})[0].(map[string]interface{})
	assert.NotEmpty(s.T(), dataRes["id"].(string))

	// Search index pattern from space
	parameters = &OptionalFindParameters{
		Search:       "test*",
		SearchFields: []string{"title"},
		Fields:       []string{"id"},
	}
	resp, err = s.API.KibanaSavedObject.Find("index-pattern", "testacc", parameters)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	dataRes = resp["saved_objects"].([]interface{})[0].(map[string]interface{})
	assert.NotEmpty(s.T(), dataRes["id"].(string))

	// Update index pattern
	dataJSON = `{"attributes": {"title": "test-pattern2-*"}}`
	err = json.Unmarshal([]byte(dataJSON), &data)
	if err != nil {
		panic(err)
	}
	resp, err = s.API.KibanaSavedObject.Update(data, "index-pattern", "test", "default")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	assert.Equal(s.T(), "test", resp["id"])
	assert.Equal(s.T(), "test-pattern2-*", resp["attributes"].(map[string]interface{})["title"])

	// Update index pattern in space
	dataJSON = `{"attributes": {"title": "test-pattern2-*"}}`
	err = json.Unmarshal([]byte(dataJSON), &data)
	if err != nil {
		panic(err)
	}
	resp, err = s.API.KibanaSavedObject.Update(data, "index-pattern", "test2", "testacc")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	assert.Equal(s.T(), "test2", resp["id"])
	assert.Equal(s.T(), "test-pattern2-*", resp["attributes"].(map[string]interface{})["title"])

	// Export index pattern
	request := []map[string]string{
		{
			"type": "index-pattern",
			"id":   "test",
		},
	}
	dataExported, err := s.API.KibanaSavedObject.Export(nil, request, true, "default")
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), dataExported)

	// Export index pattern from space
	request = []map[string]string{
		{
			"type": "index-pattern",
			"id":   "test2",
		},
	}
	dataExported, err = s.API.KibanaSavedObject.Export(nil, request, true, "testacc")
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), dataExported)

	// import index pattern
	b, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	resp2, err := s.API.KibanaSavedObject.Import(b, true, "default")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp2)
	assert.Equal(s.T(), true, resp2["success"])

	// import index pattern in space
	b, err = json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	resp, err = s.API.KibanaSavedObject.Import(b, true, "testacc")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	assert.Equal(s.T(), true, resp["success"])

	// Delete index pattern
	err = s.API.KibanaSavedObject.Delete("index-pattern", "test", "default")
	assert.NoError(s.T(), err)
	resp, err = s.API.KibanaSavedObject.Get("index-pattern", "test", "default")
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), resp)

	// Delete index pattern in space
	err = s.API.KibanaSavedObject.Delete("index-pattern", "test2", "testacc")
	assert.NoError(s.T(), err)
	resp, err = s.API.KibanaSavedObject.Get("index-pattern", "test2", "testacc")
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), resp)
}
