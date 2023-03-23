# ConfigPropertiesCasesWebhook

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CreateCommentJson** | Pointer to **string** | A JSON payload sent to the create comment URL to create a case comment. You can use variables to add Kibana Cases data to the payload. The required variable is &#x60;case.comment&#x60;. Due to Mustache template variables (the text enclosed in triple braces, for example, &#x60;{{{case.title}}}&#x60;), the JSON is not validated when you create the connector. The JSON is validated once the Mustache variables have been placed when the REST method runs. Manually ensure that the JSON is valid, disregarding the Mustache variables, so the later validation will pass.  | [optional] 
**CreateCommentMethod** | Pointer to **string** | The REST API HTTP request method to create a case comment in the third-party system. Valid values are &#x60;patch&#x60;, &#x60;post&#x60;, and &#x60;put&#x60;.  | [optional] [default to "put"]
**CreateCommentUrl** | Pointer to **string** | The REST API URL to create a case comment by ID in the third-party system. You can use a variable to add the external system ID to the URL. If you are using the &#x60;xpack.actions.allowedHosts setting&#x60;, add the hostname to the allowed hosts.  | [optional] 
**CreateIncidentJson** | **string** | A JSON payload sent to the create case URL to create a case. You can use variables to add case data to the payload. Required variables are &#x60;case.title&#x60; and &#x60;case.description&#x60;. Due to Mustache template variables (which is the text enclosed in triple braces, for example, &#x60;{{{case.title}}}&#x60;), the JSON is not validated when you create the connector. The JSON is validated after the Mustache variables have been placed when REST method runs. Manually ensure that the JSON is valid to avoid future validation errors; disregard Mustache variables during your review.  | 
**CreateIncidentMethod** | Pointer to **string** | The REST API HTTP request method to create a case in the third-party system. Valid values are &#x60;patch&#x60;, &#x60;post&#x60;, and &#x60;put&#x60;.  | [optional] [default to "post"]
**CreateIncidentResponseKey** | **string** | The JSON key in the create case response that contains the external case ID. | 
**CreateIncidentUrl** | **string** | The REST API URL to create a case in the third-party system. If you are using the &#x60;xpack.actions.allowedHosts&#x60; setting, add the hostname to the allowed hosts.  | 
**GetIncidentResponseExternalTitleKey** | **string** | The JSON key in get case response that contains the external case title. | 
**GetIncidentUrl** | **string** | The REST API URL to get the case by ID from the third-party system. If you are using the &#x60;xpack.actions.allowedHosts&#x60; setting, add the hostname to the allowed hosts. You can use a variable to add the external system ID to the URL. Due to Mustache template variables (the text enclosed in triple braces, for example, &#x60;{{{case.title}}}&#x60;), the JSON is not validated when you create the connector. The JSON is validated after the Mustache variables have been placed when REST method runs. Manually ensure that the JSON is valid, disregarding the Mustache variables, so the later validation will pass.  | 
**HasAuth** | Pointer to **bool** | If true, a username and password for login type authentication must be provided. | [optional] [default to true]
**Headers** | Pointer to **string** | A set of key-value pairs sent as headers with the request URLs for the create case, update case, get case, and create comment methods.  | [optional] 
**UpdateIncidentJson** | **string** | The JSON payload sent to the update case URL to update the case. You can use variables to add Kibana Cases data to the payload. Required variables are &#x60;case.title&#x60; and &#x60;case.description&#x60;. Due to Mustache template variables (which is the text enclosed in triple braces, for example, &#x60;{{{case.title}}}&#x60;), the JSON is not validated when you create the connector. The JSON is validated after the Mustache variables have been placed when REST method runs. Manually ensure that the JSON is valid to avoid future validation errors; disregard Mustache variables during your review.  | 
**UpdateIncidentMethod** | Pointer to **string** | The REST API HTTP request method to update the case in the third-party system. Valid values are &#x60;patch&#x60;, &#x60;post&#x60;, and &#x60;put&#x60;.  | [optional] [default to "put"]
**UpdateIncidentUrl** | **string** | The REST API URL to update the case by ID in the third-party system. You can use a variable to add the external system ID to the URL. If you are using the &#x60;xpack.actions.allowedHosts&#x60; setting, add the hostname to the allowed hosts.  | 
**ViewIncidentUrl** | **string** | The URL to view the case in the external system. You can use variables to add the external system ID or external system title to the URL.  | 

## Methods

### NewConfigPropertiesCasesWebhook

`func NewConfigPropertiesCasesWebhook(createIncidentJson string, createIncidentResponseKey string, createIncidentUrl string, getIncidentResponseExternalTitleKey string, getIncidentUrl string, updateIncidentJson string, updateIncidentUrl string, viewIncidentUrl string, ) *ConfigPropertiesCasesWebhook`

NewConfigPropertiesCasesWebhook instantiates a new ConfigPropertiesCasesWebhook object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewConfigPropertiesCasesWebhookWithDefaults

`func NewConfigPropertiesCasesWebhookWithDefaults() *ConfigPropertiesCasesWebhook`

NewConfigPropertiesCasesWebhookWithDefaults instantiates a new ConfigPropertiesCasesWebhook object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCreateCommentJson

`func (o *ConfigPropertiesCasesWebhook) GetCreateCommentJson() string`

GetCreateCommentJson returns the CreateCommentJson field if non-nil, zero value otherwise.

### GetCreateCommentJsonOk

`func (o *ConfigPropertiesCasesWebhook) GetCreateCommentJsonOk() (*string, bool)`

GetCreateCommentJsonOk returns a tuple with the CreateCommentJson field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateCommentJson

`func (o *ConfigPropertiesCasesWebhook) SetCreateCommentJson(v string)`

SetCreateCommentJson sets CreateCommentJson field to given value.

### HasCreateCommentJson

`func (o *ConfigPropertiesCasesWebhook) HasCreateCommentJson() bool`

HasCreateCommentJson returns a boolean if a field has been set.

### GetCreateCommentMethod

`func (o *ConfigPropertiesCasesWebhook) GetCreateCommentMethod() string`

GetCreateCommentMethod returns the CreateCommentMethod field if non-nil, zero value otherwise.

### GetCreateCommentMethodOk

`func (o *ConfigPropertiesCasesWebhook) GetCreateCommentMethodOk() (*string, bool)`

GetCreateCommentMethodOk returns a tuple with the CreateCommentMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateCommentMethod

`func (o *ConfigPropertiesCasesWebhook) SetCreateCommentMethod(v string)`

SetCreateCommentMethod sets CreateCommentMethod field to given value.

### HasCreateCommentMethod

`func (o *ConfigPropertiesCasesWebhook) HasCreateCommentMethod() bool`

HasCreateCommentMethod returns a boolean if a field has been set.

### GetCreateCommentUrl

`func (o *ConfigPropertiesCasesWebhook) GetCreateCommentUrl() string`

GetCreateCommentUrl returns the CreateCommentUrl field if non-nil, zero value otherwise.

### GetCreateCommentUrlOk

`func (o *ConfigPropertiesCasesWebhook) GetCreateCommentUrlOk() (*string, bool)`

GetCreateCommentUrlOk returns a tuple with the CreateCommentUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateCommentUrl

`func (o *ConfigPropertiesCasesWebhook) SetCreateCommentUrl(v string)`

SetCreateCommentUrl sets CreateCommentUrl field to given value.

### HasCreateCommentUrl

`func (o *ConfigPropertiesCasesWebhook) HasCreateCommentUrl() bool`

HasCreateCommentUrl returns a boolean if a field has been set.

### GetCreateIncidentJson

`func (o *ConfigPropertiesCasesWebhook) GetCreateIncidentJson() string`

GetCreateIncidentJson returns the CreateIncidentJson field if non-nil, zero value otherwise.

### GetCreateIncidentJsonOk

`func (o *ConfigPropertiesCasesWebhook) GetCreateIncidentJsonOk() (*string, bool)`

GetCreateIncidentJsonOk returns a tuple with the CreateIncidentJson field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateIncidentJson

`func (o *ConfigPropertiesCasesWebhook) SetCreateIncidentJson(v string)`

SetCreateIncidentJson sets CreateIncidentJson field to given value.


### GetCreateIncidentMethod

`func (o *ConfigPropertiesCasesWebhook) GetCreateIncidentMethod() string`

GetCreateIncidentMethod returns the CreateIncidentMethod field if non-nil, zero value otherwise.

### GetCreateIncidentMethodOk

`func (o *ConfigPropertiesCasesWebhook) GetCreateIncidentMethodOk() (*string, bool)`

GetCreateIncidentMethodOk returns a tuple with the CreateIncidentMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateIncidentMethod

`func (o *ConfigPropertiesCasesWebhook) SetCreateIncidentMethod(v string)`

SetCreateIncidentMethod sets CreateIncidentMethod field to given value.

### HasCreateIncidentMethod

`func (o *ConfigPropertiesCasesWebhook) HasCreateIncidentMethod() bool`

HasCreateIncidentMethod returns a boolean if a field has been set.

### GetCreateIncidentResponseKey

`func (o *ConfigPropertiesCasesWebhook) GetCreateIncidentResponseKey() string`

GetCreateIncidentResponseKey returns the CreateIncidentResponseKey field if non-nil, zero value otherwise.

### GetCreateIncidentResponseKeyOk

`func (o *ConfigPropertiesCasesWebhook) GetCreateIncidentResponseKeyOk() (*string, bool)`

GetCreateIncidentResponseKeyOk returns a tuple with the CreateIncidentResponseKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateIncidentResponseKey

`func (o *ConfigPropertiesCasesWebhook) SetCreateIncidentResponseKey(v string)`

SetCreateIncidentResponseKey sets CreateIncidentResponseKey field to given value.


### GetCreateIncidentUrl

`func (o *ConfigPropertiesCasesWebhook) GetCreateIncidentUrl() string`

GetCreateIncidentUrl returns the CreateIncidentUrl field if non-nil, zero value otherwise.

### GetCreateIncidentUrlOk

`func (o *ConfigPropertiesCasesWebhook) GetCreateIncidentUrlOk() (*string, bool)`

GetCreateIncidentUrlOk returns a tuple with the CreateIncidentUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateIncidentUrl

`func (o *ConfigPropertiesCasesWebhook) SetCreateIncidentUrl(v string)`

SetCreateIncidentUrl sets CreateIncidentUrl field to given value.


### GetGetIncidentResponseExternalTitleKey

`func (o *ConfigPropertiesCasesWebhook) GetGetIncidentResponseExternalTitleKey() string`

GetGetIncidentResponseExternalTitleKey returns the GetIncidentResponseExternalTitleKey field if non-nil, zero value otherwise.

### GetGetIncidentResponseExternalTitleKeyOk

`func (o *ConfigPropertiesCasesWebhook) GetGetIncidentResponseExternalTitleKeyOk() (*string, bool)`

GetGetIncidentResponseExternalTitleKeyOk returns a tuple with the GetIncidentResponseExternalTitleKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGetIncidentResponseExternalTitleKey

`func (o *ConfigPropertiesCasesWebhook) SetGetIncidentResponseExternalTitleKey(v string)`

SetGetIncidentResponseExternalTitleKey sets GetIncidentResponseExternalTitleKey field to given value.


### GetGetIncidentUrl

`func (o *ConfigPropertiesCasesWebhook) GetGetIncidentUrl() string`

GetGetIncidentUrl returns the GetIncidentUrl field if non-nil, zero value otherwise.

### GetGetIncidentUrlOk

`func (o *ConfigPropertiesCasesWebhook) GetGetIncidentUrlOk() (*string, bool)`

GetGetIncidentUrlOk returns a tuple with the GetIncidentUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGetIncidentUrl

`func (o *ConfigPropertiesCasesWebhook) SetGetIncidentUrl(v string)`

SetGetIncidentUrl sets GetIncidentUrl field to given value.


### GetHasAuth

`func (o *ConfigPropertiesCasesWebhook) GetHasAuth() bool`

GetHasAuth returns the HasAuth field if non-nil, zero value otherwise.

### GetHasAuthOk

`func (o *ConfigPropertiesCasesWebhook) GetHasAuthOk() (*bool, bool)`

GetHasAuthOk returns a tuple with the HasAuth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHasAuth

`func (o *ConfigPropertiesCasesWebhook) SetHasAuth(v bool)`

SetHasAuth sets HasAuth field to given value.

### HasHasAuth

`func (o *ConfigPropertiesCasesWebhook) HasHasAuth() bool`

HasHasAuth returns a boolean if a field has been set.

### GetHeaders

`func (o *ConfigPropertiesCasesWebhook) GetHeaders() string`

GetHeaders returns the Headers field if non-nil, zero value otherwise.

### GetHeadersOk

`func (o *ConfigPropertiesCasesWebhook) GetHeadersOk() (*string, bool)`

GetHeadersOk returns a tuple with the Headers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHeaders

`func (o *ConfigPropertiesCasesWebhook) SetHeaders(v string)`

SetHeaders sets Headers field to given value.

### HasHeaders

`func (o *ConfigPropertiesCasesWebhook) HasHeaders() bool`

HasHeaders returns a boolean if a field has been set.

### GetUpdateIncidentJson

`func (o *ConfigPropertiesCasesWebhook) GetUpdateIncidentJson() string`

GetUpdateIncidentJson returns the UpdateIncidentJson field if non-nil, zero value otherwise.

### GetUpdateIncidentJsonOk

`func (o *ConfigPropertiesCasesWebhook) GetUpdateIncidentJsonOk() (*string, bool)`

GetUpdateIncidentJsonOk returns a tuple with the UpdateIncidentJson field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateIncidentJson

`func (o *ConfigPropertiesCasesWebhook) SetUpdateIncidentJson(v string)`

SetUpdateIncidentJson sets UpdateIncidentJson field to given value.


### GetUpdateIncidentMethod

`func (o *ConfigPropertiesCasesWebhook) GetUpdateIncidentMethod() string`

GetUpdateIncidentMethod returns the UpdateIncidentMethod field if non-nil, zero value otherwise.

### GetUpdateIncidentMethodOk

`func (o *ConfigPropertiesCasesWebhook) GetUpdateIncidentMethodOk() (*string, bool)`

GetUpdateIncidentMethodOk returns a tuple with the UpdateIncidentMethod field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateIncidentMethod

`func (o *ConfigPropertiesCasesWebhook) SetUpdateIncidentMethod(v string)`

SetUpdateIncidentMethod sets UpdateIncidentMethod field to given value.

### HasUpdateIncidentMethod

`func (o *ConfigPropertiesCasesWebhook) HasUpdateIncidentMethod() bool`

HasUpdateIncidentMethod returns a boolean if a field has been set.

### GetUpdateIncidentUrl

`func (o *ConfigPropertiesCasesWebhook) GetUpdateIncidentUrl() string`

GetUpdateIncidentUrl returns the UpdateIncidentUrl field if non-nil, zero value otherwise.

### GetUpdateIncidentUrlOk

`func (o *ConfigPropertiesCasesWebhook) GetUpdateIncidentUrlOk() (*string, bool)`

GetUpdateIncidentUrlOk returns a tuple with the UpdateIncidentUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateIncidentUrl

`func (o *ConfigPropertiesCasesWebhook) SetUpdateIncidentUrl(v string)`

SetUpdateIncidentUrl sets UpdateIncidentUrl field to given value.


### GetViewIncidentUrl

`func (o *ConfigPropertiesCasesWebhook) GetViewIncidentUrl() string`

GetViewIncidentUrl returns the ViewIncidentUrl field if non-nil, zero value otherwise.

### GetViewIncidentUrlOk

`func (o *ConfigPropertiesCasesWebhook) GetViewIncidentUrlOk() (*string, bool)`

GetViewIncidentUrlOk returns a tuple with the ViewIncidentUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetViewIncidentUrl

`func (o *ConfigPropertiesCasesWebhook) SetViewIncidentUrl(v string)`

SetViewIncidentUrl sets ViewIncidentUrl field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


