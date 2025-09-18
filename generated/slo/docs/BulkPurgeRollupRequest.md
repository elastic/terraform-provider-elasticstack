# BulkPurgeRollupRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**List** | **[]string** | An array of slo ids | 
**PurgePolicy** | [**BulkPurgeRollupRequestPurgePolicy**](BulkPurgeRollupRequestPurgePolicy.md) |  | 

## Methods

### NewBulkPurgeRollupRequest

`func NewBulkPurgeRollupRequest(list []string, purgePolicy BulkPurgeRollupRequestPurgePolicy, ) *BulkPurgeRollupRequest`

NewBulkPurgeRollupRequest instantiates a new BulkPurgeRollupRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBulkPurgeRollupRequestWithDefaults

`func NewBulkPurgeRollupRequestWithDefaults() *BulkPurgeRollupRequest`

NewBulkPurgeRollupRequestWithDefaults instantiates a new BulkPurgeRollupRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetList

`func (o *BulkPurgeRollupRequest) GetList() []string`

GetList returns the List field if non-nil, zero value otherwise.

### GetListOk

`func (o *BulkPurgeRollupRequest) GetListOk() (*[]string, bool)`

GetListOk returns a tuple with the List field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetList

`func (o *BulkPurgeRollupRequest) SetList(v []string)`

SetList sets List field to given value.


### GetPurgePolicy

`func (o *BulkPurgeRollupRequest) GetPurgePolicy() BulkPurgeRollupRequestPurgePolicy`

GetPurgePolicy returns the PurgePolicy field if non-nil, zero value otherwise.

### GetPurgePolicyOk

`func (o *BulkPurgeRollupRequest) GetPurgePolicyOk() (*BulkPurgeRollupRequestPurgePolicy, bool)`

GetPurgePolicyOk returns a tuple with the PurgePolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPurgePolicy

`func (o *BulkPurgeRollupRequest) SetPurgePolicy(v BulkPurgeRollupRequestPurgePolicy)`

SetPurgePolicy sets PurgePolicy field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


