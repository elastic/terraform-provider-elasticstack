# SloResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The identifier of the SLO. | 
**Name** | **string** | The name of the SLO. | 
**Description** | **string** | The description of the SLO. | 
**Indicator** | [**SloResponseIndicator**](SloResponseIndicator.md) |  | 
**TimeWindow** | [**TimeWindow**](TimeWindow.md) |  | 
**BudgetingMethod** | [**BudgetingMethod**](BudgetingMethod.md) |  | 
**Objective** | [**Objective**](Objective.md) |  | 
**Settings** | [**Settings**](Settings.md) |  | 
**Revision** | **float64** | The SLO revision | 
**Summary** | [**Summary**](Summary.md) |  | 
**Enabled** | **bool** | Indicate if the SLO is enabled | 
**GroupBy** | **string** | optional group by field to use to generate an SLO per distinct value | 
**InstanceId** | **string** | the value derived from the groupBy field, if present, otherwise &#39;*&#39; | 
**CreatedAt** | **string** | The creation date | 
**UpdatedAt** | **string** | The last update date | 

## Methods

### NewSloResponse

`func NewSloResponse(id string, name string, description string, indicator SloResponseIndicator, timeWindow TimeWindow, budgetingMethod BudgetingMethod, objective Objective, settings Settings, revision float64, summary Summary, enabled bool, groupBy string, instanceId string, createdAt string, updatedAt string, ) *SloResponse`

NewSloResponse instantiates a new SloResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSloResponseWithDefaults

`func NewSloResponseWithDefaults() *SloResponse`

NewSloResponseWithDefaults instantiates a new SloResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *SloResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *SloResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *SloResponse) SetId(v string)`

SetId sets Id field to given value.


### GetName

`func (o *SloResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *SloResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *SloResponse) SetName(v string)`

SetName sets Name field to given value.


### GetDescription

`func (o *SloResponse) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *SloResponse) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *SloResponse) SetDescription(v string)`

SetDescription sets Description field to given value.


### GetIndicator

`func (o *SloResponse) GetIndicator() SloResponseIndicator`

GetIndicator returns the Indicator field if non-nil, zero value otherwise.

### GetIndicatorOk

`func (o *SloResponse) GetIndicatorOk() (*SloResponseIndicator, bool)`

GetIndicatorOk returns a tuple with the Indicator field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIndicator

`func (o *SloResponse) SetIndicator(v SloResponseIndicator)`

SetIndicator sets Indicator field to given value.


### GetTimeWindow

`func (o *SloResponse) GetTimeWindow() TimeWindow`

GetTimeWindow returns the TimeWindow field if non-nil, zero value otherwise.

### GetTimeWindowOk

`func (o *SloResponse) GetTimeWindowOk() (*TimeWindow, bool)`

GetTimeWindowOk returns a tuple with the TimeWindow field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimeWindow

`func (o *SloResponse) SetTimeWindow(v TimeWindow)`

SetTimeWindow sets TimeWindow field to given value.


### GetBudgetingMethod

`func (o *SloResponse) GetBudgetingMethod() BudgetingMethod`

GetBudgetingMethod returns the BudgetingMethod field if non-nil, zero value otherwise.

### GetBudgetingMethodOk

`func (o *SloResponse) GetBudgetingMethodOk() (*BudgetingMethod, bool)`

GetBudgetingMethodOk returns a tuple with the BudgetingMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBudgetingMethod

`func (o *SloResponse) SetBudgetingMethod(v BudgetingMethod)`

SetBudgetingMethod sets BudgetingMethod field to given value.


### GetObjective

`func (o *SloResponse) GetObjective() Objective`

GetObjective returns the Objective field if non-nil, zero value otherwise.

### GetObjectiveOk

`func (o *SloResponse) GetObjectiveOk() (*Objective, bool)`

GetObjectiveOk returns a tuple with the Objective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjective

`func (o *SloResponse) SetObjective(v Objective)`

SetObjective sets Objective field to given value.


### GetSettings

`func (o *SloResponse) GetSettings() Settings`

GetSettings returns the Settings field if non-nil, zero value otherwise.

### GetSettingsOk

`func (o *SloResponse) GetSettingsOk() (*Settings, bool)`

GetSettingsOk returns a tuple with the Settings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSettings

`func (o *SloResponse) SetSettings(v Settings)`

SetSettings sets Settings field to given value.


### GetRevision

`func (o *SloResponse) GetRevision() float64`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *SloResponse) GetRevisionOk() (*float64, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *SloResponse) SetRevision(v float64)`

SetRevision sets Revision field to given value.


### GetSummary

`func (o *SloResponse) GetSummary() Summary`

GetSummary returns the Summary field if non-nil, zero value otherwise.

### GetSummaryOk

`func (o *SloResponse) GetSummaryOk() (*Summary, bool)`

GetSummaryOk returns a tuple with the Summary field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSummary

`func (o *SloResponse) SetSummary(v Summary)`

SetSummary sets Summary field to given value.


### GetEnabled

`func (o *SloResponse) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *SloResponse) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *SloResponse) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.


### GetGroupBy

`func (o *SloResponse) GetGroupBy() string`

GetGroupBy returns the GroupBy field if non-nil, zero value otherwise.

### GetGroupByOk

`func (o *SloResponse) GetGroupByOk() (*string, bool)`

GetGroupByOk returns a tuple with the GroupBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroupBy

`func (o *SloResponse) SetGroupBy(v string)`

SetGroupBy sets GroupBy field to given value.


### GetInstanceId

`func (o *SloResponse) GetInstanceId() string`

GetInstanceId returns the InstanceId field if non-nil, zero value otherwise.

### GetInstanceIdOk

`func (o *SloResponse) GetInstanceIdOk() (*string, bool)`

GetInstanceIdOk returns a tuple with the InstanceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstanceId

`func (o *SloResponse) SetInstanceId(v string)`

SetInstanceId sets InstanceId field to given value.


### GetCreatedAt

`func (o *SloResponse) GetCreatedAt() string`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *SloResponse) GetCreatedAtOk() (*string, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *SloResponse) SetCreatedAt(v string)`

SetCreatedAt sets CreatedAt field to given value.


### GetUpdatedAt

`func (o *SloResponse) GetUpdatedAt() string`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *SloResponse) GetUpdatedAtOk() (*string, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *SloResponse) SetUpdatedAt(v string)`

SetUpdatedAt sets UpdatedAt field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


