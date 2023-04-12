package models

type KibanaActionConnector struct {
	ConnectorID      string
	SpaceID          string
	Name             string
	ConnectorTypeID  string
	ConfigJSON       string
	SecretsJSON      string
	IsDeprecated     bool
	IsMissingSecrets bool
	IsPreconfigured  bool
}
