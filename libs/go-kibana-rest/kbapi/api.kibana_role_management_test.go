package kbapi

import (
	"github.com/stretchr/testify/assert"
)

func (s *KBAPITestSuite) TestKibanaRoleManagement() {

	// Create new role
	kibanaRole := &KibanaRole{
		Name: "test",
		Elasticsearch: &KibanaRoleElasticsearch{
			Indices: []KibanaRoleElasticsearchIndice{
				{
					Names: []string{
						"*",
					},
					Privileges: []string{
						"read",
					},
				},
			},
		},
		Kibana: []KibanaRoleKibana{
			{
				Base: []string{
					"read",
				},
			},
		},
	}
	kibanaRole, err := s.API.KibanaRoleManagement.CreateOrUpdate(kibanaRole)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), kibanaRole)
	assert.Equal(s.T(), "test", kibanaRole.Name)

	// Get role
	kibanaRole, err = s.API.KibanaRoleManagement.Get("test")
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), kibanaRole)
	assert.Equal(s.T(), "test", kibanaRole.Name)

	// List role
	kibanaRoles, err := s.API.KibanaRoleManagement.List()
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), kibanaRoles)

	// Delete role
	err = s.API.KibanaRoleManagement.Delete("test")
	assert.NoError(s.T(), err)
	kibanaRole, err = s.API.KibanaRoleManagement.Get("test")
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), kibanaRole)

}
