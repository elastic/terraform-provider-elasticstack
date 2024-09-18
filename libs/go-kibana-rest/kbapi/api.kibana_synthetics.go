package kbapi

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

const (
	basePathKibanaSynthetics = "/api/synthetics"
	privateLocationsSuffix   = "/private_locations"
	monitorsSuffix           = "/monitors"

	Http    MonitorType = "http"
	Tcp     MonitorType = "tcp"
	Icmp    MonitorType = "icmp"
	Browser MonitorType = "browser"

	Every1Minute    MonitorSchedule = 1
	Every2Minutes                   = 2
	Every3Minutes                   = 3
	Every5Minutes                   = 5
	Every10Minutes                  = 10
	Every15Minutes                  = 15
	Every20Minutes                  = 20
	Every30Minutes                  = 30
	Every60Minutes                  = 60
	Every120Minutes                 = 120
	Every240Minutes                 = 240

	Japan         MonitorLocation = "japan"
	India                         = "india"
	Singapore                     = "singapore"
	AustraliaEast                 = "australia_east"
	UnitedKingdom                 = "united_kingdom"
	Germany                       = "germany"
	CanadaEast                    = "canada_east"
	Brazil                        = "brazil"
	USEast                        = "us_east"
	USWest                        = "us_west"

	ModeAll HttpMonitorMode = "all"
	ModeAny                 = "any"

	ScreenshotOn             ScreenshotOption = "on"
	ScreenshotOff                             = "off"
	ScreenshotOnlyOfFailures                  = "only-on-failure"
)

var plMu sync.Mutex

type MonitorFields interface {
	APIRequest(cfg SyntheticsMonitorConfig) interface{}
}

type KibanaError struct {
	Code    int    `json:"statusCode,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

type MonitorID string
type MonitorType string
type MonitorLocation string
type MonitorSchedule int
type HttpMonitorMode string
type ScreenshotOption string

type JsonObject map[string]interface{}

type KibanaSyntheticsMonitorAPI struct {
	Add    KibanaSyntheticsMonitorAdd
	Delete KibanaSyntheticsMonitorDelete
	Get    KibanaSyntheticsMonitorGet
	Update KibanaSyntheticsMonitorUpdate
}

type KibanaSyntheticsPrivateLocationAPI struct {
	Create KibanaSyntheticsPrivateLocationCreate
	Delete KibanaSyntheticsPrivateLocationDelete
	Get    KibanaSyntheticsPrivateLocationGet
}

type SyntheticsStatusConfig struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type MonitorAlertConfig struct {
	Status *SyntheticsStatusConfig `json:"status,omitempty"`
	Tls    *SyntheticsStatusConfig `json:"tls,omitempty"`
}

type ICMPMonitorFields struct {
	Host string `json:"host"`
	Wait string `json:"wait,omitempty"`
}

type BrowserMonitorFields struct {
	InlineScript      string           `json:"inline_script"`
	Screenshots       ScreenshotOption `json:"screenshots,omitempty"`
	SyntheticsArgs    []string         `json:"synthetics_args,omitempty"`
	IgnoreHttpsErrors *bool            `json:"ignore_https_errors,omitempty"`
	PlaywrightOptions JsonObject       `json:"playwright_options,omitempty"`
}

type TCPMonitorFields struct {
	Host                  string   `json:"host"`
	SslVerificationMode   string   `json:"ssl.verification_mode,omitempty"`
	SslSupportedProtocols []string `json:"ssl.supported_protocols,omitempty"`
	CheckSend             string   `json:"check.send,omitempty"`
	CheckReceive          string   `json:"check.receive,omitempty"`
	ProxyUrl              string   `json:"proxy_url,omitempty"`
	ProxyUseLocalResolver *bool    `json:"proxy_use_local_resolver,omitempty"`
}

type HTTPMonitorFields struct {
	Url                   string          `json:"url"`
	SslVerificationMode   string          `json:"ssl.verification_mode,omitempty"`
	SslSupportedProtocols []string        `json:"ssl.supported_protocols,omitempty"`
	MaxRedirects          string          `json:"max_redirects,omitempty"`
	Mode                  HttpMonitorMode `json:"mode,omitempty"`
	Ipv4                  *bool           `json:"ipv4,omitempty"`
	Ipv6                  *bool           `json:"ipv6,omitempty"`
	Username              string          `json:"username,omitempty"`
	Password              string          `json:"password,omitempty"`
	ProxyHeader           JsonObject      `json:"proxy_headers,omitempty"`
	ProxyUrl              string          `json:"proxy_url,omitempty"`
	Response              JsonObject      `json:"response,omitempty"`
	Check                 JsonObject      `json:"check,omitempty"`
}

type SyntheticsMonitorConfig struct {
	Name             string              `json:"name"`
	Schedule         MonitorSchedule     `json:"schedule,omitempty"`
	Locations        []MonitorLocation   `json:"locations,omitempty"`
	PrivateLocations []string            `json:"private_locations,omitempty"`
	Enabled          *bool               `json:"enabled,omitempty"`
	Tags             []string            `json:"tags,omitempty"`
	Alert            *MonitorAlertConfig `json:"alert,omitempty"`
	APMServiceName   string              `json:"service.name,omitempty"`
	TimeoutSeconds   int                 `json:"timeout,omitempty"`
	Namespace        string              `json:"namespace,omitempty"`
	Params           JsonObject          `json:"params,omitempty"`
	RetestOnFailure  *bool               `json:"retest_on_failure,omitempty"`
}

type MonitorScheduleConfig struct {
	Number string `json:"number"`
	Unit   string `json:"unit"`
}

type SyntheticGeoConfig struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type MonitorLocationConfig struct {
	Id               string              `json:"id"`
	Label            string              `json:"label"`
	Geo              *SyntheticGeoConfig `json:"geo,omitempty"`
	IsServiceManaged bool                `json:"isServiceManaged"`
}

type PrivateLocationConfig struct {
	Label         string              `json:"label"`
	AgentPolicyId string              `json:"agentPolicyId"`
	Tags          []string            `json:"tags,omitempty"`
	Geo           *SyntheticGeoConfig `json:"geo,omitempty"`
}

type PrivateLocation struct {
	Id        string `json:"id"`
	Namespace string `json:"namespace,omitempty"`
	PrivateLocationConfig
}

type MonitorDeleteStatus struct {
	Id      MonitorID `json:"id"`
	Deleted bool      `json:"deleted"`
}

type SyntheticsMonitor struct {
	Name           string                  `json:"name"`
	Type           MonitorType             `json:"type"`
	ConfigId       MonitorID               `json:"config_id"`
	Id             MonitorID               `json:"id"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	Namespace      string                  `json:"namespace"`
	Enabled        *bool                   `json:"enabled,omitempty"`
	Alert          *MonitorAlertConfig     `json:"alert,omitempty"`
	Schedule       *MonitorScheduleConfig  `json:"schedule,omitempty"`
	Tags           []string                `json:"tags,omitempty"`
	APMServiceName string                  `json:"service.name,omitempty"`
	Timeout        json.Number             `json:"timeout,omitempty"`
	Locations      []MonitorLocationConfig `json:"locations,omitempty"`
	Origin         string                  `json:"origin,omitempty"`
	Params         JsonObject              `json:"params,omitempty"`
	MaxAttempts    int                     `json:"max_attempts"`
	Revision       int                     `json:"revision,omitempty"`
	Ui             JsonObject              `json:"__ui,omitempty"`
	//http
	Url                         string          `json:"url,omitempty"`
	Mode                        HttpMonitorMode `json:"mode"`
	MaxRedirects                string          `json:"max_redirects"`
	Ipv4                        *bool           `json:"ipv4,omitempty"`
	Ipv6                        *bool           `json:"ipv6,omitempty"`
	Username                    string          `json:"username,omitempty"`
	Password                    string          `json:"password,omitempty"`
	ProxyHeaders                JsonObject      `json:"proxy_headers,omitempty"`
	CheckResponseBodyPositive   []string        `json:"check.response.body.positive,omitempty"`
	CheckResponseStatus         []string        `json:"check.response.status,omitempty"`
	ResponseIncludeBody         string          `json:"response.include_body,omitempty"`
	ResponseIncludeHeaders      bool            `json:"response.include_headers,omitempty"`
	ResponseIncludeBodyMaxBytes string          `json:"response.include_body_max_bytes,omitempty"`
	CheckRequestBody            JsonObject      `json:"check.request.body,omitempty"`
	CheckRequestHeaders         JsonObject      `json:"check.request.headers,omitempty"`
	CheckRequestMethod          string          `json:"check.request.method,omitempty"`
	//http and tcp
	ProxyUrl              string   `json:"proxy_url,omitempty"`
	SslVerificationMode   string   `json:"ssl.verification_mode"`
	SslSupportedProtocols []string `json:"ssl.supported_protocols"`
	//tcp and icmp
	Host string `json:"host,omitempty"`
	//tcp
	ProxyUseLocalResolver *bool  `json:"proxy_use_local_resolver,omitempty"`
	CheckSend             string `json:"check.send,omitempty"`
	CheckReceive          string `json:"check.receive,omitempty"`
	//icmp
	Wait json.Number `json:"wait,omitempty"`
	//browser
	Screenshots       string     `json:"screenshots,omitempty"`
	IgnoreHttpsErrors *bool      `json:"ignore_https_errors,omitempty"`
	InlineScript      string     `json:"inline_script"`
	SyntheticsArgs    []string   `json:"synthetics_args,omitempty"`
	PlaywrightOptions JsonObject `json:"playwright_options,omitempty"`
}

type MonitorTypeConfig struct {
	Type MonitorType `json:"type"`
}

func (f HTTPMonitorFields) APIRequest(config SyntheticsMonitorConfig) interface{} {

	mType := MonitorTypeConfig{Type: Http}

	return struct {
		SyntheticsMonitorConfig
		MonitorTypeConfig
		HTTPMonitorFields
	}{
		config,
		mType,
		f,
	}
}

func (f TCPMonitorFields) APIRequest(config SyntheticsMonitorConfig) interface{} {

	mType := MonitorTypeConfig{Type: Tcp}

	return struct {
		SyntheticsMonitorConfig
		MonitorTypeConfig
		TCPMonitorFields
	}{
		config,
		mType,
		f,
	}
}

func (f ICMPMonitorFields) APIRequest(config SyntheticsMonitorConfig) interface{} {

	mType := MonitorTypeConfig{Type: Icmp}

	return struct {
		SyntheticsMonitorConfig
		MonitorTypeConfig
		ICMPMonitorFields
	}{
		config,
		mType,
		f,
	}
}

func (f BrowserMonitorFields) APIRequest(config SyntheticsMonitorConfig) interface{} {

	mType := MonitorTypeConfig{Type: Browser}

	return struct {
		SyntheticsMonitorConfig
		MonitorTypeConfig
		BrowserMonitorFields
	}{
		config,
		mType,
		f,
	}
}

type KibanaSyntheticsMonitorAdd func(ctx context.Context, config SyntheticsMonitorConfig, fields MonitorFields, namespace string) (*SyntheticsMonitor, error)

type KibanaSyntheticsMonitorUpdate func(ctx context.Context, id MonitorID, config SyntheticsMonitorConfig, fields MonitorFields, namespace string) (*SyntheticsMonitor, error)

type KibanaSyntheticsMonitorGet func(ctx context.Context, id MonitorID, namespace string) (*SyntheticsMonitor, error)

type KibanaSyntheticsMonitorDelete func(ctx context.Context, namespace string, ids ...MonitorID) ([]MonitorDeleteStatus, error)

type KibanaSyntheticsPrivateLocationCreate func(ctx context.Context, pLoc PrivateLocationConfig) (*PrivateLocation, error)

type KibanaSyntheticsPrivateLocationGet func(ctx context.Context, idOrLabel string) (*PrivateLocation, error)

type KibanaSyntheticsPrivateLocationDelete func(ctx context.Context, id string) error

func newKibanaSyntheticsPrivateLocationGetFunc(c *resty.Client) KibanaSyntheticsPrivateLocationGet {
	return func(ctx context.Context, idOrLabel string) (*PrivateLocation, error) {
		if idOrLabel == "" {
			return nil, APIError{
				Code:    404,
				Message: "Private location id or label is empty",
			}
		}

		path := basePathWithId("", privateLocationsSuffix, idOrLabel)
		log.Debugf("URL to get private locations: %s", path)
		resp, err := c.R().SetContext(ctx).Get(path)
		if err = handleKibanaError(err, resp); err != nil {
			return nil, err
		}
		return unmarshal(resp, PrivateLocation{})
	}
}

func newKibanaSyntheticsPrivateLocationCreateFunc(c *resty.Client) KibanaSyntheticsPrivateLocationCreate {
	return func(ctx context.Context, pLoc PrivateLocationConfig) (*PrivateLocation, error) {
		plMu.Lock()
		defer plMu.Unlock()

		path := basePath("", privateLocationsSuffix)
		log.Debugf("URL to create private locations: %s", path)
		resp, err := c.R().SetContext(ctx).SetBody(pLoc).Post(path)
		if err = handleKibanaError(err, resp); err != nil {
			return nil, err
		}
		return unmarshal(resp, PrivateLocation{})
	}
}

func newKibanaSyntheticsPrivateLocationDeleteFunc(c *resty.Client) KibanaSyntheticsPrivateLocationDelete {
	return func(ctx context.Context, id string) error {
		plMu.Lock()
		defer plMu.Unlock()

		path := basePathWithId("", privateLocationsSuffix, id)
		log.Debugf("URL to delete private locations: %s", path)
		resp, err := c.R().SetContext(ctx).Delete(path)
		err = handleKibanaError(err, resp)
		return err
	}
}

func newKibanaSyntheticsMonitorGetFunc(c *resty.Client) KibanaSyntheticsMonitorGet {
	return func(ctx context.Context, id MonitorID, namespace string) (*SyntheticsMonitor, error) {
		path := basePathWithId(namespace, monitorsSuffix, id)
		log.Debugf("URL to get monitor: %s", path)

		resp, err := c.R().SetContext(ctx).Get(path)
		if err := handleKibanaError(err, resp); err != nil {
			return nil, err
		}
		return unmarshal(resp, SyntheticsMonitor{})
	}
}

func newKibanaSyntheticsMonitorDeleteFunc(c *resty.Client) KibanaSyntheticsMonitorDelete {
	return func(ctx context.Context, namespace string, ids ...MonitorID) ([]MonitorDeleteStatus, error) {
		path := basePath(namespace, monitorsSuffix)
		log.Debugf("URL to delete monitors: %s", path)

		resp, err := c.R().SetContext(ctx).SetBody(map[string]interface{}{
			"ids": ids,
		}).Delete(path)
		if err = handleKibanaError(err, resp); err != nil {
			return nil, err
		}

		result, err := unmarshal(resp, []MonitorDeleteStatus{})
		return *result, err
	}
}

func newKibanaSyntheticsMonitorUpdateFunc(c *resty.Client) KibanaSyntheticsMonitorUpdate {
	return func(ctx context.Context, id MonitorID, config SyntheticsMonitorConfig, fields MonitorFields, namespace string) (*SyntheticsMonitor, error) {

		path := basePathWithId(namespace, monitorsSuffix, id)
		log.Debugf("URL to update monitor: %s", path)
		data := fields.APIRequest(config)
		resp, err := c.R().SetContext(ctx).SetBody(data).Put(path)
		if err := handleKibanaError(err, resp); err != nil {
			return nil, err
		}
		return unmarshal(resp, SyntheticsMonitor{})
	}
}

func newKibanaSyntheticsMonitorAddFunc(c *resty.Client) KibanaSyntheticsMonitorAdd {
	return func(ctx context.Context, config SyntheticsMonitorConfig, fields MonitorFields, namespace string) (*SyntheticsMonitor, error) {

		path := basePath(namespace, monitorsSuffix)
		log.Debugf("URL to create monitor: %s", path)
		data := fields.APIRequest(config)
		resp, err := c.R().SetContext(ctx).SetBody(data).Post(path)
		if err := handleKibanaError(err, resp); err != nil {
			return nil, err
		}
		return unmarshal(resp, SyntheticsMonitor{})
	}
}

func unmarshal[T interface{}](resp *resty.Response, result T) (*T, error) {
	respBody := resp.Body()
	err := json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func handleKibanaError(err error, resp *resty.Response) error {
	if err != nil {
		return err
	}
	log.Debug("Response: ", resp)
	if resp.StatusCode() >= 300 {
		kibanaErr := KibanaError{}
		err := json.Unmarshal(resp.Body(), &kibanaErr)
		if err != nil {
			return NewAPIError(resp.StatusCode(), resp.Status(), err)
		}
		return NewAPIError(resp.StatusCode(), kibanaErr.Message, kibanaErr.Error)
	}
	return nil
}

func basePathWithId(namespace, suffix string, id any) string {
	return fmt.Sprintf("%s/%s", basePath(namespace, suffix), id)
}

func basePath(namespace, suffix string) string {
	return namespaceBasesPath(namespace, basePathKibanaSynthetics, suffix)
}

func namespaceBasesPath(namespace, basePath, suffix string) string {
	if namespace == "" || namespace == "default" {
		return fmt.Sprintf("%s%s", basePath, suffix)
	}

	return fmt.Sprintf("/s/%s%s%s", namespace, basePath, suffix)
}
