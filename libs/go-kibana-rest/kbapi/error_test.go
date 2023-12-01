package kbapi

import "github.com/stretchr/testify/assert"

func (s *KBAPITestSuite) TestError() {

	err := NewAPIError(404, "test %s error", "plop")
	assert.Equal(s.T(), 404, err.Code)
	assert.Equal(s.T(), "test plop error", err.Error())
}
