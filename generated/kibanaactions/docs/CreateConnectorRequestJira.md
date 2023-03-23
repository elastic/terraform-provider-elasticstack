# CreateConnectorRequestJira

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesJira**](ConfigPropertiesJira.md) |  | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | [**SecretsPropertiesJira**](SecretsPropertiesJira.md) |  | 

## Methods

### NewCreateConnectorRequestJira

`func NewCreateConnectorRequestJira(config ConfigPropertiesJira, connectorTypeId string, name string, secrets SecretsPropertiesJira, ) *CreateConnectorRequestJira`

NewCreateConnectorRequestJira instantiates a new CreateConnectorRequestJira object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateConnectorRequestJiraWithDefaults

`func NewCreateConnectorRequestJiraWithDefaults() *CreateConnectorRequestJira`

NewCreateConnectorRequestJiraWithDefaults instantiates a new CreateConnectorRequestJira object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *CreateConnectorRequestJira) GetConfig() ConfigPropertiesJira`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *CreateConnectorRequestJira) GetConfigOk() (*ConfigPropertiesJira, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *CreateConnectorRequestJira) SetConfig(v ConfigPropertiesJira)`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *CreateConnectorRequestJira) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *CreateConnectorRequestJira) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *CreateConnectorRequestJira) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetName

`func (o *CreateConnectorRequestJira) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateConnectorRequestJira) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateConnectorRequestJira) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *CreateConnectorRequestJira) GetSecrets() SecretsPropertiesJira`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *CreateConnectorRequestJira) GetSecretsOk() (*SecretsPropertiesJira, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *CreateConnectorRequestJira) SetSecrets(v SecretsPropertiesJira)`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


