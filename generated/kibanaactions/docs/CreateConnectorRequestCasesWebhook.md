# CreateConnectorRequestCasesWebhook

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesCasesWebhook**](ConfigPropertiesCasesWebhook.md) |  | 
**ConnectorTypeId** | **string** | The type of connector. | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | Pointer to [**SecretsPropertiesCasesWebhook**](SecretsPropertiesCasesWebhook.md) |  | [optional] 

## Methods

### NewCreateConnectorRequestCasesWebhook

`func NewCreateConnectorRequestCasesWebhook(config ConfigPropertiesCasesWebhook, connectorTypeId string, name string, ) *CreateConnectorRequestCasesWebhook`

NewCreateConnectorRequestCasesWebhook instantiates a new CreateConnectorRequestCasesWebhook object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCreateConnectorRequestCasesWebhookWithDefaults

`func NewCreateConnectorRequestCasesWebhookWithDefaults() *CreateConnectorRequestCasesWebhook`

NewCreateConnectorRequestCasesWebhookWithDefaults instantiates a new CreateConnectorRequestCasesWebhook object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *CreateConnectorRequestCasesWebhook) GetConfig() ConfigPropertiesCasesWebhook`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *CreateConnectorRequestCasesWebhook) GetConfigOk() (*ConfigPropertiesCasesWebhook, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *CreateConnectorRequestCasesWebhook) SetConfig(v ConfigPropertiesCasesWebhook)`

SetConfig sets Config field to given value.


### GetConnectorTypeId

`func (o *CreateConnectorRequestCasesWebhook) GetConnectorTypeId() string`

GetConnectorTypeId returns the ConnectorTypeId field if non-nil, zero value otherwise.

### GetConnectorTypeIdOk

`func (o *CreateConnectorRequestCasesWebhook) GetConnectorTypeIdOk() (*string, bool)`

GetConnectorTypeIdOk returns a tuple with the ConnectorTypeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectorTypeId

`func (o *CreateConnectorRequestCasesWebhook) SetConnectorTypeId(v string)`

SetConnectorTypeId sets ConnectorTypeId field to given value.


### GetName

`func (o *CreateConnectorRequestCasesWebhook) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CreateConnectorRequestCasesWebhook) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CreateConnectorRequestCasesWebhook) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *CreateConnectorRequestCasesWebhook) GetSecrets() SecretsPropertiesCasesWebhook`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *CreateConnectorRequestCasesWebhook) GetSecretsOk() (*SecretsPropertiesCasesWebhook, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *CreateConnectorRequestCasesWebhook) SetSecrets(v SecretsPropertiesCasesWebhook)`

SetSecrets sets Secrets field to given value.

### HasSecrets

`func (o *CreateConnectorRequestCasesWebhook) HasSecrets() bool`

HasSecrets returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


