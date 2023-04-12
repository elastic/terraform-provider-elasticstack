# CreateConnectorRequestEmail

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | **map[string]interface{}** | Defines properties for connectors when type is &#x60;.email&#x60;. | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | **map[string]interface{}** | Defines secrets for connectors when type is &#x60;.email&#x60;. | 

## Methods

### NewCreateConnectorRequestEmail

`func NewCreateConnectorRequestEmail(config map[string]interface{}, connectorTypeId string, name string, secrets map[string]interface{}, ) *CreateConnectorRequestEmail`

NewCreateConnectorRequestEmail instantiates a new CreateConnectorRequestEmail object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateConnectorRequestEmailWithDefaults

`func NewCreateConnectorRequestEmailWithDefaults() *CreateConnectorRequestEmail`

NewCreateConnectorRequestEmailWithDefaults instantiates a new CreateConnectorRequestEmail object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *CreateConnectorRequestEmail) GetConfig() map[string]interface{}`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *CreateConnectorRequestEmail) GetConfigOk() (*map[string]interface{}, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *CreateConnectorRequestEmail) SetConfig(v map[string]interface{})`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *CreateConnectorRequestEmail) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *CreateConnectorRequestEmail) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *CreateConnectorRequestEmail) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetName

`func (o *CreateConnectorRequestEmail) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateConnectorRequestEmail) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateConnectorRequestEmail) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *CreateConnectorRequestEmail) GetSecrets() map[string]interface{}`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *CreateConnectorRequestEmail) GetSecretsOk() (*map[string]interface{}, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *CreateConnectorRequestEmail) SetSecrets(v map[string]interface{})`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


