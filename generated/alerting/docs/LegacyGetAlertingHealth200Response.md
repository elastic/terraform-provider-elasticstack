# LegacyGetAlertingHealth200Response

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AlertingFrameworkHealth** | Pointer to [**LegacyGetAlertingHealth200ResponseAlertingFrameworkHealth**](LegacyGetAlertingHealth200ResponseAlertingFrameworkHealth.md) |  | [optional] 
**HasPermanentEncryptionKey** | Pointer to **bool** | If &#x60;false&#x60;, the encrypted saved object plugin does not have a permanent encryption key. | [optional] 
**IsSufficientlySecure** | Pointer to **bool** | If &#x60;false&#x60;, security is enabled but TLS is not. | [optional] 

## Methods

### NewLegacyGetAlertingHealth200Response

`func NewLegacyGetAlertingHealth200Response() *LegacyGetAlertingHealth200Response`

NewLegacyGetAlertingHealth200Response instantiates a new LegacyGetAlertingHealth200Response object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewLegacyGetAlertingHealth200ResponseWithDefaults

`func NewLegacyGetAlertingHealth200ResponseWithDefaults() *LegacyGetAlertingHealth200Response`

NewLegacyGetAlertingHealth200ResponseWithDefaults instantiates a new LegacyGetAlertingHealth200Response object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAlertingFrameworkHealth

`func (o *LegacyGetAlertingHealth200Response) GetAlertingFrameworkHealth() LegacyGetAlertingHealth200ResponseAlertingFrameworkHealth`

GetAlertingFrameworkHealth returns the AlertingFrameworkHealth field if non-nil, zero value otherwise.

### GetAlertingFrameworkHealthOk

`func (o *LegacyGetAlertingHealth200Response) GetAlertingFrameworkHealthOk() (*LegacyGetAlertingHealth200ResponseAlertingFrameworkHealth, bool)`

GetAlertingFrameworkHealthOk returns a tuple with the AlertingFrameworkHealth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlertingFrameworkHealth

`func (o *LegacyGetAlertingHealth200Response) SetAlertingFrameworkHealth(v LegacyGetAlertingHealth200ResponseAlertingFrameworkHealth)`

SetAlertingFrameworkHealth sets AlertingFrameworkHealth field to given value.

### HasAlertingFrameworkHealth

`func (o *LegacyGetAlertingHealth200Response) HasAlertingFrameworkHealth() bool`

HasAlertingFrameworkHealth returns a boolean if a field has been set.

### GetHasPermanentEncryptionKey

`func (o *LegacyGetAlertingHealth200Response) GetHasPermanentEncryptionKey() bool`

GetHasPermanentEncryptionKey returns the HasPermanentEncryptionKey field if non-nil, zero value otherwise.

### GetHasPermanentEncryptionKeyOk

`func (o *LegacyGetAlertingHealth200Response) GetHasPermanentEncryptionKeyOk() (*bool, bool)`

GetHasPermanentEncryptionKeyOk returns a tuple with the HasPermanentEncryptionKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHasPermanentEncryptionKey

`func (o *LegacyGetAlertingHealth200Response) SetHasPermanentEncryptionKey(v bool)`

SetHasPermanentEncryptionKey sets HasPermanentEncryptionKey field to given value.

### HasHasPermanentEncryptionKey

`func (o *LegacyGetAlertingHealth200Response) HasHasPermanentEncryptionKey() bool`

HasHasPermanentEncryptionKey returns a boolean if a field has been set.

### GetIsSufficientlySecure

`func (o *LegacyGetAlertingHealth200Response) GetIsSufficientlySecure() bool`

GetIsSufficientlySecure returns the IsSufficientlySecure field if non-nil, zero value otherwise.

### GetIsSufficientlySecureOk

`func (o *LegacyGetAlertingHealth200Response) GetIsSufficientlySecureOk() (*bool, bool)`

GetIsSufficientlySecureOk returns a tuple with the IsSufficientlySecure field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsSufficientlySecure

`func (o *LegacyGetAlertingHealth200Response) SetIsSufficientlySecure(v bool)`

SetIsSufficientlySecure sets IsSufficientlySecure field to given value.

### HasIsSufficientlySecure

`func (o *LegacyGetAlertingHealth200Response) HasIsSufficientlySecure() bool`

HasIsSufficientlySecure returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


