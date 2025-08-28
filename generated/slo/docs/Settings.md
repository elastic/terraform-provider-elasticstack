# Settings

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SyncField** | Pointer to **string** | The date field that is used to identify new documents in the source. It is strongly recommended to use a field that contains the ingest timestamp. If you use a different field, you might need to set the delay such that it accounts for data transmission delays. When unspecified, we use the indicator timestamp field. | [optional] 
**SyncDelay** | Pointer to **string** | The time delay in minutes between the current time and the latest source data time. Increasing the value will delay any alerting. The default value is 1 minute. The minimum value is 1m and the maximum is 359m. It should always be greater then source index refresh interval. | [optional] [default to "1m"]
**Frequency** | Pointer to **string** | The interval between checks for changes in the source data. The minimum value is 1m and the maximum is 59m. The default value is 1 minute. | [optional] [default to "1m"]
**PreventInitialBackfill** | Pointer to **bool** | Start aggregating data from the time the SLO is created, instead of backfilling data from the beginning of the time window. | [optional] [default to false]

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

### GetSyncField

`func (o *Settings) GetSyncField() string`

GetSyncField returns the SyncField field if non-nil, zero value otherwise.

### GetSyncFieldOk

`func (o *Settings) GetSyncFieldOk() (*string, bool)`

GetSyncFieldOk returns a tuple with the SyncField field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSyncField

`func (o *Settings) SetSyncField(v string)`

SetSyncField sets SyncField field to given value.

### HasSyncField

`func (o *Settings) HasSyncField() bool`

HasSyncField returns a boolean if a field has been set.

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

### GetPreventInitialBackfill

`func (o *Settings) GetPreventInitialBackfill() bool`

GetPreventInitialBackfill returns the PreventInitialBackfill field if non-nil, zero value otherwise.

### GetPreventInitialBackfillOk

`func (o *Settings) GetPreventInitialBackfillOk() (*bool, bool)`

GetPreventInitialBackfillOk returns a tuple with the PreventInitialBackfill field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreventInitialBackfill

`func (o *Settings) SetPreventInitialBackfill(v bool)`

SetPreventInitialBackfill sets PreventInitialBackfill field to given value.

### HasPreventInitialBackfill

`func (o *Settings) HasPreventInitialBackfill() bool`

HasPreventInitialBackfill returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


