package kibana

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type KBTestSuite struct {
	suite.Suite
}

func (s *KBTestSuite) SetupSuite() {

	// Init logger
	logrus.SetFormatter(new(prefixed.TextFormatter))
	logrus.SetLevel(logrus.DebugLevel)

}

func (s *KBTestSuite) SetupTest() {

	// Do somethink before each test

}

func TestKBTestSuite(t *testing.T) {
	suite.Run(t, new(KBTestSuite))
}

func (s *KBTestSuite) TestNewClient() {

	cfg := Config{
		Address:          "http://127.0.0.1:5601",
		Username:         "elastic",
		Password:         "changeme",
		DisableVerifySSL: true,
	}

	client, err := NewClient(cfg)

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), client)

}

func (s *KBTestSuite) TestNewDefaultClient() {

	client, err := NewDefaultClient()

	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), client)

}
