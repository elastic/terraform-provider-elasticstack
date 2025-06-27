# SloWithSummaryResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The identifier of the SLO. | 
**Name** | **string** | The name of the SLO. | 
**Description** | **string** | The description of the SLO. | 
**Indicator** | [**SloWithSummaryResponseIndicator**](SloWithSummaryResponseIndicator.md) |  | 
**TimeWindow** | [**TimeWindow**](TimeWindow.md) |  | 
**BudgetingMethod** | [**BudgetingMethod**](BudgetingMethod.md) |  | 
**Objective** | [**Objective**](Objective.md) |  | 
**Settings** | [**Settings**](Settings.md) |  | 
**Revision** | **float32** | The SLO revision | 
**Summary** | [**Summary**](Summary.md) |  | 
**Enabled** | **bool** | Indicate if the SLO is enabled | 
**GroupBy** | [**GroupBy**](GroupBy.md) |  | 
**InstanceId** | **string** | the value derived from the groupBy field, if present, otherwise &#39;*&#39; | 
**Tags** | **[]string** | List of tags | 
**CreatedAt** | **string** | The creation date | 
**UpdatedAt** | **string** | The last update date | 
**Version** | **float32** | The internal SLO version | 

## Methods

### NewSloWithSummaryResponse

`func NewSloWithSummaryResponse(id string, name string, description string, indicator SloWithSummaryResponseIndicator, timeWindow TimeWindow, budgetingMethod BudgetingMethod, objective Objective, settings Settings, revision float32, summary Summary, enabled bool, groupBy GroupBy, instanceId string, tags []string, createdAt string, updatedAt string, version float32, ) *SloWithSummaryResponse`

NewSloWithSummaryResponse instantiates a new SloWithSummaryResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSloWithSummaryResponseWithDefaults

`func NewSloWithSummaryResponseWithDefaults() *SloWithSummaryResponse`

NewSloWithSummaryResponseWithDefaults instantiates a new SloWithSummaryResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *SloWithSummaryResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SloWithSummaryResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SloWithSummaryResponse) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *SloWithSummaryResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *SloWithSummaryResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *SloWithSummaryResponse) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *SloWithSummaryResponse) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *SloWithSummaryResponse) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *SloWithSummaryResponse) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetIndicator

`func (o *SloWithSummaryResponse) GetIndicator() SloWithSummaryResponseIndicator`

GetIndicator returns the Indicator field if non-nil, zero value otherwise.

### GetIndicatorOk

`func (o *SloWithSummaryResponse) GetIndicatorOk() (*SloWithSummaryResponseIndicator, bool)`

GetIndicatorOk returns a tuple with the Indicator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndicator

`func (o *SloWithSummaryResponse) SetIndicator(v SloWithSummaryResponseIndicator)`

SetIndicator sets Indicator field to given value.


### GetTimeWindow

`func (o *SloWithSummaryResponse) GetTimeWindow() TimeWindow`

GetTimeWindow returns the TimeWindow field if non-nil, zero value otherwise.

### GetTimeWindowOk

`func (o *SloWithSummaryResponse) GetTimeWindowOk() (*TimeWindow, bool)`

GetTimeWindowOk returns a tuple with the TimeWindow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeWindow

`func (o *SloWithSummaryResponse) SetTimeWindow(v TimeWindow)`

SetTimeWindow sets TimeWindow field to given value.


### GetBudgetingMethod

`func (o *SloWithSummaryResponse) GetBudgetingMethod() BudgetingMethod`

GetBudgetingMethod returns the BudgetingMethod field if non-nil, zero value otherwise.

### GetBudgetingMethodOk

`func (o *SloWithSummaryResponse) GetBudgetingMethodOk() (*BudgetingMethod, bool)`

GetBudgetingMethodOk returns a tuple with the BudgetingMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBudgetingMethod

`func (o *SloWithSummaryResponse) SetBudgetingMethod(v BudgetingMethod)`

SetBudgetingMethod sets BudgetingMethod field to given value.


### GetObjective

`func (o *SloWithSummaryResponse) GetObjective() Objective`

GetObjective returns the Objective field if non-nil, zero value otherwise.

### GetObjectiveOk

`func (o *SloWithSummaryResponse) GetObjectiveOk() (*Objective, bool)`

GetObjectiveOk returns a tuple with the Objective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjective

`func (o *SloWithSummaryResponse) SetObjective(v Objective)`

SetObjective sets Objective field to given value.


### GetSettings

`func (o *SloWithSummaryResponse) GetSettings() Settings`

GetSettings returns the Settings field if non-nil, zero value otherwise.

### GetSettingsOk

`func (o *SloWithSummaryResponse) GetSettingsOk() (*Settings, bool)`

GetSettingsOk returns a tuple with the Settings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSettings

`func (o *SloWithSummaryResponse) SetSettings(v Settings)`

SetSettings sets Settings field to given value.


### GetRevision

`func (o *SloWithSummaryResponse) GetRevision() float32`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *SloWithSummaryResponse) GetRevisionOk() (*float32, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *SloWithSummaryResponse) SetRevision(v float32)`

SetRevision sets Revision field to given value.


### GetSummary

`func (o *SloWithSummaryResponse) GetSummary() Summary`

GetSummary returns the Summary field if non-nil, zero value otherwise.

### GetSummaryOk

`func (o *SloWithSummaryResponse) GetSummaryOk() (*Summary, bool)`

GetSummaryOk returns a tuple with the Summary field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSummary

`func (o *SloWithSummaryResponse) SetSummary(v Summary)`

SetSummary sets Summary field to given value.


### GetEnabled

`func (o *SloWithSummaryResponse) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *SloWithSummaryResponse) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *SloWithSummaryResponse) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.


### GetGroupBy

`func (o *SloWithSummaryResponse) GetGroupBy() GroupBy`

GetGroupBy returns the GroupBy field if non-nil, zero value otherwise.

### GetGroupByOk

`func (o *SloWithSummaryResponse) GetGroupByOk() (*GroupBy, bool)`

GetGroupByOk returns a tuple with the GroupBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroupBy

`func (o *SloWithSummaryResponse) SetGroupBy(v GroupBy)`

SetGroupBy sets GroupBy field to given value.


### GetInstanceId

`func (o *SloWithSummaryResponse) GetInstanceId() string`

GetInstanceId returns the InstanceId field if non-nil, zero value otherwise.

### GetInstanceIdOk

`func (o *SloWithSummaryResponse) GetInstanceIdOk() (*string, bool)`

GetInstanceIdOk returns a tuple with the InstanceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceId

`func (o *SloWithSummaryResponse) SetInstanceId(v string)`

SetInstanceId sets InstanceId field to given value.


### GetTags

`func (o *SloWithSummaryResponse) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *SloWithSummaryResponse) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *SloWithSummaryResponse) SetTags(v []string)`

SetTags sets Tags field to given value.


### GetCreatedAt

`func (o *SloWithSummaryResponse) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *SloWithSummaryResponse) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *SloWithSummaryResponse) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *SloWithSummaryResponse) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *SloWithSummaryResponse) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *SloWithSummaryResponse) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.


### GetVersion

`func (o *SloWithSummaryResponse) GetVersion() float32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *SloWithSummaryResponse) GetVersionOk() (*float32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *SloWithSummaryResponse) SetVersion(v float32)`

SetVersion sets Version field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


