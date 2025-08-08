# FindSloDefinitionsResponseOneOf1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Page** | Pointer to **float64** | for backward compability | [optional] [default to 1]
**PerPage** | Pointer to **float64** | for backward compability | [optional] 
**Size** | Pointer to **float64** |  | [optional] 
**SearchAfter** | Pointer to **[]string** | the cursor to provide to get the next paged results | [optional] 
**Total** | Pointer to **float64** |  | [optional] 
**Results** | Pointer to [**[]SloWithSummaryResponse**](SloWithSummaryResponse.md) |  | [optional] 

## Methods

### NewFindSloDefinitionsResponseOneOf1

`func NewFindSloDefinitionsResponseOneOf1() *FindSloDefinitionsResponseOneOf1`

NewFindSloDefinitionsResponseOneOf1 instantiates a new FindSloDefinitionsResponseOneOf1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFindSloDefinitionsResponseOneOf1WithDefaults

`func NewFindSloDefinitionsResponseOneOf1WithDefaults() *FindSloDefinitionsResponseOneOf1`

NewFindSloDefinitionsResponseOneOf1WithDefaults instantiates a new FindSloDefinitionsResponseOneOf1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPage

`func (o *FindSloDefinitionsResponseOneOf1) GetPage() float64`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *FindSloDefinitionsResponseOneOf1) GetPageOk() (*float64, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *FindSloDefinitionsResponseOneOf1) SetPage(v float64)`

SetPage sets Page field to given value.

### HasPage

`func (o *FindSloDefinitionsResponseOneOf1) HasPage() bool`

HasPage returns a boolean if a field has been set.

### GetPerPage

`func (o *FindSloDefinitionsResponseOneOf1) GetPerPage() float64`

GetPerPage returns the PerPage field if non-nil, zero value otherwise.

### GetPerPageOk

`func (o *FindSloDefinitionsResponseOneOf1) GetPerPageOk() (*float64, bool)`

GetPerPageOk returns a tuple with the PerPage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPerPage

`func (o *FindSloDefinitionsResponseOneOf1) SetPerPage(v float64)`

SetPerPage sets PerPage field to given value.

### HasPerPage

`func (o *FindSloDefinitionsResponseOneOf1) HasPerPage() bool`

HasPerPage returns a boolean if a field has been set.

### GetSize

`func (o *FindSloDefinitionsResponseOneOf1) GetSize() float64`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *FindSloDefinitionsResponseOneOf1) GetSizeOk() (*float64, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *FindSloDefinitionsResponseOneOf1) SetSize(v float64)`

SetSize sets Size field to given value.

### HasSize

`func (o *FindSloDefinitionsResponseOneOf1) HasSize() bool`

HasSize returns a boolean if a field has been set.

### GetSearchAfter

`func (o *FindSloDefinitionsResponseOneOf1) GetSearchAfter() []string`

GetSearchAfter returns the SearchAfter field if non-nil, zero value otherwise.

### GetSearchAfterOk

`func (o *FindSloDefinitionsResponseOneOf1) GetSearchAfterOk() (*[]string, bool)`

GetSearchAfterOk returns a tuple with the SearchAfter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSearchAfter

`func (o *FindSloDefinitionsResponseOneOf1) SetSearchAfter(v []string)`

SetSearchAfter sets SearchAfter field to given value.

### HasSearchAfter

`func (o *FindSloDefinitionsResponseOneOf1) HasSearchAfter() bool`

HasSearchAfter returns a boolean if a field has been set.

### GetTotal

`func (o *FindSloDefinitionsResponseOneOf1) GetTotal() float64`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *FindSloDefinitionsResponseOneOf1) GetTotalOk() (*float64, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *FindSloDefinitionsResponseOneOf1) SetTotal(v float64)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *FindSloDefinitionsResponseOneOf1) HasTotal() bool`

HasTotal returns a boolean if a field has been set.

### GetResults

`func (o *FindSloDefinitionsResponseOneOf1) GetResults() []SloWithSummaryResponse`

GetResults returns the Results field if non-nil, zero value otherwise.

### GetResultsOk

`func (o *FindSloDefinitionsResponseOneOf1) GetResultsOk() (*[]SloWithSummaryResponse, bool)`

GetResultsOk returns a tuple with the Results field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResults

`func (o *FindSloDefinitionsResponseOneOf1) SetResults(v []SloWithSummaryResponse)`

SetResults sets Results field to given value.

### HasResults

`func (o *FindSloDefinitionsResponseOneOf1) HasResults() bool`

HasResults returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


