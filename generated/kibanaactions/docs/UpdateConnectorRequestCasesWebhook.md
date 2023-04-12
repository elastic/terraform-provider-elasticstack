# UpdateConnectorRequestCasesWebhook

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Config** | [**ConfigPropertiesCasesWebhook**](ConfigPropertiesCasesWebhook.md) |  | 
**Name** | **string** | The display name for the connector. | 
**Secrets** | Pointer to [**SecretsPropertiesCasesWebhook**](SecretsPropertiesCasesWebhook.md) |  | [optional] 

## Methods

### NewUpdateConnectorRequestCasesWebhook

`func NewUpdateConnectorRequestCasesWebhook(config ConfigPropertiesCasesWebhook, name string, ) *UpdateConnectorRequestCasesWebhook`

NewUpdateConnectorRequestCasesWebhook instantiates a new UpdateConnectorRequestCasesWebhook object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateConnectorRequestCasesWebhookWithDefaults

`func NewUpdateConnectorRequestCasesWebhookWithDefaults() *UpdateConnectorRequestCasesWebhook`

NewUpdateConnectorRequestCasesWebhookWithDefaults instantiates a new UpdateConnectorRequestCasesWebhook object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfig

`func (o *UpdateConnectorRequestCasesWebhook) GetConfig() ConfigPropertiesCasesWebhook`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *UpdateConnectorRequestCasesWebhook) GetConfigOk() (*ConfigPropertiesCasesWebhook, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *UpdateConnectorRequestCasesWebhook) SetConfig(v ConfigPropertiesCasesWebhook)`

SetConfig sets Config field to given value.


### GetName

`func (o *UpdateConnectorRequestCasesWebhook) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdateConnectorRequestCasesWebhook) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdateConnectorRequestCasesWebhook) SetName(v string)`

SetName sets Name field to given value.


### GetSecrets

`func (o *UpdateConnectorRequestCasesWebhook) GetSecrets() SecretsPropertiesCasesWebhook`

GetSecrets returns the Secrets field if non-nil, zero value otherwise.

### GetSecretsOk

`func (o *UpdateConnectorRequestCasesWebhook) GetSecretsOk() (*SecretsPropertiesCasesWebhook, bool)`

GetSecretsOk returns a tuple with the Secrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecrets

`func (o *UpdateConnectorRequestCasesWebhook) SetSecrets(v SecretsPropertiesCasesWebhook)`

SetSecrets sets Secrets field to given value.

### HasSecrets

`func (o *UpdateConnectorRequestCasesWebhook) HasSecrets() bool`

HasSecrets returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


