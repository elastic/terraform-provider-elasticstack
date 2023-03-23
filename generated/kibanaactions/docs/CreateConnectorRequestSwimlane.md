# CreateConnectorRequestSwimlane

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesSwimlane**](ConfigPropertiesSwimlane.md) |  | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | [**SecretsPropertiesSwimlane**](SecretsPropertiesSwimlane.md) |  | 

## Methods

### NewCreateConnectorRequestSwimlane

`func NewCreateConnectorRequestSwimlane(config ConfigPropertiesSwimlane, connectorTypeId string, name string, secrets SecretsPropertiesSwimlane, ) *CreateConnectorRequestSwimlane`

NewCreateConnectorRequestSwimlane instantiates a new CreateConnectorRequestSwimlane object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateConnectorRequestSwimlaneWithDefaults

`func NewCreateConnectorRequestSwimlaneWithDefaults() *CreateConnectorRequestSwimlane`

NewCreateConnectorRequestSwimlaneWithDefaults instantiates a new CreateConnectorRequestSwimlane object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *CreateConnectorRequestSwimlane) GetConfig() ConfigPropertiesSwimlane`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *CreateConnectorRequestSwimlane) GetConfigOk() (*ConfigPropertiesSwimlane, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *CreateConnectorRequestSwimlane) SetConfig(v ConfigPropertiesSwimlane)`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *CreateConnectorRequestSwimlane) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *CreateConnectorRequestSwimlane) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *CreateConnectorRequestSwimlane) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetName

`func (o *CreateConnectorRequestSwimlane) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateConnectorRequestSwimlane) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateConnectorRequestSwimlane) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *CreateConnectorRequestSwimlane) GetSecrets() SecretsPropertiesSwimlane`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *CreateConnectorRequestSwimlane) GetSecretsOk() (*SecretsPropertiesSwimlane, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *CreateConnectorRequestSwimlane) SetSecrets(v SecretsPropertiesSwimlane)`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


