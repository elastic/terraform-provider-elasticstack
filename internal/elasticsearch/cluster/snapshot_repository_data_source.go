package cluster

import (
	"context"
	"fmt"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceSnapshotRespository() *schema.Resource {
	commonStdSettings := map[string]*schema.Schema{
		"max_number_of_snapshots": {
			Description: "Maximum number of snapshots the repository can contain.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
	}

	commonSettings := map[string]*schema.Schema{
		"chunk_size": {
			Description: "Maximum size of files in snapshots.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"compress": {
			Description: "If true, metadata files, such as index mappings and settings, are compressed in snapshots.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"max_snapshot_bytes_per_sec": {
			Description: "Maximum snapshot creation rate per node.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"max_restore_bytes_per_sec": {
			Description: "Maximum snapshot restore rate per node.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"readonly": {
			Description: "If true, the repository is read-only.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
	}

	// -- repos specific settings

	fsSettings := map[string]*schema.Schema{
		"location": {
			Description: "Location of the shared filesystem used to store and retrieve snapshots.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	urlSettings := map[string]*schema.Schema{
		"url": {
			Description: "URL location of the root of the shared filesystem repository.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"http_max_retries": {
			Description: "Maximum number of retries for http and https URLs.",
			Type:        schema.TypeInt,
			Computed:    true,
		},
		"http_socket_timeout": {
			Description: "Maximum wait time for data transfers over a connection.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	gcsSettings := map[string]*schema.Schema{
		"bucket": {
			Description: "The name of the bucket to be used for snapshots.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"client": {
			Description: "The name of the client to use to connect to Google Cloud Storage.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"base_path": {
			Description: "Specifies the path within the bucket to the repository data. Defaults to the root of the bucket.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	azureSettings := map[string]*schema.Schema{
		"container": {
			Description: "Container name. You must create the Azure container before creating the repository.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"client": {
			Description: "Azure named client to use.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"base_path": {
			Description: "Specifies the path within the container to the repository data.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"location_mode": {
			Description: "Location mode. `primary_only` or `secondary_only`. See: https://docs.microsoft.com/en-us/azure/storage/common/storage-redundancy",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	s3Settings := map[string]*schema.Schema{
		"bucket": {
			Description: "Name of the S3 bucket to use for snapshots.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"client": {
			Description: "The name of the S3 client to use to connect to S3.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"base_path": {
			Description: "Specifies the path to the repository data within its bucket.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"server_side_encryption": {
			Description: "When true, files are encrypted server-side using AES-256 algorithm.",
			Type:        schema.TypeBool,
			Computed:    true,
		},
		"buffer_size": {
			Description: "Minimum threshold below which the chunk is uploaded using a single request.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"canned_acl": {
			Description: "The S3 repository supports all S3 canned ACLs.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"storage_class": {
			Description: "Sets the S3 storage class for objects stored in the snapshot repository.",
			Type:        schema.TypeString,
			Computed:    true,
		},
	}

	hdfsSettings := map[string]*schema.Schema{
		"uri": {
			Description: `The uri address for hdfs. ex: "hdfs://<host>:<port>/".`,
			Type:        schema.TypeString,
			Computed:    true,
		},
		"path": {
			Description: "The file path within the filesystem where data is stored/loaded.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"load_defaults": {
			Description: "Whether to load the default Hadoop configuration or not.",
			Type:        schema.TypeBool,
			Computed:    true,
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
			Description: "Name of the snapshot repository.",
			Type:        schema.TypeString,
			Required:    true,
		},
		"type": {
			Description: "Repository type.",
			Type:        schema.TypeString,
			Computed:    true,
		},
		"fs": {
			Description: "Shared filesystem repository. Set only if the type of the fetched repo is `fs`.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, commonStdSettings, fsSettings),
			},
		},
		"url": {
			Description: "URL repository. Set only if the type of the fetched repo is `url`.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, commonStdSettings, urlSettings),
			},
		},
		"gcs": {
			Description: "Google Cloud Storage service as a repository. Set only if the type of the fetched repo is `gcs`.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, gcsSettings),
			},
		},
		"azure": {
			Description: "Azure Blob storage as a repository. Set only if the type of the fetched repo is `azure`.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, azureSettings),
			},
		},
		"s3": {
			Description: "AWS S3 as a repository. Set only if the type of the fetched repo is `s3`.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, s3Settings),
			},
		},
		"hdfs": {
			Description: "HDFS File System as a repository. Set only if the type of the fetched repo is `hdfs`.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: utils.MergeSchemaMaps(commonSettings, hdfsSettings),
			},
		},
	}

	utils.AddConnectionSchema(snapRepoSchema)

	return &schema.Resource{
		Description: "Gets information about the registered snapshot repositories.",

		ReadContext: dataSourceSnapRepoRead,

		Schema: snapRepoSchema,
	}
}

func dataSourceSnapRepoRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClient(d, meta)
	if diags.HasError() {
		return diags
	}
	repoName := d.Get("name").(string)
	id, diags := client.ID(ctx, repoName)
	if diags.HasError() {
		return diags
	}
	currentRepo, diags := elasticsearch.GetSnapshotRepository(ctx, client, repoName)
	if diags.HasError() {
		return diags
	}

	// get the schema of the Elem of the current repo type
	schemaSettings := DataSourceSnapshotRespository().Schema[currentRepo.Type].Elem.(*schema.Resource).Schema
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

	if err := d.Set("type", currentRepo.Type); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id.String())
	return diags
}
