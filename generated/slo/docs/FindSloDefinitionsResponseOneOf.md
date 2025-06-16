# FindSloDefinitionsResponseOneOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Page** | Pointer to **float32** |  | [optional] 
**PerPage** | Pointer to **float32** |  | [optional] 
**Total** | Pointer to **float32** |  | [optional] 
**Results** | Pointer to [**[]SloWithSummaryResponse**](SloWithSummaryResponse.md) |  | [optional] 

## Methods

### NewFindSloDefinitionsResponseOneOf

`func NewFindSloDefinitionsResponseOneOf() *FindSloDefinitionsResponseOneOf`

NewFindSloDefinitionsResponseOneOf instantiates a new FindSloDefinitionsResponseOneOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFindSloDefinitionsResponseOneOfWithDefaults

`func NewFindSloDefinitionsResponseOneOfWithDefaults() *FindSloDefinitionsResponseOneOf`

NewFindSloDefinitionsResponseOneOfWithDefaults instantiates a new FindSloDefinitionsResponseOneOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPage

`func (o *FindSloDefinitionsResponseOneOf) GetPage() float32`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *FindSloDefinitionsResponseOneOf) GetPageOk() (*float32, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *FindSloDefinitionsResponseOneOf) SetPage(v float32)`

SetPage sets Page field to given value.

### HasPage

`func (o *FindSloDefinitionsResponseOneOf) HasPage() bool`

HasPage returns a boolean if a field has been set.

### GetPerPage

`func (o *FindSloDefinitionsResponseOneOf) GetPerPage() float32`

GetPerPage returns the PerPage field if non-nil, zero value otherwise.

### GetPerPageOk

`func (o *FindSloDefinitionsResponseOneOf) GetPerPageOk() (*float32, bool)`

GetPerPageOk returns a tuple with the PerPage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPerPage

`func (o *FindSloDefinitionsResponseOneOf) SetPerPage(v float32)`

SetPerPage sets PerPage field to given value.

### HasPerPage

`func (o *FindSloDefinitionsResponseOneOf) HasPerPage() bool`

HasPerPage returns a boolean if a field has been set.

### GetTotal

`func (o *FindSloDefinitionsResponseOneOf) GetTotal() float32`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *FindSloDefinitionsResponseOneOf) GetTotalOk() (*float32, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *FindSloDefinitionsResponseOneOf) SetTotal(v float32)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *FindSloDefinitionsResponseOneOf) HasTotal() bool`

HasTotal returns a boolean if a field has been set.

### GetResults

`func (o *FindSloDefinitionsResponseOneOf) GetResults() []SloWithSummaryResponse`

GetResults returns the Results field if non-nil, zero value otherwise.

### GetResultsOk

`func (o *FindSloDefinitionsResponseOneOf) GetResultsOk() (*[]SloWithSummaryResponse, bool)`

GetResultsOk returns a tuple with the Results field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResults

`func (o *FindSloDefinitionsResponseOneOf) SetResults(v []SloWithSummaryResponse)`

SetResults sets Results field to given value.

### HasResults

`func (o *FindSloDefinitionsResponseOneOf) HasResults() bool`

HasResults returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


