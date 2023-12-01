package kbapi

import (
	"github.com/stretchr/testify/assert"
)

func (s *KBAPITestSuite) TestKibanaShortenURL() {

	// Create new shorten URL
	shortenURL := &ShortenURL{
		LocatorId: "LEGACY_SHORT_URL_LOCATOR",
		Params: map[string]any{
			"url": "/app/kibana#/dashboard?_g=()&_a=(description:'',filters:!(),fullScreenMode:!f,options:(hidePanelTitles:!f,useMargins:!t),panels:!((embeddableConfig:(),gridData:(h:15,i:'1',w:24,x:0,y:0),id:'8f4d0c00-4c86-11e8-b3d7-01146121b73d',panelIndex:'1',type:visualization,version:'7.0.0-alpha1')),query:(language:lucene,query:''),timeRestore:!f,title:'New%20Dashboard',viewMode:edit)",
		},
	}
	shortenURLResponse, err := s.API.KibanaShortenURL.Create(shortenURL)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), shortenURLResponse)
	assert.NotEmpty(s.T(), shortenURLResponse.ID)
}
