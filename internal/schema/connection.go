package schema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
	deprecationMessage := "This property will be removed in a future provider version. Configure the Elasticsearch connection via the provider configuration instead."

	if isProviderConfiguration {
		withEnvDefault = func(key string, dv interface{}) schema.SchemaDefaultFunc { return schema.EnvDefaultFunc(key, dv) }
		deprecationMessage = ""

		// RequireWith validation isn't compatible when used in conjunction with DefaultFunc
		usernameRequiredWithValidation = nil
		passwordRequiredWithValidation = nil
	}

	return &schema.Schema{
		Description: fmt.Sprintf("Elasticsearch connection configuration block. %s", deprecationMessage),
		Deprecated:  deprecationMessage,
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
					ConflictsWith: []string{usernamePath, passwordPath},
				},
				"endpoints": {
					Description: "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
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
					DefaultFunc:  schema.EnvDefaultFunc("ELASTICSEARCH_USERNAME", nil),
					RequiredWith: []string{"kibana.0.password"},
				},
				"password": {
					Description:  "Password to use for API authentication to Kibana.",
					Type:         schema.TypeString,
					Optional:     true,
					Sensitive:    true,
					DefaultFunc:  schema.EnvDefaultFunc("ELASTICSEARCH_PASSWORD", nil),
					RequiredWith: []string{"kibana.0.username"},
				},
				"endpoints": {
					Description: "A comma-separated list of endpoints where the terraform provider will point to, this must include the http(s) schema and port number.",
					Type:        schema.TypeList,
					MaxItems:    1, // Current API restriction
					Required:    true,
					Sensitive:   true,
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
