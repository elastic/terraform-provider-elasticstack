# SecretsPropertiesServicenow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClientSecret** | Pointer to **string** | The client secret assigned to your OAuth application. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;. | [optional] 
**Password** | Pointer to **string** | The password for HTTP basic authentication. This property is required when &#x60;isOAuth&#x60; is &#x60;false&#x60;. | [optional] 
**PrivateKey** | Pointer to **string** | The RSA private key that you created for use in ServiceNow. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60;. | [optional] 
**PrivateKeyPassword** | Pointer to **string** | The password for the RSA private key. This property is required when &#x60;isOAuth&#x60; is &#x60;true&#x60; and you set a password on your private key. | [optional] 
**Username** | Pointer to **string** | The username for HTTP basic authentication. This property is required when &#x60;isOAuth&#x60; is &#x60;false&#x60;. | [optional] 

## Methods

### NewSecretsPropertiesServicenow

`func NewSecretsPropertiesServicenow() *SecretsPropertiesServicenow`

NewSecretsPropertiesServicenow instantiates a new SecretsPropertiesServicenow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretsPropertiesServicenowWithDefaults

`func NewSecretsPropertiesServicenowWithDefaults() *SecretsPropertiesServicenow`

NewSecretsPropertiesServicenowWithDefaults instantiates a new SecretsPropertiesServicenow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClientSecret

`func (o *SecretsPropertiesServicenow) GetClientSecret() string`

GetClientSecret returns the ClientSecret field if non-nil, zero value otherwise.

### GetClientSecretOk

`func (o *SecretsPropertiesServicenow) GetClientSecretOk() (*string, bool)`

GetClientSecretOk returns a tuple with the ClientSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientSecret

`func (o *SecretsPropertiesServicenow) SetClientSecret(v string)`

SetClientSecret sets ClientSecret field to given value.

### HasClientSecret

`func (o *SecretsPropertiesServicenow) HasClientSecret() bool`

HasClientSecret returns a boolean if a field has been set.

### GetPassword

`func (o *SecretsPropertiesServicenow) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *SecretsPropertiesServicenow) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *SecretsPropertiesServicenow) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *SecretsPropertiesServicenow) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetPrivateKey

`func (o *SecretsPropertiesServicenow) GetPrivateKey() string`

GetPrivateKey returns the PrivateKey field if non-nil, zero value otherwise.

### GetPrivateKeyOk

`func (o *SecretsPropertiesServicenow) GetPrivateKeyOk() (*string, bool)`

GetPrivateKeyOk returns a tuple with the PrivateKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivateKey

`func (o *SecretsPropertiesServicenow) SetPrivateKey(v string)`

SetPrivateKey sets PrivateKey field to given value.

### HasPrivateKey

`func (o *SecretsPropertiesServicenow) HasPrivateKey() bool`

HasPrivateKey returns a boolean if a field has been set.

### GetPrivateKeyPassword

`func (o *SecretsPropertiesServicenow) GetPrivateKeyPassword() string`

GetPrivateKeyPassword returns the PrivateKeyPassword field if non-nil, zero value otherwise.

### GetPrivateKeyPasswordOk

`func (o *SecretsPropertiesServicenow) GetPrivateKeyPasswordOk() (*string, bool)`

GetPrivateKeyPasswordOk returns a tuple with the PrivateKeyPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivateKeyPassword

`func (o *SecretsPropertiesServicenow) SetPrivateKeyPassword(v string)`

SetPrivateKeyPassword sets PrivateKeyPassword field to given value.

### HasPrivateKeyPassword

`func (o *SecretsPropertiesServicenow) HasPrivateKeyPassword() bool`

HasPrivateKeyPassword returns a boolean if a field has been set.

### GetUsername

`func (o *SecretsPropertiesServicenow) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *SecretsPropertiesServicenow) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *SecretsPropertiesServicenow) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *SecretsPropertiesServicenow) HasUsername() bool`

HasUsername returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


