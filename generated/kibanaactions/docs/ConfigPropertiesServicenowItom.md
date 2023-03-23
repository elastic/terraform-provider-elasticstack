# ConfigPropertiesServicenowItom

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ApiUrl** | **string** | The ServiceNow instance URL. | 
**ClientId** | Pointer to **string** | The client ID assigned to your OAuth application. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;.  | [optional] 
**IsOAuth** | Pointer to **bool** | The type of authentication to use. The default value is false, which means basic authentication is used instead of open authorization (OAuth).  | [optional] [default to false]
**JwtKeyId** | Pointer to **string** | The key identifier assigned to the JWT verifier map of your OAuth application. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;.  | [optional] 
**UserIdentifierValue** | Pointer to **string** | The identifier to use for OAuth authentication. This identifier should be the user field you selected when you created an OAuth JWT API endpoint for external clients in your ServiceNow instance. For example, if the selected user field is &#x60;Email&#x60;, the user identifier should be the user&#39;s email address. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;.  | [optional] 

## Methods

### NewConfigPropertiesServicenowItom

`func NewConfigPropertiesServicenowItom(apiUrl string, ) *ConfigPropertiesServicenowItom`

NewConfigPropertiesServicenowItom instantiates a new ConfigPropertiesServicenowItom object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigPropertiesServicenowItomWithDefaults

`func NewConfigPropertiesServicenowItomWithDefaults() *ConfigPropertiesServicenowItom`

NewConfigPropertiesServicenowItomWithDefaults instantiates a new ConfigPropertiesServicenowItom object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetApiUrl

`func (o *ConfigPropertiesServicenowItom) GetApiUrl() string`

GetApiUrl returns the ApiUrl field if non-nil, zero value otherwise.

### GetApiUrlOk

`func (o *ConfigPropertiesServicenowItom) GetApiUrlOk() (*string, bool)`

GetApiUrlOk returns a tuple with the ApiUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetApiUrl

`func (o *ConfigPropertiesServicenowItom) SetApiUrl(v string)`

SetApiUrl sets ApiUrl field to given value.


### GetClientId

`func (o *ConfigPropertiesServicenowItom) GetClientId() string`

GetClientId returns the ClientId field if non-nil, zero value otherwise.

### GetClientIdOk

`func (o *ConfigPropertiesServicenowItom) GetClientIdOk() (*string, bool)`

GetClientIdOk returns a tuple with the ClientId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientId

`func (o *ConfigPropertiesServicenowItom) SetClientId(v string)`

SetClientId sets ClientId field to given value.

### HasClientId

`func (o *ConfigPropertiesServicenowItom) HasClientId() bool`

HasClientId returns a boolean if a field has been set.

### GetIsOAuth

`func (o *ConfigPropertiesServicenowItom) GetIsOAuth() bool`

GetIsOAuth returns the IsOAuth field if non-nil, zero value otherwise.

### GetIsOAuthOk

`func (o *ConfigPropertiesServicenowItom) GetIsOAuthOk() (*bool, bool)`

GetIsOAuthOk returns a tuple with the IsOAuth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsOAuth

`func (o *ConfigPropertiesServicenowItom) SetIsOAuth(v bool)`

SetIsOAuth sets IsOAuth field to given value.

### HasIsOAuth

`func (o *ConfigPropertiesServicenowItom) HasIsOAuth() bool`

HasIsOAuth returns a boolean if a field has been set.

### GetJwtKeyId

`func (o *ConfigPropertiesServicenowItom) GetJwtKeyId() string`

GetJwtKeyId returns the JwtKeyId field if non-nil, zero value otherwise.

### GetJwtKeyIdOk

`func (o *ConfigPropertiesServicenowItom) GetJwtKeyIdOk() (*string, bool)`

GetJwtKeyIdOk returns a tuple with the JwtKeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJwtKeyId

`func (o *ConfigPropertiesServicenowItom) SetJwtKeyId(v string)`

SetJwtKeyId sets JwtKeyId field to given value.

### HasJwtKeyId

`func (o *ConfigPropertiesServicenowItom) HasJwtKeyId() bool`

HasJwtKeyId returns a boolean if a field has been set.

### GetUserIdentifierValue

`func (o *ConfigPropertiesServicenowItom) GetUserIdentifierValue() string`

GetUserIdentifierValue returns the UserIdentifierValue field if non-nil, zero value otherwise.

### GetUserIdentifierValueOk

`func (o *ConfigPropertiesServicenowItom) GetUserIdentifierValueOk() (*string, bool)`

GetUserIdentifierValueOk returns a tuple with the UserIdentifierValue field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserIdentifierValue

`func (o *ConfigPropertiesServicenowItom) SetUserIdentifierValue(v string)`

SetUserIdentifierValue sets UserIdentifierValue field to given value.

### HasUserIdentifierValue

`func (o *ConfigPropertiesServicenowItom) HasUserIdentifierValue() bool`

HasUserIdentifierValue returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


