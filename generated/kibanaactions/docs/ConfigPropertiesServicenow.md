# ConfigPropertiesServicenow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiUrl** | **string** | The ServiceNow instance URL. | 
**ClientId** | Pointer to **string** | The client ID assigned to your OAuth application. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;.  | [optional] 
**IsOAuth** | Pointer to **bool** | The type of authentication to use. The default value is false, which means basic authentication is used instead of open authorization (OAuth).  | [optional] [default to false]
**JwtKeyId** | Pointer to **string** | The key identifier assigned to the JWT verifier map of your OAuth application. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;.  | [optional] 
**UserIdentifierValue** | Pointer to **string** | The identifier to use for OAuth authentication. This identifier should be the user field you selected when you created an OAuth JWT API endpoint for external clients in your ServiceNow instance. For example, if the selected user field is &#x60;Email&#x60;, the user identifier should be the user&#39;s email address. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;.  | [optional] 
**UsesTableApi** | Pointer to **bool** | Determines whether the connector uses the Table API or the Import Set API. This property is supported only for ServiceNow ITSM and ServiceNow SecOps connectors.  NOTE: If this property is set to &#x60;false&#x60;, the Elastic application should be installed in ServiceNow.  | [optional] [default to true]

## Methods

### NewConfigPropertiesServicenow

`func NewConfigPropertiesServicenow(apiUrl string, ) *ConfigPropertiesServicenow`

NewConfigPropertiesServicenow instantiates a new ConfigPropertiesServicenow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigPropertiesServicenowWithDefaults

`func NewConfigPropertiesServicenowWithDefaults() *ConfigPropertiesServicenow`

NewConfigPropertiesServicenowWithDefaults instantiates a new ConfigPropertiesServicenow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetApiUrl

`func (o *ConfigPropertiesServicenow) GetApiUrl() string`

GetApiUrl returns the ApiUrl field if non-nil, zero value otherwise.

### GetApiUrlOk

`func (o *ConfigPropertiesServicenow) GetApiUrlOk() (*string, bool)`

GetApiUrlOk returns a tuple with the ApiUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiUrl

`func (o *ConfigPropertiesServicenow) SetApiUrl(v string)`

SetApiUrl sets ApiUrl field to given value.


### GetClientId

`func (o *ConfigPropertiesServicenow) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *ConfigPropertiesServicenow) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *ConfigPropertiesServicenow) SetClientId(v string)`

SetClientId sets ClientId field to given value.

### HasClientId

`func (o *ConfigPropertiesServicenow) HasClientId() bool`

HasClientId returns a boolean if a field has been set.

### GetIsOAuth

`func (o *ConfigPropertiesServicenow) GetIsOAuth() bool`

GetIsOAuth returns the IsOAuth field if non-nil, zero value otherwise.

### GetIsOAuthOk

`func (o *ConfigPropertiesServicenow) GetIsOAuthOk() (*bool, bool)`

GetIsOAuthOk returns a tuple with the IsOAuth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsOAuth

`func (o *ConfigPropertiesServicenow) SetIsOAuth(v bool)`

SetIsOAuth sets IsOAuth field to given value.

### HasIsOAuth

`func (o *ConfigPropertiesServicenow) HasIsOAuth() bool`

HasIsOAuth returns a boolean if a field has been set.

### GetJwtKeyId

`func (o *ConfigPropertiesServicenow) GetJwtKeyId() string`

GetJwtKeyId returns the JwtKeyId field if non-nil, zero value otherwise.

### GetJwtKeyIdOk

`func (o *ConfigPropertiesServicenow) GetJwtKeyIdOk() (*string, bool)`

GetJwtKeyIdOk returns a tuple with the JwtKeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJwtKeyId

`func (o *ConfigPropertiesServicenow) SetJwtKeyId(v string)`

SetJwtKeyId sets JwtKeyId field to given value.

### HasJwtKeyId

`func (o *ConfigPropertiesServicenow) HasJwtKeyId() bool`

HasJwtKeyId returns a boolean if a field has been set.

### GetUserIdentifierValue

`func (o *ConfigPropertiesServicenow) GetUserIdentifierValue() string`

GetUserIdentifierValue returns the UserIdentifierValue field if non-nil, zero value otherwise.

### GetUserIdentifierValueOk

`func (o *ConfigPropertiesServicenow) GetUserIdentifierValueOk() (*string, bool)`

GetUserIdentifierValueOk returns a tuple with the UserIdentifierValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserIdentifierValue

`func (o *ConfigPropertiesServicenow) SetUserIdentifierValue(v string)`

SetUserIdentifierValue sets UserIdentifierValue field to given value.

### HasUserIdentifierValue

`func (o *ConfigPropertiesServicenow) HasUserIdentifierValue() bool`

HasUserIdentifierValue returns a boolean if a field has been set.

### GetUsesTableApi

`func (o *ConfigPropertiesServicenow) GetUsesTableApi() bool`

GetUsesTableApi returns the UsesTableApi field if non-nil, zero value otherwise.

### GetUsesTableApiOk

`func (o *ConfigPropertiesServicenow) GetUsesTableApiOk() (*bool, bool)`

GetUsesTableApiOk returns a tuple with the UsesTableApi field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsesTableApi

`func (o *ConfigPropertiesServicenow) SetUsesTableApi(v bool)`

SetUsesTableApi sets UsesTableApi field to given value.

### HasUsesTableApi

`func (o *ConfigPropertiesServicenow) HasUsesTableApi() bool`

HasUsesTableApi returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


