# BulkPurgeRollupRequestPurgePolicyOneOf1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PurgeType** | Pointer to **string** | Specifies whether documents will be purged based on a specific age or on a timestamp | [optional] 
**Timestamp** | Pointer to **string** | The timestamp to determine which documents to purge, formatted in ISO. This value should be older than the applicable time window of every SLO provided. | [optional] 

## Methods

### NewBulkPurgeRollupRequestPurgePolicyOneOf1

`func NewBulkPurgeRollupRequestPurgePolicyOneOf1() *BulkPurgeRollupRequestPurgePolicyOneOf1`

NewBulkPurgeRollupRequestPurgePolicyOneOf1 instantiates a new BulkPurgeRollupRequestPurgePolicyOneOf1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBulkPurgeRollupRequestPurgePolicyOneOf1WithDefaults

`func NewBulkPurgeRollupRequestPurgePolicyOneOf1WithDefaults() *BulkPurgeRollupRequestPurgePolicyOneOf1`

NewBulkPurgeRollupRequestPurgePolicyOneOf1WithDefaults instantiates a new BulkPurgeRollupRequestPurgePolicyOneOf1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPurgeType

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf1) GetPurgeType() string`

GetPurgeType returns the PurgeType field if non-nil, zero value otherwise.

### GetPurgeTypeOk

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf1) GetPurgeTypeOk() (*string, bool)`

GetPurgeTypeOk returns a tuple with the PurgeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPurgeType

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf1) SetPurgeType(v string)`

SetPurgeType sets PurgeType field to given value.

### HasPurgeType

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf1) HasPurgeType() bool`

HasPurgeType returns a boolean if a field has been set.

### GetTimestamp

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf1) GetTimestamp() string`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf1) GetTimestampOk() (*string, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf1) SetTimestamp(v string)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf1) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


