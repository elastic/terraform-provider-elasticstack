# CreateConnectorRequestXmatters

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | **map[string]interface{}** | Defines properties for connectors when type is &#x60;.xmatters&#x60;. | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | **map[string]interface{}** | Defines secrets for connectors when type is &#x60;.xmatters&#x60;. | 

## Methods

### NewCreateConnectorRequestXmatters

`func NewCreateConnectorRequestXmatters(config map[string]interface{}, connectorTypeId string, name string, secrets map[string]interface{}, ) *CreateConnectorRequestXmatters`

NewCreateConnectorRequestXmatters instantiates a new CreateConnectorRequestXmatters object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateConnectorRequestXmattersWithDefaults

`func NewCreateConnectorRequestXmattersWithDefaults() *CreateConnectorRequestXmatters`

NewCreateConnectorRequestXmattersWithDefaults instantiates a new CreateConnectorRequestXmatters object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *CreateConnectorRequestXmatters) GetConfig() map[string]interface{}`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *CreateConnectorRequestXmatters) GetConfigOk() (*map[string]interface{}, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *CreateConnectorRequestXmatters) SetConfig(v map[string]interface{})`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *CreateConnectorRequestXmatters) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *CreateConnectorRequestXmatters) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *CreateConnectorRequestXmatters) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetName

`func (o *CreateConnectorRequestXmatters) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateConnectorRequestXmatters) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateConnectorRequestXmatters) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *CreateConnectorRequestXmatters) GetSecrets() map[string]interface{}`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *CreateConnectorRequestXmatters) GetSecretsOk() (*map[string]interface{}, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *CreateConnectorRequestXmatters) SetSecrets(v map[string]interface{})`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


