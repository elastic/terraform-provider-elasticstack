package kbapi

import (
	"github.com/stretchr/testify/assert"
)

func (s *KBAPITestSuite) TestKibanaStatus() {

	// List kibana space
	kibanaStatus, err := s.API.KibanaStatus.Get()
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), kibanaStatus)
}
