package cluster

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/models"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceSnapshotRepository() *schema.Resource {
	commonStdSettings := map[string]*schema.Schema{
		"max_number_of_snapshots": {
			Description:  "Maximum number of snapshots the repository can contain.",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      500,
			ValidateFunc: validation.IntAtLeast(1),
		},
	}

	commonSettings := map[string]*schema.Schema{
		"chunk_size": {
			Description: "Maximum size of files in snapshots.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"compress": {
			Description: "If true, metadata files, such as index mappings and settings, are compressed in snapshots.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"max_snapshot_bytes_per_sec": {
			Description: "Maximum snapshot creation rate per node.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "40mb",
		},
		"max_restore_bytes_per_sec": {
			Description: "Maximum snapshot restore rate per node.",
			Type:        schema.TypeString,
			Optional:    true,
		},
		"readonly": {
			Description: "If true, the repository is read-only.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
	}

	// -- repos specific settings

	fsSettings := map[string]*schema.Schema{
		"location": {
			Description: "Location of the shared filesystem used to store and retrieve snapshots.",
			Type:        schema.TypeString,
			Required:    true,
		},
	}

	urlSettings := map[string]*schema.Schema{
		"url": {
			Description:  "URL location of the root of the shared filesystem repository.",
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringMatch(regexp.MustCompile("^(file:|ftp:|http:|https:|jar:)"), "Url following protocols supported: file, ftp, http, https, jar"),
		},
		"http_max_retries": {
			Description:  "Maximum number of retries for http and https URLs.",
			Type:         schema.TypeInt,
			Optional:     true,
			Default:      5,
			ValidateFunc: validation.IntAtLeast(0),
		},
		"http_socket_timeout": {
			Description: "Maximum wait time for data transfers over a connection.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "50s",
		},
	}

	gcsSettings := map[string]*schema.Schema{
		"bucket": {
			Description: "The name of the bucket to be used for snapshots.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"client": {
			Description: "The name of the client to use to connect to Google Cloud Storage.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "default",
		},
		"base_path": {
			Description: "Specifies the path within the bucket to the repository data. Defaults to the root of the bucket.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
	}

	azureSettings := map[string]*schema.Schema{
		"container": {
			Description: "Container name. You must create the Azure container before creating the repository.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"client": {
			Description: "Azure named client to use.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "default",
		},
		"base_path": {
			Description: "Specifies the path within the container to the repository data.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"location_mode": {
			Description:  "Location mode. `primary_only` or `secondary_only`. See: https://docs.microsoft.com/en-us/azure/storage/common/storage-redundancy",
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "primary_only",
			ValidateFunc: validation.StringInSlice([]string{"primary_only", "secondary_only"}, false),
		},
	}

	s3Settings := map[string]*schema.Schema{
		"bucket": {
			Description: "Name of the S3 bucket to use for snapshots.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"client": {
			Description: "The name of the S3 client to use to connect to S3.",
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "default",
		},
		"base_path": {
			Description: "Specifies the path to the repository data within its bucket.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"server_side_encryption": {
			Description: "When true, files are encrypted server-side using AES-256 algorithm.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
		},
		"buffer_size": {
			Description: "Minimum threshold below which the chunk is uploaded using a single request.",
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
		},
		"canned_acl": {
			Description:  "The S3 repository supports all S3 canned ACLs.",
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "private",
			ValidateFunc: validation.StringInSlice([]string{"private", "public-read", "public-read-write", "authenticated-read", "log-delivery-write", "bucket-owner-read", "bucket-owner-full-control"}, false),
		},
		"storage_class": {
			Description:  "Sets the S3 storage class for objects stored in the snapshot repository.",
			Type:         schema.TypeString,
			Optional:     true,
			Default:      "standard",
			ValidateFunc: validation.StringInSlice([]string{"standard", "reduced_redundancy", "standard_ia", "onezone_ia", "intelligent_tiering"}, false),
		},
	}

	hdfsSettings := map[string]*schema.Schema{
		"uri": {
			Description: `The uri address for hdfs. ex: "hdfs://<host>:<port>/".`,
			Type:        schema.TypeString,
			Required:    true,
		},
		"path": {
			Description: "The file path within the filesystem where data is stored/loaded.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"load_defaults": {
			Description: "Whether to load the default Hadoop configuration or not.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
	}

	// --

	snapRepoSchema := map[string]*schema.Schema{
		"id": {
			Description: "Internal identifier of the resource",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"name": {
			Description: "Name of the snapshot repository to register or update.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"verify": {
			Description: "If true, the request verifies the repository is functional on all master and data nodes in the cluster.",
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
		},
		"fs": {
			Description:   "Shared filesystem repository. Repositories of this type use a shared filesystem to store snapshots. This filesystem must be accessible to all master and data nodes in the cluster.",
			Type:          schema.TypeList,
			ForceNew:      true,
			Optional:      true,
			MaxItems:      1,
			ConflictsWith: []string{"url", "gcs", "azure", "s3", "hdfs"},
			ExactlyOneOf:  []string{"fs", "url", "gcs", "azure", "s3", "hdfs"},
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, commonStdSettings, fsSettings),
			},
		},
		"url": {
			Description:   "URL repository. Repositories of this type are read-only for the cluster. This means the cluster can retrieve or restore snapshots from the repository but cannot write or create snapshots in it.",
			Type:          schema.TypeList,
			ForceNew:      true,
			Optional:      true,
			MaxItems:      1,
			ConflictsWith: []string{"fs", "gcs", "azure", "s3", "hdfs"},
			ExactlyOneOf:  []string{"fs", "url", "gcs", "azure", "s3", "hdfs"},
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, commonStdSettings, urlSettings),
			},
		},
		"gcs": {
			Description:   "Support for using the Google Cloud Storage service as a repository for Snapshot/Restore. See: https://www.elastic.co/guide/en/elasticsearch/plugins/current/repository-gcs.html",
			Type:          schema.TypeList,
			ForceNew:      true,
			Optional:      true,
			MaxItems:      1,
			ConflictsWith: []string{"fs", "s3", "azure", "hdfs", "url"},
			ExactlyOneOf:  []string{"fs", "url", "gcs", "azure", "s3", "hdfs"},
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, gcsSettings),
			},
		},
		"azure": {
			Description:   "Support for using Azure Blob storage as a repository for Snapshot/Restore. See: https://www.elastic.co/guide/en/elasticsearch/plugins/current/repository-azure.html",
			Type:          schema.TypeList,
			ForceNew:      true,
			Optional:      true,
			MaxItems:      1,
			ConflictsWith: []string{"fs", "gcs", "url", "s3", "hdfs"},
			ExactlyOneOf:  []string{"fs", "url", "gcs", "azure", "s3", "hdfs"},
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, azureSettings),
			},
		},
		"s3": {
			Description:   "Support for using AWS S3 as a repository for Snapshot/Restore. See: https://www.elastic.co/guide/en/elasticsearch/plugins/current/repository-s3-repository.html",
			Type:          schema.TypeList,
			ForceNew:      true,
			Optional:      true,
			MaxItems:      1,
			ConflictsWith: []string{"fs", "url", "gcs", "azure", "hdfs"},
			ExactlyOneOf:  []string{"fs", "url", "gcs", "azure", "s3", "hdfs"},
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, s3Settings),
			},
		},
		"hdfs": {
			Description:   "Support for using HDFS File System as a repository for Snapshot/Restore. See: https://www.elastic.co/guide/en/elasticsearch/plugins/current/repository-hdfs.html",
			Type:          schema.TypeList,
			ForceNew:      true,
			Optional:      true,
			MaxItems:      1,
			ConflictsWith: []string{"fs", "url", "gcs", "azure", "s3"},
			ExactlyOneOf:  []string{"fs", "url", "gcs", "azure", "s3", "hdfs"},
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, hdfsSettings),
			},
		},
	}

	utils.AddConnectionSchema(snapRepoSchema)

	return &schema.Resource{
		Description: "Registers or updates a snapshot repository. See: https://www.elastic.co/guide/en/elasticsearch/reference/current/put-snapshot-repo-api.html and https://www.elastic.co/guide/en/elasticsearch/reference/current/snapshots-register-repository.html",

		CreateContext: resourceSnapRepoPut,
		UpdateContext: resourceSnapRepoPut,
		ReadContext:   resourceSnapRepoRead,
		DeleteContext: resourceSnapRepoDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: snapRepoSchema,
	}
}

func resourceSnapRepoPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	repoId := d.Get("name").(string)
	id, diags := client.ID(ctx, repoId)
	if diags.HasError() {
		return diags
	}

	var snapRepo models.SnapshotRepository
	snapRepo.Name = repoId
	snapRepoSettings := make(map[string]interface{})

	if v, ok := d.GetOk("verify"); ok {
		snapRepo.Verify = v.(bool)
	}

	// find supported repository types and iterate over them
	schemaTypes := ResourceSnapshotRepository().Schema
	delete(schemaTypes, "elasticsearch_connection")
	for t := range schemaTypes {
		if v, ok := d.GetOk(t); ok && reflect.TypeOf(v).Kind() == reflect.Slice {
			snapRepo.Type = t
			expandFsSettings(v.([]interface{})[0].(map[string]interface{}), snapRepoSettings)
		}
	}
	snapRepo.Settings = snapRepoSettings

	if diags := elasticsearch.PutSnapshotRepository(ctx, client, &snapRepo); diags.HasError() {
		return diags
	}
	d.SetId(id.String())
	return resourceSnapRepoRead(ctx, d, meta)
}

func expandFsSettings(source, target map[string]interface{}) {
	for k, v := range source {
		if !utils.IsEmpty(v) {
			target[k] = v
		}
	}
}

func resourceSnapRepoRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}

	currentRepo, diags := elasticsearch.GetSnapshotRepository(ctx, client, compId.ResourceId)
	if currentRepo == nil && diags == nil {
		tflog.Warn(ctx, fmt.Sprintf(`Snapshot repository "%s" not found, removing from state`, compId.ResourceId))
		d.SetId("")
		return diags
	}
	if diags.HasError() {
		return diags
	}

	if _, ok := ResourceSnapshotRepository().Schema[currentRepo.Type]; !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API responded with unsupported type of the snapshot repository.",
			Detail:   "The type of the snapshot repository is not supported.",
		})
		return diags
	}

	// get the schema of the Elem of the current repo type
	schemaSettings := ResourceSnapshotRepository().Schema[currentRepo.Type].Elem.(*schema.Resource).Schema

	settings, err := flattenRepoSettings(currentRepo, schemaSettings)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to parse snapshot repository settings.",
			Detail:   fmt.Sprintf(`Unable to parse settings returned by ES API: %v`, err),
		})
		return diags
	}
	if err := d.Set(currentRepo.Type, settings); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("name", currentRepo.Name); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func flattenRepoSettings(r *models.SnapshotRepository, s map[string]*schema.Schema) ([]interface{}, error) {
	settings := make(map[string]interface{})
	result := make([]interface{}, 1)

	// make sure the schema contains the fetched setting
	for k, v := range r.Settings {
		if schemaDef, ok := s[k]; ok && !utils.IsEmpty(v) {
			switch schemaDef.Type {
			case schema.TypeInt, schema.TypeFloat:
				i, err := strconv.Atoi(v.(string))
				if err != nil {
					return nil, fmt.Errorf(`Failed to parse value = "%v" for setting = "%s"`, v, k)
				}
				settings[k] = i
			case schema.TypeBool:
				b, err := strconv.ParseBool(v.(string))
				if err != nil {
					return nil, fmt.Errorf(`Failed to parse value = "%v" for setting = "%s"`, v, k)
				}
				settings[k] = b
			default:
				settings[k] = v
			}
		}
	}
	result[0] = settings
	return result, nil
}

func resourceSnapRepoDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}

	id := d.Id()
	compId, diags := clients.CompositeIdFromStr(id)
	if diags.HasError() {
		return diags
	}

	if diags := elasticsearch.DeleteSnapshotRepository(ctx, client, compId.ResourceId); diags.HasError() {
		return diags
	}
	return diags
}
