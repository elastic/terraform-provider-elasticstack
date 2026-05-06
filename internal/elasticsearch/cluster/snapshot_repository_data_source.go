// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cluster

import (
	"context"
	"fmt"
	"maps"
	"strconv"

	"github.com/elastic/terraform-provider-elasticstack/internal/clients"
	"github.com/elastic/terraform-provider-elasticstack/internal/clients/elasticsearch"
	"github.com/elastic/terraform-provider-elasticstack/internal/diagutil"
	"github.com/elastic/terraform-provider-elasticstack/internal/entitycore"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// -- Models

type snapshotRepositoryDataSourceModel struct {
	entitycore.ElasticsearchConnectionField
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Type  types.String `tfsdk:"type"`
	Fs    types.List   `tfsdk:"fs"`
	URL   types.List   `tfsdk:"url"`
	GCS   types.List   `tfsdk:"gcs"`
	Azure types.List   `tfsdk:"azure"`
	S3    types.List   `tfsdk:"s3"`
	HDFS  types.List   `tfsdk:"hdfs"`
}

type fsDataSourceModel struct {
	ChunkSize              types.String `tfsdk:"chunk_size"`
	Compress               types.Bool   `tfsdk:"compress"`
	MaxSnapshotBytesPerSec types.String `tfsdk:"max_snapshot_bytes_per_sec"`
	MaxRestoreBytesPerSec  types.String `tfsdk:"max_restore_bytes_per_sec"`
	Readonly               types.Bool   `tfsdk:"readonly"`
	MaxNumberOfSnapshots   types.Int64  `tfsdk:"max_number_of_snapshots"`
	Location               types.String `tfsdk:"location"`
}

type urlDataSourceModel struct {
	ChunkSize              types.String `tfsdk:"chunk_size"`
	Compress               types.Bool   `tfsdk:"compress"`
	MaxSnapshotBytesPerSec types.String `tfsdk:"max_snapshot_bytes_per_sec"`
	MaxRestoreBytesPerSec  types.String `tfsdk:"max_restore_bytes_per_sec"`
	Readonly               types.Bool   `tfsdk:"readonly"`
	MaxNumberOfSnapshots   types.Int64  `tfsdk:"max_number_of_snapshots"`
	URL                    types.String `tfsdk:"url"`
	HTTPMaxRetries         types.Int64  `tfsdk:"http_max_retries"`
	HTTPSocketTimeout      types.String `tfsdk:"http_socket_timeout"`
}

type gcsDataSourceModel struct {
	ChunkSize              types.String `tfsdk:"chunk_size"`
	Compress               types.Bool   `tfsdk:"compress"`
	MaxSnapshotBytesPerSec types.String `tfsdk:"max_snapshot_bytes_per_sec"`
	MaxRestoreBytesPerSec  types.String `tfsdk:"max_restore_bytes_per_sec"`
	Readonly               types.Bool   `tfsdk:"readonly"`
	Bucket                 types.String `tfsdk:"bucket"`
	Client                 types.String `tfsdk:"client"`
	BasePath               types.String `tfsdk:"base_path"`
}

type azureDataSourceModel struct {
	ChunkSize              types.String `tfsdk:"chunk_size"`
	Compress               types.Bool   `tfsdk:"compress"`
	MaxSnapshotBytesPerSec types.String `tfsdk:"max_snapshot_bytes_per_sec"`
	MaxRestoreBytesPerSec  types.String `tfsdk:"max_restore_bytes_per_sec"`
	Readonly               types.Bool   `tfsdk:"readonly"`
	Container              types.String `tfsdk:"container"`
	Client                 types.String `tfsdk:"client"`
	BasePath               types.String `tfsdk:"base_path"`
	LocationMode           types.String `tfsdk:"location_mode"`
}

type s3DataSourceModel struct {
	ChunkSize              types.String `tfsdk:"chunk_size"`
	Compress               types.Bool   `tfsdk:"compress"`
	MaxSnapshotBytesPerSec types.String `tfsdk:"max_snapshot_bytes_per_sec"`
	MaxRestoreBytesPerSec  types.String `tfsdk:"max_restore_bytes_per_sec"`
	Readonly               types.Bool   `tfsdk:"readonly"`
	Bucket                 types.String `tfsdk:"bucket"`
	Client                 types.String `tfsdk:"client"`
	BasePath               types.String `tfsdk:"base_path"`
	ServerSideEncryption   types.Bool   `tfsdk:"server_side_encryption"`
	BufferSize             types.String `tfsdk:"buffer_size"`
	CannedACL              types.String `tfsdk:"canned_acl"`
	StorageClass           types.String `tfsdk:"storage_class"`
	PathStyleAccess        types.Bool   `tfsdk:"path_style_access"`
}

type hdfsDataSourceModel struct {
	ChunkSize              types.String `tfsdk:"chunk_size"`
	Compress               types.Bool   `tfsdk:"compress"`
	MaxSnapshotBytesPerSec types.String `tfsdk:"max_snapshot_bytes_per_sec"`
	MaxRestoreBytesPerSec  types.String `tfsdk:"max_restore_bytes_per_sec"`
	Readonly               types.Bool   `tfsdk:"readonly"`
	URI                    types.String `tfsdk:"uri"`
	Path                   types.String `tfsdk:"path"`
	LoadDefaults           types.Bool   `tfsdk:"load_defaults"`
}

// -- Schema

func getDataSourceSchema() schema.Schema {
	commonSettings := map[string]schema.Attribute{
		"chunk_size": schema.StringAttribute{
			MarkdownDescription: "Maximum size of files in snapshots.",
			Computed:            true,
		},
		"compress": schema.BoolAttribute{
			MarkdownDescription: "If true, metadata files, such as index mappings and settings, are compressed in snapshots.",
			Computed:            true,
		},
		"max_snapshot_bytes_per_sec": schema.StringAttribute{
			MarkdownDescription: "Maximum snapshot creation rate per node.",
			Computed:            true,
		},
		"max_restore_bytes_per_sec": schema.StringAttribute{
			MarkdownDescription: "Maximum snapshot restore rate per node.",
			Computed:            true,
		},
		"readonly": schema.BoolAttribute{
			MarkdownDescription: "If true, the repository is read-only.",
			Computed:            true,
		},
	}

	commonStdSettings := map[string]schema.Attribute{
		"max_number_of_snapshots": schema.Int64Attribute{
			MarkdownDescription: "Maximum number of snapshots the repository can contain.",
			Computed:            true,
		},
	}

	fsSettings := map[string]schema.Attribute{
		"location": schema.StringAttribute{
			MarkdownDescription: "Location of the shared filesystem used to store and retrieve snapshots.",
			Computed:            true,
		},
	}

	urlSettings := map[string]schema.Attribute{
		"url": schema.StringAttribute{
			MarkdownDescription: "URL location of the root of the shared filesystem repository.",
			Computed:            true,
		},
		"http_max_retries": schema.Int64Attribute{
			MarkdownDescription: "Maximum number of retries for http and https URLs.",
			Computed:            true,
		},
		"http_socket_timeout": schema.StringAttribute{
			MarkdownDescription: "Maximum wait time for data transfers over a connection.",
			Computed:            true,
		},
	}

	gcsSettings := map[string]schema.Attribute{
		"bucket": schema.StringAttribute{
			MarkdownDescription: "The name of the bucket to be used for snapshots.",
			Computed:            true,
		},
		"client": schema.StringAttribute{
			MarkdownDescription: "The name of the client to use to connect to Google Cloud Storage.",
			Computed:            true,
		},
		"base_path": schema.StringAttribute{
			MarkdownDescription: "Specifies the path within the bucket to the repository data. Defaults to the root of the bucket.",
			Computed:            true,
		},
	}

	azureSettings := map[string]schema.Attribute{
		"container": schema.StringAttribute{
			MarkdownDescription: "Container name. You must create the Azure container before creating the repository.",
			Computed:            true,
		},
		"client": schema.StringAttribute{
			MarkdownDescription: "Azure named client to use.",
			Computed:            true,
		},
		"base_path": schema.StringAttribute{
			MarkdownDescription: "Specifies the path within the container to the repository data.",
			Computed:            true,
		},
		"location_mode": schema.StringAttribute{
			MarkdownDescription: snapshotRepositoryLocationModeDescription,
			Computed:            true,
		},
	}

	s3Settings := map[string]schema.Attribute{
		"bucket": schema.StringAttribute{
			MarkdownDescription: "Name of the S3 bucket to use for snapshots.",
			Computed:            true,
		},
		"client": schema.StringAttribute{
			MarkdownDescription: "The name of the S3 client to use to connect to S3.",
			Computed:            true,
		},
		"base_path": schema.StringAttribute{
			MarkdownDescription: "Specifies the path to the repository data within its bucket.",
			Computed:            true,
		},
		"server_side_encryption": schema.BoolAttribute{
			MarkdownDescription: "When true, files are encrypted server-side using AES-256 algorithm.",
			Computed:            true,
		},
		"buffer_size": schema.StringAttribute{
			MarkdownDescription: "Minimum threshold below which the chunk is uploaded using a single request.",
			Computed:            true,
		},
		"canned_acl": schema.StringAttribute{
			MarkdownDescription: "The S3 repository supports all S3 canned ACLs.",
			Computed:            true,
		},
		"storage_class": schema.StringAttribute{
			MarkdownDescription: "Sets the S3 storage class for objects stored in the snapshot repository.",
			Computed:            true,
		},
		"path_style_access": schema.BoolAttribute{
			MarkdownDescription: "If true, path style access pattern will be used.",
			Computed:            true,
		},
	}

	hdfsSettings := map[string]schema.Attribute{
		"uri": schema.StringAttribute{
			MarkdownDescription: `The uri address for hdfs. ex: "hdfs://<host>:<port>/".",`,
			Computed:            true,
		},
		"path": schema.StringAttribute{
			MarkdownDescription: "The file path within the filesystem where data is stored/loaded.",
			Computed:            true,
		},
		"load_defaults": schema.BoolAttribute{
			MarkdownDescription: "Whether to load the default Hadoop configuration or not.",
			Computed:            true,
		},
	}

	return schema.Schema{
		MarkdownDescription: "Gets information about the registered snapshot repositories.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the snapshot repository.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Repository type.",
				Computed:            true,
			},
			"fs": schema.ListNestedAttribute{
				MarkdownDescription: "Shared filesystem repository. Set only if the type of the fetched repo is `fs`.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: mergeAttrMaps(commonSettings, commonStdSettings, fsSettings),
				},
			},
			"url": schema.ListNestedAttribute{
				MarkdownDescription: "URL repository. Set only if the type of the fetched repo is `url`.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: mergeAttrMaps(commonSettings, commonStdSettings, urlSettings),
				},
			},
			"gcs": schema.ListNestedAttribute{
				MarkdownDescription: "Google Cloud Storage service as a repository. Set only if the type of the fetched repo is `gcs`.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: mergeAttrMaps(commonSettings, gcsSettings),
				},
			},
			"azure": schema.ListNestedAttribute{
				MarkdownDescription: "Azure Blob storage as a repository. Set only if the type of the fetched repo is `azure`.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: mergeAttrMaps(commonSettings, azureSettings),
				},
			},
			"s3": schema.ListNestedAttribute{
				MarkdownDescription: "AWS S3 as a repository. Set only if the type of the fetched repo is `s3`.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: mergeAttrMaps(commonSettings, s3Settings),
				},
			},
			"hdfs": schema.ListNestedAttribute{
				MarkdownDescription: "HDFS File System as a repository. Set only if the type of the fetched repo is `hdfs`.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: mergeAttrMaps(commonSettings, hdfsSettings),
				},
			},
		},
	}
}

func mergeAttrMaps(mapsToMerge ...map[string]schema.Attribute) map[string]schema.Attribute {
	result := make(map[string]schema.Attribute)
	for _, m := range mapsToMerge {
		maps.Copy(result, m)
	}
	return result
}

// -- Element types

func fsElementType() attr.Type {
	return getDataSourceSchema().Attributes["fs"].GetType().(attr.TypeWithElementType).ElementType()
}

func urlElementType() attr.Type {
	return getDataSourceSchema().Attributes["url"].GetType().(attr.TypeWithElementType).ElementType()
}

func gcsElementType() attr.Type {
	return getDataSourceSchema().Attributes["gcs"].GetType().(attr.TypeWithElementType).ElementType()
}

func azureElementType() attr.Type {
	return getDataSourceSchema().Attributes["azure"].GetType().(attr.TypeWithElementType).ElementType()
}

func s3ElementType() attr.Type {
	return getDataSourceSchema().Attributes["s3"].GetType().(attr.TypeWithElementType).ElementType()
}

func hdfsElementType() attr.Type {
	return getDataSourceSchema().Attributes["hdfs"].GetType().(attr.TypeWithElementType).ElementType()
}

// -- Constructor

func NewSnapshotRepositoryDataSource() datasource.DataSource {
	return entitycore.NewElasticsearchDataSource[snapshotRepositoryDataSourceModel](
		entitycore.ComponentElasticsearch,
		"snapshot_repository",
		getDataSourceSchema,
		readDataSource,
	)
}

// -- Read callback

func readDataSource(ctx context.Context, esClient *clients.ElasticsearchScopedClient, config snapshotRepositoryDataSourceModel) (snapshotRepositoryDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	repoName := config.Name.ValueString()

	id, sdkDiags := esClient.ID(ctx, repoName)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return config, diags
	}
	config.ID = types.StringValue(id.String())

	currentRepo, sdkDiags := elasticsearch.GetSnapshotRepository(ctx, esClient, repoName)
	diags.Append(diagutil.FrameworkDiagsFromSDK(sdkDiags)...)
	if diags.HasError() {
		return config, diags
	}

	if currentRepo == nil {
		diags.AddWarning(
			fmt.Sprintf("Could not find snapshot repository [%s]", repoName),
			"",
		)
		return config, diags
	}

	config.Type = types.StringValue(currentRepo.Type)

	switch currentRepo.Type {
	case "fs":
		model := flattenFsSettings(currentRepo.Settings)
		listValue, listDiags := types.ListValueFrom(ctx, fsElementType(), []fsDataSourceModel{model})
		diags.Append(listDiags...)
		if diags.HasError() {
			return config, diags
		}
		config.Fs = listValue
	case "url":
		model := flattenURLSettings(currentRepo.Settings)
		listValue, listDiags := types.ListValueFrom(ctx, urlElementType(), []urlDataSourceModel{model})
		diags.Append(listDiags...)
		if diags.HasError() {
			return config, diags
		}
		config.URL = listValue
	case "gcs":
		model := flattenGCSSettings(currentRepo.Settings)
		listValue, listDiags := types.ListValueFrom(ctx, gcsElementType(), []gcsDataSourceModel{model})
		diags.Append(listDiags...)
		if diags.HasError() {
			return config, diags
		}
		config.GCS = listValue
	case "azure":
		model := flattenAzureSettings(currentRepo.Settings)
		listValue, listDiags := types.ListValueFrom(ctx, azureElementType(), []azureDataSourceModel{model})
		diags.Append(listDiags...)
		if diags.HasError() {
			return config, diags
		}
		config.Azure = listValue
	case "s3":
		model := flattenS3Settings(currentRepo.Settings)
		listValue, listDiags := types.ListValueFrom(ctx, s3ElementType(), []s3DataSourceModel{model})
		diags.Append(listDiags...)
		if diags.HasError() {
			return config, diags
		}
		config.S3 = listValue
	case "hdfs":
		model := flattenHDFSSettings(currentRepo.Settings)
		listValue, listDiags := types.ListValueFrom(ctx, hdfsElementType(), []hdfsDataSourceModel{model})
		diags.Append(listDiags...)
		if diags.HasError() {
			return config, diags
		}
		config.HDFS = listValue
	default:
		diags.AddError(
			"API responded with unsupported type of the snapshot repository.",
			fmt.Sprintf("The type '%s' of the snapshot repository is not supported.", currentRepo.Type),
		)
		return config, diags
	}

	return config, diags
}

// -- Flatten helpers

func flattenFsSettings(settings map[string]any) fsDataSourceModel {
	return fsDataSourceModel{
		ChunkSize:              stringSetting(settings, "chunk_size"),
		Compress:               boolSetting(settings, "compress"),
		MaxSnapshotBytesPerSec: stringSetting(settings, "max_snapshot_bytes_per_sec"),
		MaxRestoreBytesPerSec:  stringSetting(settings, "max_restore_bytes_per_sec"),
		Readonly:               boolSetting(settings, "readonly"),
		MaxNumberOfSnapshots:   int64Setting(settings, "max_number_of_snapshots"),
		Location:               stringSetting(settings, "location"),
	}
}

func flattenURLSettings(settings map[string]any) urlDataSourceModel {
	return urlDataSourceModel{
		ChunkSize:              stringSetting(settings, "chunk_size"),
		Compress:               boolSetting(settings, "compress"),
		MaxSnapshotBytesPerSec: stringSetting(settings, "max_snapshot_bytes_per_sec"),
		MaxRestoreBytesPerSec:  stringSetting(settings, "max_restore_bytes_per_sec"),
		Readonly:               boolSetting(settings, "readonly"),
		MaxNumberOfSnapshots:   int64Setting(settings, "max_number_of_snapshots"),
		URL:                    stringSetting(settings, "url"),
		HTTPMaxRetries:         int64Setting(settings, "http_max_retries"),
		HTTPSocketTimeout:      stringSetting(settings, "http_socket_timeout"),
	}
}

func flattenGCSSettings(settings map[string]any) gcsDataSourceModel {
	return gcsDataSourceModel{
		ChunkSize:              stringSetting(settings, "chunk_size"),
		Compress:               boolSetting(settings, "compress"),
		MaxSnapshotBytesPerSec: stringSetting(settings, "max_snapshot_bytes_per_sec"),
		MaxRestoreBytesPerSec:  stringSetting(settings, "max_restore_bytes_per_sec"),
		Readonly:               boolSetting(settings, "readonly"),
		Bucket:                 stringSetting(settings, "bucket"),
		Client:                 stringSetting(settings, "client"),
		BasePath:               stringSetting(settings, "base_path"),
	}
}

func flattenAzureSettings(settings map[string]any) azureDataSourceModel {
	return azureDataSourceModel{
		ChunkSize:              stringSetting(settings, "chunk_size"),
		Compress:               boolSetting(settings, "compress"),
		MaxSnapshotBytesPerSec: stringSetting(settings, "max_snapshot_bytes_per_sec"),
		MaxRestoreBytesPerSec:  stringSetting(settings, "max_restore_bytes_per_sec"),
		Readonly:               boolSetting(settings, "readonly"),
		Container:              stringSetting(settings, "container"),
		Client:                 stringSetting(settings, "client"),
		BasePath:               stringSetting(settings, "base_path"),
		LocationMode:           stringSetting(settings, "location_mode"),
	}
}

func flattenS3Settings(settings map[string]any) s3DataSourceModel {
	return s3DataSourceModel{
		ChunkSize:              stringSetting(settings, "chunk_size"),
		Compress:               boolSetting(settings, "compress"),
		MaxSnapshotBytesPerSec: stringSetting(settings, "max_snapshot_bytes_per_sec"),
		MaxRestoreBytesPerSec:  stringSetting(settings, "max_restore_bytes_per_sec"),
		Readonly:               boolSetting(settings, "readonly"),
		Bucket:                 stringSetting(settings, "bucket"),
		Client:                 stringSetting(settings, "client"),
		BasePath:               stringSetting(settings, "base_path"),
		ServerSideEncryption:   boolSetting(settings, "server_side_encryption"),
		BufferSize:             stringSetting(settings, "buffer_size"),
		CannedACL:              stringSetting(settings, "canned_acl"),
		StorageClass:           stringSetting(settings, "storage_class"),
		PathStyleAccess:        boolSetting(settings, "path_style_access"),
	}
}

func flattenHDFSSettings(settings map[string]any) hdfsDataSourceModel {
	return hdfsDataSourceModel{
		ChunkSize:              stringSetting(settings, "chunk_size"),
		Compress:               boolSetting(settings, "compress"),
		MaxSnapshotBytesPerSec: stringSetting(settings, "max_snapshot_bytes_per_sec"),
		MaxRestoreBytesPerSec:  stringSetting(settings, "max_restore_bytes_per_sec"),
		Readonly:               boolSetting(settings, "readonly"),
		URI:                    stringSetting(settings, "uri"),
		Path:                   stringSetting(settings, "path"),
		LoadDefaults:           boolSetting(settings, "load_defaults"),
	}
}

func stringSetting(settings map[string]any, key string) types.String {
	v, ok := settings[key]
	if !ok || v == nil {
		return types.StringNull()
	}
	switch val := v.(type) {
	case string:
		return types.StringValue(val)
	default:
		return types.StringValue(fmt.Sprintf("%v", val))
	}
}

func boolSetting(settings map[string]any, key string) types.Bool {
	v, ok := settings[key]
	if !ok || v == nil {
		return types.BoolNull()
	}
	switch val := v.(type) {
	case bool:
		return types.BoolValue(val)
	case string:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return types.BoolNull()
		}
		return types.BoolValue(b)
	default:
		return types.BoolNull()
	}
}

func int64Setting(settings map[string]any, key string) types.Int64 {
	v, ok := settings[key]
	if !ok || v == nil {
		return types.Int64Null()
	}
	switch val := v.(type) {
	case int:
		return types.Int64Value(int64(val))
	case int64:
		return types.Int64Value(val)
	case float64:
		return types.Int64Value(int64(val))
	case string:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return types.Int64Null()
		}
		return types.Int64Value(i)
	default:
		return types.Int64Null()
	}
}
