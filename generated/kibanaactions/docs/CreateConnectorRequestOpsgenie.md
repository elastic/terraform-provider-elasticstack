# CreateConnectorRequestOpsgenie

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesOpsgenie**](ConfigPropertiesOpsgenie.md) |  | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | [**SecretsPropertiesOpsgenie**](SecretsPropertiesOpsgenie.md) |  | 

## Methods

### NewCreateConnectorRequestOpsgenie

`func NewCreateConnectorRequestOpsgenie(config ConfigPropertiesOpsgenie, connectorTypeId string, name string, secrets SecretsPropertiesOpsgenie, ) *CreateConnectorRequestOpsgenie`

NewCreateConnectorRequestOpsgenie instantiates a new CreateConnectorRequestOpsgenie object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateConnectorRequestOpsgenieWithDefaults

`func NewCreateConnectorRequestOpsgenieWithDefaults() *CreateConnectorRequestOpsgenie`

NewCreateConnectorRequestOpsgenieWithDefaults instantiates a new CreateConnectorRequestOpsgenie object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *CreateConnectorRequestOpsgenie) GetConfig() ConfigPropertiesOpsgenie`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *CreateConnectorRequestOpsgenie) GetConfigOk() (*ConfigPropertiesOpsgenie, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *CreateConnectorRequestOpsgenie) SetConfig(v ConfigPropertiesOpsgenie)`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *CreateConnectorRequestOpsgenie) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *CreateConnectorRequestOpsgenie) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *CreateConnectorRequestOpsgenie) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetName

`func (o *CreateConnectorRequestOpsgenie) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateConnectorRequestOpsgenie) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateConnectorRequestOpsgenie) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *CreateConnectorRequestOpsgenie) GetSecrets() SecretsPropertiesOpsgenie`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *CreateConnectorRequestOpsgenie) GetSecretsOk() (*SecretsPropertiesOpsgenie, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *CreateConnectorRequestOpsgenie) SetSecrets(v SecretsPropertiesOpsgenie)`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


