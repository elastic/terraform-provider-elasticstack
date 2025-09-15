# BulkPurgeRollupRequestPurgePolicy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PurgeType** | Pointer to **string** | Specifies whether documents will be purged based on a specific age or on a timestamp | [optional] 
**Age** | Pointer to **string** | The duration to determine which documents to purge, formatted as {duration}{unit}. This value should be greater than or equal to the time window of every SLO provided. | [optional] 
**Timestamp** | Pointer to **string** | The timestamp to determine which documents to purge, formatted in ISO. This value should be older than the applicable time window of every SLO provided. | [optional] 

## Methods

### NewBulkPurgeRollupRequestPurgePolicy

`func NewBulkPurgeRollupRequestPurgePolicy() *BulkPurgeRollupRequestPurgePolicy`

NewBulkPurgeRollupRequestPurgePolicy instantiates a new BulkPurgeRollupRequestPurgePolicy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBulkPurgeRollupRequestPurgePolicyWithDefaults

`func NewBulkPurgeRollupRequestPurgePolicyWithDefaults() *BulkPurgeRollupRequestPurgePolicy`

NewBulkPurgeRollupRequestPurgePolicyWithDefaults instantiates a new BulkPurgeRollupRequestPurgePolicy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPurgeType

`func (o *BulkPurgeRollupRequestPurgePolicy) GetPurgeType() string`

GetPurgeType returns the PurgeType field if non-nil, zero value otherwise.

### GetPurgeTypeOk

`func (o *BulkPurgeRollupRequestPurgePolicy) GetPurgeTypeOk() (*string, bool)`

GetPurgeTypeOk returns a tuple with the PurgeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPurgeType

`func (o *BulkPurgeRollupRequestPurgePolicy) SetPurgeType(v string)`

SetPurgeType sets PurgeType field to given value.

### HasPurgeType

`func (o *BulkPurgeRollupRequestPurgePolicy) HasPurgeType() bool`

HasPurgeType returns a boolean if a field has been set.

### GetAge

`func (o *BulkPurgeRollupRequestPurgePolicy) GetAge() string`

GetAge returns the Age field if non-nil, zero value otherwise.

### GetAgeOk

`func (o *BulkPurgeRollupRequestPurgePolicy) GetAgeOk() (*string, bool)`

GetAgeOk returns a tuple with the Age field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAge

`func (o *BulkPurgeRollupRequestPurgePolicy) SetAge(v string)`

SetAge sets Age field to given value.

### HasAge

`func (o *BulkPurgeRollupRequestPurgePolicy) HasAge() bool`

HasAge returns a boolean if a field has been set.

### GetTimestamp

`func (o *BulkPurgeRollupRequestPurgePolicy) GetTimestamp() string`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *BulkPurgeRollupRequestPurgePolicy) GetTimestampOk() (*string, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *BulkPurgeRollupRequestPurgePolicy) SetTimestamp(v string)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *BulkPurgeRollupRequestPurgePolicy) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


