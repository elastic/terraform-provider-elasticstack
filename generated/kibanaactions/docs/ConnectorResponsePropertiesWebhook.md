# ConnectorResponsePropertiesWebhook

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | **map[string]interface{}** | Defines properties for connectors when type is &#x60;.webhook&#x60;. | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Id** | **string** | The identifier for the connector. | 
**IsDeprecated** | **bool** | Indicates whether the connector type is deprecated. | 
**IsMissingSecrets** | Pointer to **bool** | Indicates whether secrets are missing for the connector. Secrets configuration properties vary depending on the connector type. | [optional] 
**IsPreconfigured** | **bool** | Indicates whether it is a preconfigured connector. If true, the &#x60;config&#x60; and &#x60;is_missing_secrets&#x60; properties are omitted from the response. | 
**Name** | **string** | The display name for the connector. | 

## Methods

### NewConnectorResponsePropertiesWebhook

`func NewConnectorResponsePropertiesWebhook(config map[string]interface{}, connectorTypeId string, id string, isDeprecated bool, isPreconfigured bool, name string, ) *ConnectorResponsePropertiesWebhook`

NewConnectorResponsePropertiesWebhook instantiates a new ConnectorResponsePropertiesWebhook object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConnectorResponsePropertiesWebhookWithDefaults

`func NewConnectorResponsePropertiesWebhookWithDefaults() *ConnectorResponsePropertiesWebhook`

NewConnectorResponsePropertiesWebhookWithDefaults instantiates a new ConnectorResponsePropertiesWebhook object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *ConnectorResponsePropertiesWebhook) GetConfig() map[string]interface{}`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *ConnectorResponsePropertiesWebhook) GetConfigOk() (*map[string]interface{}, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *ConnectorResponsePropertiesWebhook) SetConfig(v map[string]interface{})`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *ConnectorResponsePropertiesWebhook) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *ConnectorResponsePropertiesWebhook) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *ConnectorResponsePropertiesWebhook) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetId

`func (o *ConnectorResponsePropertiesWebhook) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ConnectorResponsePropertiesWebhook) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ConnectorResponsePropertiesWebhook) SetId(v string)`

SetId sets Id field to given value.


### GetIsDeprecated

`func (o *ConnectorResponsePropertiesWebhook) GetIsDeprecated() bool`

GetIsDeprecated returns the IsDeprecated field if non-nil, zero value otherwise.

### GetIsDeprecatedOk

`func (o *ConnectorResponsePropertiesWebhook) GetIsDeprecatedOk() (*bool, bool)`

GetIsDeprecatedOk returns a tuple with the IsDeprecated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDeprecated

`func (o *ConnectorResponsePropertiesWebhook) SetIsDeprecated(v bool)`

SetIsDeprecated sets IsDeprecated field to given value.


### GetIsMissingSecrets

`func (o *ConnectorResponsePropertiesWebhook) GetIsMissingSecrets() bool`

GetIsMissingSecrets returns the IsMissingSecrets field if non-nil, zero value otherwise.

### GetIsMissingSecretsOk

`func (o *ConnectorResponsePropertiesWebhook) GetIsMissingSecretsOk() (*bool, bool)`

GetIsMissingSecretsOk returns a tuple with the IsMissingSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsMissingSecrets

`func (o *ConnectorResponsePropertiesWebhook) SetIsMissingSecrets(v bool)`

SetIsMissingSecrets sets IsMissingSecrets field to given value.

### HasIsMissingSecrets

`func (o *ConnectorResponsePropertiesWebhook) HasIsMissingSecrets() bool`

HasIsMissingSecrets returns a boolean if a field has been set.

### GetIsPreconfigured

`func (o *ConnectorResponsePropertiesWebhook) GetIsPreconfigured() bool`

GetIsPreconfigured returns the IsPreconfigured field if non-nil, zero value otherwise.

### GetIsPreconfiguredOk

`func (o *ConnectorResponsePropertiesWebhook) GetIsPreconfiguredOk() (*bool, bool)`

GetIsPreconfiguredOk returns a tuple with the IsPreconfigured field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsPreconfigured

`func (o *ConnectorResponsePropertiesWebhook) SetIsPreconfigured(v bool)`

SetIsPreconfigured sets IsPreconfigured field to given value.


### GetName

`func (o *ConnectorResponsePropertiesWebhook) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ConnectorResponsePropertiesWebhook) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ConnectorResponsePropertiesWebhook) SetName(v string)`

SetName sets Name field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


