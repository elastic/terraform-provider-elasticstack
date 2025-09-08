# KqlWithFiltersOneOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**KqlQuery** | Pointer to **string** |  | [optional] 
**Filters** | Pointer to [**[]Filter**](Filter.md) |  | [optional] 

## Methods

### NewKqlWithFiltersOneOf

`func NewKqlWithFiltersOneOf() *KqlWithFiltersOneOf`

NewKqlWithFiltersOneOf instantiates a new KqlWithFiltersOneOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewKqlWithFiltersOneOfWithDefaults

`func NewKqlWithFiltersOneOfWithDefaults() *KqlWithFiltersOneOf`

NewKqlWithFiltersOneOfWithDefaults instantiates a new KqlWithFiltersOneOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKqlQuery

`func (o *KqlWithFiltersOneOf) GetKqlQuery() string`

GetKqlQuery returns the KqlQuery field if non-nil, zero value otherwise.

### GetKqlQueryOk

`func (o *KqlWithFiltersOneOf) GetKqlQueryOk() (*string, bool)`

GetKqlQueryOk returns a tuple with the KqlQuery field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKqlQuery

`func (o *KqlWithFiltersOneOf) SetKqlQuery(v string)`

SetKqlQuery sets KqlQuery field to given value.

### HasKqlQuery

`func (o *KqlWithFiltersOneOf) HasKqlQuery() bool`

HasKqlQuery returns a boolean if a field has been set.

### GetFilters

`func (o *KqlWithFiltersOneOf) GetFilters() []Filter`

GetFilters returns the Filters field if non-nil, zero value otherwise.

### GetFiltersOk

`func (o *KqlWithFiltersOneOf) GetFiltersOk() (*[]Filter, bool)`

GetFiltersOk returns a tuple with the Filters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilters

`func (o *KqlWithFiltersOneOf) SetFilters(v []Filter)`

SetFilters sets Filters field to given value.

### HasFilters

`func (o *KqlWithFiltersOneOf) HasFilters() bool`

HasFilters returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


