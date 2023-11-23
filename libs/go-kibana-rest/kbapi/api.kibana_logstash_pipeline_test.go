package kbapi

import (
	"github.com/stretchr/testify/assert"
)

func (s *KBAPITestSuite) TestKibanaLogstashPipeline() {

	// Create new logstash pipeline
	logstashPipeline := &LogstashPipeline{
		ID:          "test",
		Description: "Acceptance test",
		Pipeline:    "input { stdin {} } output { stdout {} }",
		Settings: map[string]interface{}{
			"queue.type": "persisted",
		},
	}
	logstashPipeline, err := s.API.KibanaLogstashPipeline.CreateOrUpdate(logstashPipeline)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), logstashPipeline)
	assert.Equal(s.T(), "test", logstashPipeline.ID)

	// Get logstash pipeline
	logstashPipeline, err = s.API.KibanaLogstashPipeline.Get("test")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), logstashPipeline)
	assert.Equal(s.T(), "test", logstashPipeline.ID)

	// List logstash pipeline
	logstashPipelines, err := s.API.KibanaLogstashPipeline.List()
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), logstashPipelines)

	// Delete logstash pipeline
	err = s.API.KibanaLogstashPipeline.Delete("test")
	assert.NoError(s.T(), err)
	logstashPipeline, err = s.API.KibanaLogstashPipeline.Get("test")
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), logstashPipeline)

}
