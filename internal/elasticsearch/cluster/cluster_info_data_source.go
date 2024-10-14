package cluster

import (
	"context"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceClusterInfo() *schema.Resource {
	versionSchema := map[string]*schema.Schema{
		"build_date": {
			Description: "Build date.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
		"build_flavor": {
			Description: "Build Flavor.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
		"build_hash": {
			Description: "Short hash of the last git commit in this release.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
		"build_snapshot": {
			Description: "Build Snapshot.",
			Type:        schema.TypeBool,
			Computed:    true,
			Required:    false,
		},
		"build_type": {
			Description: "Build Type.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
		"lucene_version": {
			Description: "Lucene Version.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
		"minimum_index_compatibility_version": {
			Description: "Minimum index compatibility version.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
		"minimum_wire_compatibility_version": {
			Description: "Minimum wire compatibility version.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
		"number": {
			Description: "Elasticsearch version number.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
	}

	clusterInfoSchema := map[string]*schema.Schema{
		"version": {
			Description: "Contains statistics about the number of nodes selected by the request's node filters.",
			Type:        schema.TypeList,
			Computed:    true,
			Required:    false,
			Elem: &schema.Resource{
				Schema: versionSchema,
			},
		},
		"cluster_name": {
			Description: "Name of the cluster, based on the Cluster name setting setting.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
		"cluster_uuid": {
			Description: "Unique identifier for the cluster.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
		"name": {
			Description: "Name of the node.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
		"tagline": {
			Description: "Elasticsearh tag line.",
			Type:        schema.TypeString,
			Computed:    true,
			Required:    false,
		},
	}

	return &schema.Resource{
		Description: "Gets information about the Elastic cluster.",

		ReadContext: dataSourceClusterInfoRead,

		Schema: clusterInfoSchema,
	}
}

func dataSourceClusterInfoRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diags := clients.NewApiClientFromSDKResource(d, meta)
	if diags.HasError() {
		return diags
	}
	info, diags := elasticsearch.GetClusterInfo(ctx, client)
	if diags.HasError() {
		return diags
	}
	d.SetId(info.ClusterUUID)

	if err := d.Set("cluster_uuid", info.ClusterUUID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("cluster_name", info.ClusterName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", info.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tagline", info.Tagline); err != nil {
		return diag.FromErr(err)
	}

	version := map[string]interface{}{}
	version["build_date"] = info.Version.BuildDate.String()
	version["build_flavor"] = info.Version.BuildFlavor
	version["build_hash"] = info.Version.BuildHash
	version["build_snapshot"] = info.Version.BuildSnapshot
	version["build_type"] = info.Version.BuildType
	version["lucene_version"] = info.Version.LuceneVersion
	version["minimum_index_compatibility_version"] = info.Version.MinimumIndexCompatibilityVersion
	version["minimum_wire_compatibility_version"] = info.Version.MinimumWireCompatibilityVersion
	version["number"] = info.Version.Number
	if err := d.Set("version", []interface{}{version}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
