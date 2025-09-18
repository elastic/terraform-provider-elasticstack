# BulkPurgeRollupRequestPurgePolicyOneOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PurgeType** | Pointer to **string** | Specifies whether documents will be purged based on a specific age or on a timestamp | [optional] 
**Age** | Pointer to **string** | The duration to determine which documents to purge, formatted as {duration}{unit}. This value should be greater than or equal to the time window of every SLO provided. | [optional] 

## Methods

### NewBulkPurgeRollupRequestPurgePolicyOneOf

`func NewBulkPurgeRollupRequestPurgePolicyOneOf() *BulkPurgeRollupRequestPurgePolicyOneOf`

NewBulkPurgeRollupRequestPurgePolicyOneOf instantiates a new BulkPurgeRollupRequestPurgePolicyOneOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBulkPurgeRollupRequestPurgePolicyOneOfWithDefaults

`func NewBulkPurgeRollupRequestPurgePolicyOneOfWithDefaults() *BulkPurgeRollupRequestPurgePolicyOneOf`

NewBulkPurgeRollupRequestPurgePolicyOneOfWithDefaults instantiates a new BulkPurgeRollupRequestPurgePolicyOneOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPurgeType

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf) GetPurgeType() string`

GetPurgeType returns the PurgeType field if non-nil, zero value otherwise.

### GetPurgeTypeOk

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf) GetPurgeTypeOk() (*string, bool)`

GetPurgeTypeOk returns a tuple with the PurgeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPurgeType

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf) SetPurgeType(v string)`

SetPurgeType sets PurgeType field to given value.

### HasPurgeType

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf) HasPurgeType() bool`

HasPurgeType returns a boolean if a field has been set.

### GetAge

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf) GetAge() string`

GetAge returns the Age field if non-nil, zero value otherwise.

### GetAgeOk

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf) GetAgeOk() (*string, bool)`

GetAgeOk returns a tuple with the Age field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAge

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf) SetAge(v string)`

SetAge sets Age field to given value.

### HasAge

`func (o *BulkPurgeRollupRequestPurgePolicyOneOf) HasAge() bool`

HasAge returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


