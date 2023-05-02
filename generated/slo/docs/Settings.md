# Settings

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SyncDelay** | Pointer to **string** | The synch delay to apply to the transform. Default 1m | [optional] 
**Frequency** | Pointer to **string** | Configure how often the transform runs, default 1m | [optional] 

## Methods

### NewSettings

`func NewSettings() *Settings`

NewSettings instantiates a new Settings object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSettingsWithDefaults

`func NewSettingsWithDefaults() *Settings`

NewSettingsWithDefaults instantiates a new Settings object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSyncDelay

`func (o *Settings) GetSyncDelay() string`

GetSyncDelay returns the SyncDelay field if non-nil, zero value otherwise.

### GetSyncDelayOk

`func (o *Settings) GetSyncDelayOk() (*string, bool)`

GetSyncDelayOk returns a tuple with the SyncDelay field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSyncDelay

`func (o *Settings) SetSyncDelay(v string)`

SetSyncDelay sets SyncDelay field to given value.

### HasSyncDelay

`func (o *Settings) HasSyncDelay() bool`

HasSyncDelay returns a boolean if a field has been set.

### GetFrequency

`func (o *Settings) GetFrequency() string`

GetFrequency returns the Frequency field if non-nil, zero value otherwise.

### GetFrequencyOk

`func (o *Settings) GetFrequencyOk() (*string, bool)`

GetFrequencyOk returns a tuple with the Frequency field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrequency

`func (o *Settings) SetFrequency(v string)`

SetFrequency sets Frequency field to given value.

### HasFrequency

`func (o *Settings) HasFrequency() bool`

HasFrequency returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


