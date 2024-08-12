# ParamsPropertyInfraMetricThreshold

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Criteria** | Pointer to [**[]OneOf**](OneOf.md) |  | [optional] 
**GroupBy** | Pointer to **NullableString** |  | [optional] 
**FilterQuery** | Pointer to **string** |  | [optional] 
**SourceId** | Pointer to **string** |  | [optional] 
**AlertOnNoData** | Pointer to **bool** |  | [optional] 
**AlertOnGroupDisappear** | Pointer to **bool** |  | [optional] 

## Methods

### NewParamsPropertyInfraMetricThreshold

`func NewParamsPropertyInfraMetricThreshold() *ParamsPropertyInfraMetricThreshold`

NewParamsPropertyInfraMetricThreshold instantiates a new ParamsPropertyInfraMetricThreshold object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewParamsPropertyInfraMetricThresholdWithDefaults

`func NewParamsPropertyInfraMetricThresholdWithDefaults() *ParamsPropertyInfraMetricThreshold`

NewParamsPropertyInfraMetricThresholdWithDefaults instantiates a new ParamsPropertyInfraMetricThreshold object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCriteria

`func (o *ParamsPropertyInfraMetricThreshold) GetCriteria() []*OneOf`

GetCriteria returns the Criteria field if non-nil, zero value otherwise.

### GetCriteriaOk

`func (o *ParamsPropertyInfraMetricThreshold) GetCriteriaOk() (*[]*OneOf, bool)`

GetCriteriaOk returns a tuple with the Criteria field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCriteria

`func (o *ParamsPropertyInfraMetricThreshold) SetCriteria(v []*OneOf)`

SetCriteria sets Criteria field to given value.

### HasCriteria

`func (o *ParamsPropertyInfraMetricThreshold) HasCriteria() bool`

HasCriteria returns a boolean if a field has been set.

### GetGroupBy

`func (o *ParamsPropertyInfraMetricThreshold) GetGroupBy() string`

GetGroupBy returns the GroupBy field if non-nil, zero value otherwise.

### GetGroupByOk

`func (o *ParamsPropertyInfraMetricThreshold) GetGroupByOk() (*string, bool)`

GetGroupByOk returns a tuple with the GroupBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroupBy

`func (o *ParamsPropertyInfraMetricThreshold) SetGroupBy(v string)`

SetGroupBy sets GroupBy field to given value.

### HasGroupBy

`func (o *ParamsPropertyInfraMetricThreshold) HasGroupBy() bool`

HasGroupBy returns a boolean if a field has been set.

### SetGroupByNil

`func (o *ParamsPropertyInfraMetricThreshold) SetGroupByNil(b bool)`

 SetGroupByNil sets the value for GroupBy to be an explicit nil

### UnsetGroupBy
`func (o *ParamsPropertyInfraMetricThreshold) UnsetGroupBy()`

UnsetGroupBy ensures that no value is present for GroupBy, not even an explicit nil
### GetFilterQuery

`func (o *ParamsPropertyInfraMetricThreshold) GetFilterQuery() string`

GetFilterQuery returns the FilterQuery field if non-nil, zero value otherwise.

### GetFilterQueryOk

`func (o *ParamsPropertyInfraMetricThreshold) GetFilterQueryOk() (*string, bool)`

GetFilterQueryOk returns a tuple with the FilterQuery field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilterQuery

`func (o *ParamsPropertyInfraMetricThreshold) SetFilterQuery(v string)`

SetFilterQuery sets FilterQuery field to given value.

### HasFilterQuery

`func (o *ParamsPropertyInfraMetricThreshold) HasFilterQuery() bool`

HasFilterQuery returns a boolean if a field has been set.

### GetSourceId

`func (o *ParamsPropertyInfraMetricThreshold) GetSourceId() string`

GetSourceId returns the SourceId field if non-nil, zero value otherwise.

### GetSourceIdOk

`func (o *ParamsPropertyInfraMetricThreshold) GetSourceIdOk() (*string, bool)`

GetSourceIdOk returns a tuple with the SourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceId

`func (o *ParamsPropertyInfraMetricThreshold) SetSourceId(v string)`

SetSourceId sets SourceId field to given value.

### HasSourceId

`func (o *ParamsPropertyInfraMetricThreshold) HasSourceId() bool`

HasSourceId returns a boolean if a field has been set.

### GetAlertOnNoData

`func (o *ParamsPropertyInfraMetricThreshold) GetAlertOnNoData() bool`

GetAlertOnNoData returns the AlertOnNoData field if non-nil, zero value otherwise.

### GetAlertOnNoDataOk

`func (o *ParamsPropertyInfraMetricThreshold) GetAlertOnNoDataOk() (*bool, bool)`

GetAlertOnNoDataOk returns a tuple with the AlertOnNoData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlertOnNoData

`func (o *ParamsPropertyInfraMetricThreshold) SetAlertOnNoData(v bool)`

SetAlertOnNoData sets AlertOnNoData field to given value.

### HasAlertOnNoData

`func (o *ParamsPropertyInfraMetricThreshold) HasAlertOnNoData() bool`

HasAlertOnNoData returns a boolean if a field has been set.

### GetAlertOnGroupDisappear

`func (o *ParamsPropertyInfraMetricThreshold) GetAlertOnGroupDisappear() bool`

GetAlertOnGroupDisappear returns the AlertOnGroupDisappear field if non-nil, zero value otherwise.

### GetAlertOnGroupDisappearOk

`func (o *ParamsPropertyInfraMetricThreshold) GetAlertOnGroupDisappearOk() (*bool, bool)`

GetAlertOnGroupDisappearOk returns a tuple with the AlertOnGroupDisappear field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAlertOnGroupDisappear

`func (o *ParamsPropertyInfraMetricThreshold) SetAlertOnGroupDisappear(v bool)`

SetAlertOnGroupDisappear sets AlertOnGroupDisappear field to given value.

### HasAlertOnGroupDisappear

`func (o *ParamsPropertyInfraMetricThreshold) HasAlertOnGroupDisappear() bool`

HasAlertOnGroupDisappear returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


