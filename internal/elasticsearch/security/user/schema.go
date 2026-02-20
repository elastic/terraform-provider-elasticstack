package securityuser

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	providerschema "github.com/elastic/terraform-provider-elasticstack/internal/schema"
)

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = GetSchema()
}

func GetSchema() schema.Schema {
	const usernameAllowedCharsError = "must contain alphanumeric characters (a-z, A-Z, 0-9), spaces, punctuation, and printable symbols " +
		"in the Basic Latin (ASCII) block. Leading or trailing whitespace is not allowed"

	return schema.Schema{
		MarkdownDescription: userResourceDescription,
		Blocks: map[string]schema.Block{
			"elasticsearch_connection": providerschema.GetEsFWConnectionBlock(false),
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: usernameDescription,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1024),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[[:graph:]]+$`),
						usernameAllowedCharsError,
					),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The user's password. Passwords must be at least 6 characters long. Note: Consider using `password_wo` for better security with ephemeral resources.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 128),
					stringvalidator.ConflictsWith(path.MatchRoot("password_hash"), path.MatchRoot("password_wo")),
					stringvalidator.PreferWriteOnlyAttribute(path.MatchRoot("password_wo")),
				},
			},
			"password_hash": schema.StringAttribute{
				MarkdownDescription: passwordHashDescription,
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 128),
					stringvalidator.ConflictsWith(path.MatchRoot("password"), path.MatchRoot("password_wo")),
				},
			},
			"password_wo": schema.StringAttribute{
				MarkdownDescription: passwordWriteOnlyDescription,
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(6, 128),
					stringvalidator.ConflictsWith(path.MatchRoot("password"), path.MatchRoot("password_hash")),
				},
			},
			"password_wo_version": schema.StringAttribute{
				MarkdownDescription: passwordWriteOnlyVersionDescription,
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRoot("password_wo")),
				},
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "The full name of the user.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email of the user.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"roles": schema.SetAttribute{
				MarkdownDescription: "A set of roles the user has. The roles determine the user's access permissions.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"metadata": schema.StringAttribute{
				MarkdownDescription: "Arbitrary metadata that you want to associate with the user.",
				CustomType:          jsontypes.NormalizedType{},
				Optional:            true,
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the user is enabled. The default value is true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
	}
}
