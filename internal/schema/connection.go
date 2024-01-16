package schema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func GetEsFWConnectionBlock(keyName string) fwschema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName("username")
	passwordPath := path.MatchRelative().AtParent().AtName("password")
	apiKeyPath := path.MatchRelative().AtParent().AtName("api_key")
	bearerTokenPath := path.MatchRelative().AtParent().AtName("bearer_token")
	caFilePath := path.MatchRelative().AtParent().AtName("ca_file")
	caDataPath := path.MatchRelative().AtParent().AtName("ca_data")
	certFilePath := path.MatchRelative().AtParent().AtName("cert_file")
	certDataPath := path.MatchRelative().AtParent().AtName("cert_data")
	keyFilePath := path.MatchRelative().AtParent().AtName("key_file")
	keyDataPath := path.MatchRelative().AtParent().AtName("key_data")

	return fwschema.ListNestedBlock{
		MarkdownDescription: "Elasticsearch connection configuration block. ",
		Description:         "Elasticsearch connection configuration block. ",
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"username": fwschema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Elasticsearch.",
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				"password": fwschema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Elasticsearch.",
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				"api_key": fwschema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Elasticsearch",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, bearerTokenPath),
					},
				},
				"bearer_token": fwschema.StringAttribute{
					MarkdownDescription: "Bearer Token to use for authentication to Elasticsearch",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
					},
				},
				"es_client_authentication": fwschema.StringAttribute{
					MarkdownDescription: "ES Client Authentication field to be used with the bearer token",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath, apiKeyPath),
						stringvalidator.AlsoRequires(bearerTokenPath),
					},
				},
				"endpoints": fwschema.ListAttribute{
					MarkdownDescription: "A list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				"insecure": fwschema.BoolAttribute{
					MarkdownDescription: "Disable TLS certificate validation",
					Optional:            true,
				},
				"ca_file": fwschema.StringAttribute{
					MarkdownDescription: "Path to a custom Certificate Authority certificate",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(caDataPath),
					},
				},
				"ca_data": fwschema.StringAttribute{
					MarkdownDescription: "PEM-encoded custom Certificate Authority certificate",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(caFilePath),
					},
				},
				"cert_file": fwschema.StringAttribute{
					MarkdownDescription: "Path to a file containing the PEM encoded certificate for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(keyFilePath),
						stringvalidator.ConflictsWith(caDataPath, keyDataPath),
					},
				},
				"key_file": fwschema.StringAttribute{
					MarkdownDescription: "Path to a file containing the PEM encoded private key for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(certFilePath),
						stringvalidator.ConflictsWith(certDataPath, keyDataPath),
					},
				},
				"cert_data": fwschema.StringAttribute{
					MarkdownDescription: "PEM encoded certificate for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(keyDataPath),
						stringvalidator.ConflictsWith(certFilePath, keyFilePath),
					},
				},
				"key_data": fwschema.StringAttribute{
					MarkdownDescription: "PEM encoded private key for client auth",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(certDataPath),
						stringvalidator.ConflictsWith(certFilePath, keyFilePath),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

func GetKbFWConnectionBlock() fwschema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName("username")
	passwordPath := path.MatchRelative().AtParent().AtName("password")

	return fwschema.ListNestedBlock{
		MarkdownDescription: "Kibana connection configuration block.",
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"api_key": fwschema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Kibana",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath, passwordPath),
					},
				},
				"username": fwschema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Kibana.",
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				"password": fwschema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Kibana.",
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				"endpoints": fwschema.ListAttribute{
					MarkdownDescription: "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Optional:            true,
					Sensitive:           true,
					ElementType:         types.StringType,
				},
				"ca_certs": fwschema.ListAttribute{
					MarkdownDescription: "A list of paths to CA certificates to validate the certificate presented by the Kibana server.",
					Optional:            true,
					ElementType:         types.StringType,
				},
				"insecure": fwschema.BoolAttribute{
					MarkdownDescription: "Disable TLS certificate validation",
					Optional:            true,
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

func GetFleetFWConnectionBlock() fwschema.Block {
	usernamePath := path.MatchRelative().AtParent().AtName("username")
	passwordPath := path.MatchRelative().AtParent().AtName("password")

	return fwschema.ListNestedBlock{
		MarkdownDescription: "Fleet connection configuration block.",
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"username": fwschema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Fleet.",
					Optional:            true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(passwordPath)},
				},
				"password": fwschema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Fleet.",
					Optional:            true,
					Sensitive:           true,
					Validators:          []validator.String{stringvalidator.AlsoRequires(usernamePath)},
				},
				"api_key": fwschema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Fleet.",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(usernamePath),
						stringvalidator.ConflictsWith(passwordPath),
					},
				},
				"endpoint": fwschema.StringAttribute{
					MarkdownDescription: "The Fleet server where the terraform provider will point to, this must include the http(s) schema and port number.",
					Optional:            true,
					Sensitive:           true,
				},
				"ca_certs": fwschema.ListAttribute{
					MarkdownDescription: "A list of paths to CA certificates to validate the certificate presented by the Fleet server.",
					Optional:            true,
					ElementType:         types.StringType,
				},
				"insecure": fwschema.BoolAttribute{
					MarkdownDescription: "Disable TLS certificate validation",
					Optional:            true,
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

func GetEsConnectionSchema(keyName string, isProviderConfiguration bool) *schema.Schema {
	usernamePath := makePathRef(keyName, "username")
	passwordPath := makePathRef(keyName, "password")
	apiKeyPath := makePathRef(keyName, "api_key")
	bearerTokenPath := makePathRef(keyName, "bearer_token")
	caFilePath := makePathRef(keyName, "ca_file")
	caDataPath := makePathRef(keyName, "ca_data")
	certFilePath := makePathRef(keyName, "cert_file")
	certDataPath := makePathRef(keyName, "cert_data")
	keyFilePath := makePathRef(keyName, "key_file")
	keyDataPath := makePathRef(keyName, "key_data")

	usernameRequiredWithValidation := []string{passwordPath}
	passwordRequiredWithValidation := []string{usernamePath}

	withEnvDefault := func(key string, dv interface{}) schema.SchemaDefaultFunc { return nil }

	if isProviderConfiguration {
		withEnvDefault = func(key string, dv interface{}) schema.SchemaDefaultFunc { return schema.EnvDefaultFunc(key, dv) }

		// RequireWith validation isn't compatible when used in conjunction with DefaultFunc
		usernameRequiredWithValidation = nil
		passwordRequiredWithValidation = nil
	}

	return &schema.Schema{
		Description: fmt.Sprintf("Elasticsearch connection configuration block. %s", getDeprecationMessage(isProviderConfiguration)),
		Deprecated:  getDeprecationMessage(isProviderConfiguration),
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"username": {
					Description:  "Username to use for API authentication to Elasticsearch.",
					Type:         schema.TypeString,
					Optional:     true,
					DefaultFunc:  withEnvDefault("ELASTICSEARCH_USERNAME", nil),
					RequiredWith: usernameRequiredWithValidation,
				},
				"password": {
					Description:  "Password to use for API authentication to Elasticsearch.",
					Type:         schema.TypeString,
					Optional:     true,
					Sensitive:    true,
					DefaultFunc:  withEnvDefault("ELASTICSEARCH_PASSWORD", nil),
					RequiredWith: passwordRequiredWithValidation,
				},
				"api_key": {
					Description:   "API Key to use for authentication to Elasticsearch",
					Type:          schema.TypeString,
					Optional:      true,
					Sensitive:     true,
					DefaultFunc:   withEnvDefault("ELASTICSEARCH_API_KEY", nil),
					ConflictsWith: []string{usernamePath, passwordPath, bearerTokenPath},
				},
				"bearer_token": {
					Description:   "Bearer Token to use for authentication to Elasticsearch",
					Type:          schema.TypeString,
					Optional:      true,
					Sensitive:     true,
					DefaultFunc:   withEnvDefault("ELASTICSEARCH_BEARER_TOKEN", nil),
					ConflictsWith: []string{usernamePath, passwordPath, apiKeyPath},
				},
				"es_client_authentication": {
					Description: "ES Client Authentication field to be used with the bearer token",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: withEnvDefault("ELASTICSEARCH_ES_CLIENT_AUTHENTICATION", nil),
				},
				"endpoints": {
					Description: "A list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Type:        schema.TypeList,
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"insecure": {
					Description: "Disable TLS certificate validation",
					Type:        schema.TypeBool,
					Optional:    true,
					DefaultFunc: withEnvDefault("ELASTICSEARCH_INSECURE", false),
				},
				"ca_file": {
					Description:   "Path to a custom Certificate Authority certificate",
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{caDataPath},
				},
				"ca_data": {
					Description:   "PEM-encoded custom Certificate Authority certificate",
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{caFilePath},
				},
				"cert_file": {
					Description:   "Path to a file containing the PEM encoded certificate for client auth",
					Type:          schema.TypeString,
					Optional:      true,
					RequiredWith:  []string{keyFilePath},
					ConflictsWith: []string{certDataPath, keyDataPath},
				},
				"key_file": {
					Description:   "Path to a file containing the PEM encoded private key for client auth",
					Type:          schema.TypeString,
					Optional:      true,
					RequiredWith:  []string{certFilePath},
					ConflictsWith: []string{certDataPath, keyDataPath},
				},
				"cert_data": {
					Description:   "PEM encoded certificate for client auth",
					Type:          schema.TypeString,
					Optional:      true,
					RequiredWith:  []string{keyDataPath},
					ConflictsWith: []string{certFilePath, keyFilePath},
				},
				"key_data": {
					Description:   "PEM encoded private key for client auth",
					Type:          schema.TypeString,
					Optional:      true,
					Sensitive:     true,
					RequiredWith:  []string{certDataPath},
					ConflictsWith: []string{certFilePath, keyFilePath},
				},
			},
		},
	}
}

func GetKibanaConnectionSchema() *schema.Schema {
	withEnvDefault := func(key string, dv interface{}) schema.SchemaDefaultFunc { return nil }
	return &schema.Schema{
		Description: "Kibana connection configuration block.",
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"api_key": {
					Description:   "API Key to use for authentication to Kibana",
					Type:          schema.TypeString,
					Optional:      true,
					Sensitive:     true,
					DefaultFunc:   withEnvDefault("KIBANA_API_KEY", nil),
					ConflictsWith: []string{"kibana.0.password", "kibana.0.username"},
				},
				"username": {
					Description:  "Username to use for API authentication to Kibana.",
					Type:         schema.TypeString,
					Optional:     true,
					RequiredWith: []string{"kibana.0.password"},
				},
				"password": {
					Description:  "Password to use for API authentication to Kibana.",
					Type:         schema.TypeString,
					Optional:     true,
					Sensitive:    true,
					RequiredWith: []string{"kibana.0.username"},
				},
				"endpoints": {
					Description: "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Type:        schema.TypeList,
					Optional:    true,
					Sensitive:   true,
					MaxItems:    1, // Current API restriction
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"ca_certs": {
					Description: "A list of paths to CA certificates to validate the certificate presented by the Kibana server.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"insecure": {
					Description: "Disable TLS certificate validation",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
				},
			},
		},
	}
}

func GetFleetConnectionSchema() *schema.Schema {
	return &schema.Schema{
		Description: "Fleet connection configuration block.",
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"username": {
					Description:  "Username to use for API authentication to Fleet.",
					Type:         schema.TypeString,
					Optional:     true,
					RequiredWith: []string{"fleet.0.password"},
				},
				"password": {
					Description:  "Password to use for API authentication to Fleet.",
					Type:         schema.TypeString,
					Optional:     true,
					Sensitive:    true,
					RequiredWith: []string{"fleet.0.username"},
				},
				"api_key": {
					Description: "API Key to use for authentication to Fleet.",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
				},
				"endpoint": {
					Description: "The Fleet server where the terraform provider will point to, this must include the http(s) schema and port number.",
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
				},
				"ca_certs": {
					Description: "A list of paths to CA certificates to validate the certificate presented by the Fleet server.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"insecure": {
					Description: "Disable TLS certificate validation",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
				},
			},
		},
	}
}

func makePathRef(keyName string, keyValue string) string {
	return fmt.Sprintf("%s.0.%s", keyName, keyValue)
}

func getDeprecationMessage(isProviderConfiguration bool) string {
	if isProviderConfiguration {
		return ""
	}
	return "This property will be removed in a future provider version. Configure the Elasticsearch connection via the provider configuration instead."
}
