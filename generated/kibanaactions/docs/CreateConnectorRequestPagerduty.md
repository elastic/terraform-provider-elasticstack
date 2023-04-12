# CreateConnectorRequestPagerduty

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | **map[string]interface{}** | Defines properties for connectors when type is &#x60;.pagerduty&#x60;. | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | **map[string]interface{}** | Defines secrets for connectors when type is &#x60;.pagerduty&#x60;. | 

## Methods

### NewCreateConnectorRequestPagerduty

`func NewCreateConnectorRequestPagerduty(config map[string]interface{}, connectorTypeId string, name string, secrets map[string]interface{}, ) *CreateConnectorRequestPagerduty`

NewCreateConnectorRequestPagerduty instantiates a new CreateConnectorRequestPagerduty object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateConnectorRequestPagerdutyWithDefaults

`func NewCreateConnectorRequestPagerdutyWithDefaults() *CreateConnectorRequestPagerduty`

NewCreateConnectorRequestPagerdutyWithDefaults instantiates a new CreateConnectorRequestPagerduty object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *CreateConnectorRequestPagerduty) GetConfig() map[string]interface{}`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *CreateConnectorRequestPagerduty) GetConfigOk() (*map[string]interface{}, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *CreateConnectorRequestPagerduty) SetConfig(v map[string]interface{})`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *CreateConnectorRequestPagerduty) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *CreateConnectorRequestPagerduty) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *CreateConnectorRequestPagerduty) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetName

`func (o *CreateConnectorRequestPagerduty) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateConnectorRequestPagerduty) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateConnectorRequestPagerduty) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *CreateConnectorRequestPagerduty) GetSecrets() map[string]interface{}`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *CreateConnectorRequestPagerduty) GetSecretsOk() (*map[string]interface{}, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *CreateConnectorRequestPagerduty) SetSecrets(v map[string]interface{})`

SetSecrets sets Secrets field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


