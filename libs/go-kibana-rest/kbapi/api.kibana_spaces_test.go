package kbapi

import (
	"github.com/stretchr/testify/assert"
)

func (s *KBAPITestSuite) TestKibanaSpaces() {

	// List kibana space
	kibanaSpaces, err := s.API.KibanaSpaces.List()
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), kibanaSpaces)

	// Get the default Space
	kibanaSpace, err := s.API.KibanaSpaces.Get(kibanaSpaces[0].ID)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), kibanaSpaces[0].ID, kibanaSpace.ID)
	assert.Equal(s.T(), "Default", kibanaSpace.Name)

	// Create new space
	kibanaSpace = &KibanaSpace{
		ID:          "test",
		Name:        "test",
		Description: "My test",
	}
	kibanaSpace, err = s.KibanaSpaces.Create(kibanaSpace)
	assert.NoError(s.T(), err)
	assert.NotEmpty(s.T(), kibanaSpace.ID)

	// Update space
	kibanaSpace.Name = "test2"
	kibanaSpace, err = s.KibanaSpaces.Update(kibanaSpace)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "test2", kibanaSpace.Name)

	// Copy object on space
	parameter := &KibanaSpaceCopySavedObjectParameter{
		Spaces:            []string{"test"},
		IncludeReferences: true,
		Overwrite:         false,
		CreateNewCopies:   true,
		Objects: []KibanaSpaceObjectParameter{
			{
				Type: "config",
				ID:   "8.5.0",
			},
		},
	}
	err = s.KibanaSpaces.CopySavedObjects(parameter, "")
	assert.NoError(s.T(), err)

	// Delete space
	err = s.KibanaSpaces.Delete(kibanaSpace.ID)
	assert.NoError(s.T(), err)
	kibanaSpace, err = s.KibanaSpaces.Get(kibanaSpace.ID)
	assert.NoError(s.T(), err)
	assert.Nil(s.T(), kibanaSpace)

}
