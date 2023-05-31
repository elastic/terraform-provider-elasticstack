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

func GetEsFWConnectionBlock(keyName string, isProviderConfiguration bool) fwschema.Block {
	usernamePath := makePathRef(keyName, "username")
	passwordPath := makePathRef(keyName, "password")
	caFilePath := makePathRef(keyName, "ca_file")
	caDataPath := makePathRef(keyName, "ca_data")
	certFilePath := makePathRef(keyName, "cert_file")
	certDataPath := makePathRef(keyName, "cert_data")
	keyFilePath := makePathRef(keyName, "key_file")
	keyDataPath := makePathRef(keyName, "key_data")

	usernameValidators := []validator.String{stringvalidator.AlsoRequires(path.MatchRoot(passwordPath))}
	passwordValidators := []validator.String{stringvalidator.AlsoRequires(path.MatchRoot(usernamePath))}

	if isProviderConfiguration {
		// RequireWith validation isn't compatible when used in conjunction with DefaultFunc
		usernameValidators = nil
		passwordValidators = nil
	}

	return fwschema.ListNestedBlock{
		MarkdownDescription: fmt.Sprintf("Elasticsearch connection configuration block. %s", getDeprecationMessage(isProviderConfiguration)),
		DeprecationMessage:  getDeprecationMessage(isProviderConfiguration),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"username": fwschema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Elasticsearch.",
					Optional:            true,
					Validators:          usernameValidators,
				},
				"password": fwschema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Elasticsearch.",
					Optional:            true,
					Sensitive:           true,
					Validators:          passwordValidators,
				},
				"api_key": fwschema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Elasticsearch",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("username")),
						stringvalidator.ConflictsWith(path.MatchRoot(passwordPath)),
					},
				},
				"endpoints": fwschema.ListAttribute{
					MarkdownDescription: "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
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
						stringvalidator.ConflictsWith(path.MatchRoot(caDataPath)),
					},
				},
				"ca_data": fwschema.StringAttribute{
					MarkdownDescription: "PEM-encoded custom Certificate Authority certificate",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.MatchRoot(caFilePath)),
					},
				},
				"cert_file": fwschema.StringAttribute{
					MarkdownDescription: "Path to a file containing the PEM encoded certificate for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(path.MatchRoot(keyFilePath)),
						stringvalidator.ConflictsWith(path.MatchRoot(certDataPath)),
						stringvalidator.ConflictsWith(path.MatchRoot(keyDataPath)),
					},
				},
				"key_file": fwschema.StringAttribute{
					MarkdownDescription: "Path to a file containing the PEM encoded private key for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(path.MatchRoot(certFilePath)),
						stringvalidator.ConflictsWith(path.MatchRoot(certDataPath)),
						stringvalidator.ConflictsWith(path.MatchRoot(keyDataPath)),
					},
				},
				"cert_data": fwschema.StringAttribute{
					MarkdownDescription: "PEM encoded certificate for client auth",
					Optional:            true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(path.MatchRoot(keyDataPath)),
						stringvalidator.ConflictsWith(path.MatchRoot(certFilePath)),
						stringvalidator.ConflictsWith(path.MatchRoot(keyFilePath)),
					},
				},
				"key_data": fwschema.StringAttribute{
					MarkdownDescription: "PEM encoded private key for client auth",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.AlsoRequires(path.MatchRoot(certDataPath)),
						stringvalidator.ConflictsWith(path.MatchRoot(certFilePath)),
						stringvalidator.ConflictsWith(path.MatchRoot(keyFilePath)),
					},
				},
			},
		},
		Validators: []validator.List{
			listvalidator.SizeAtMost(1),
		},
	}
}

func GetKbFWConnectionBlock(keyName string, isProviderConfiguration bool) fwschema.Block {
	usernamePath := makePathRef(keyName, "username")
	passwordPath := makePathRef(keyName, "password")

	usernameValidators := []validator.String{stringvalidator.AlsoRequires(path.MatchRoot(passwordPath))}
	passwordValidators := []validator.String{stringvalidator.AlsoRequires(path.MatchRoot(usernamePath))}

	return fwschema.ListNestedBlock{
		MarkdownDescription: fmt.Sprintf("Kibana connection configuration block. %s", getDeprecationMessage(isProviderConfiguration)),
		DeprecationMessage:  getDeprecationMessage(isProviderConfiguration),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"username": fwschema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Kibana.",
					Optional:            true,
					Validators:          usernameValidators,
				},
				"password": fwschema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Kibana.",
					Optional:            true,
					Sensitive:           true,
					Validators:          passwordValidators,
				},
				"endpoints": fwschema.ListAttribute{
					MarkdownDescription: "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Optional:            true,
					Sensitive:           true,
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

func GetFleetFWConnectionBlock(keyName string, isProviderConfiguration bool) fwschema.Block {
	usernamePath := makePathRef(keyName, "username")
	passwordPath := makePathRef(keyName, "password")

	usernameValidators := []validator.String{stringvalidator.AlsoRequires(path.MatchRoot(passwordPath))}
	passwordValidators := []validator.String{stringvalidator.AlsoRequires(path.MatchRoot(usernamePath))}

	return fwschema.ListNestedBlock{
		MarkdownDescription: fmt.Sprintf("Kibana connection configuration block. %s", getDeprecationMessage(isProviderConfiguration)),
		DeprecationMessage:  getDeprecationMessage(isProviderConfiguration),
		NestedObject: fwschema.NestedBlockObject{
			Attributes: map[string]fwschema.Attribute{
				"username": fwschema.StringAttribute{
					MarkdownDescription: "Username to use for API authentication to Fleet.",
					Optional:            true,
					Validators:          usernameValidators,
				},
				"password": fwschema.StringAttribute{
					MarkdownDescription: "Password to use for API authentication to Fleet.",
					Optional:            true,
					Sensitive:           true,
					Validators:          passwordValidators,
				},
				"api_key": fwschema.StringAttribute{
					MarkdownDescription: "API Key to use for authentication to Fleet.",
					Optional:            true,
					Sensitive:           true,
					Validators: []validator.String{
						stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("username")),
						stringvalidator.ConflictsWith(path.MatchRoot(passwordPath)),
					},
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
					Description:  fmt.Sprintf("Elasticsearch connection configuration block. %s", getDeprecationMessage(isProviderConfiguration)),
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
					ConflictsWith: []string{usernamePath, passwordPath},
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
	return &schema.Schema{
		Description: "Kibana connection configuration block.",
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
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
					Description: "A list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Type:        schema.TypeList,
					Optional:    true,
					Sensitive:   true,
					MaxItems:    1, // Current API restriction
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
					Description: "API key to use for API authentication to Fleet.",
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
