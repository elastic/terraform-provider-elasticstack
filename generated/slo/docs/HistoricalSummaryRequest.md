# HistoricalSummaryRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SloIds** | **[]string** | The list of SLO identifiers to get the historical summary for | 

## Methods

### NewHistoricalSummaryRequest

`func NewHistoricalSummaryRequest(sloIds []string, ) *HistoricalSummaryRequest`

NewHistoricalSummaryRequest instantiates a new HistoricalSummaryRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHistoricalSummaryRequestWithDefaults

`func NewHistoricalSummaryRequestWithDefaults() *HistoricalSummaryRequest`

NewHistoricalSummaryRequestWithDefaults instantiates a new HistoricalSummaryRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSloIds

`func (o *HistoricalSummaryRequest) GetSloIds() []string`

GetSloIds returns the SloIds field if non-nil, zero value otherwise.

### GetSloIdsOk

`func (o *HistoricalSummaryRequest) GetSloIdsOk() (*[]string, bool)`

GetSloIdsOk returns a tuple with the SloIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSloIds

`func (o *HistoricalSummaryRequest) SetSloIds(v []string)`

SetSloIds sets SloIds field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


