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

package snapshot_repository

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"maps"
)

// ensure the interface is satisfied at compile time
var _ validator.Object = blockRequiredAttrValidator{}

const schemaVersion int64 = 1

func GetSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Version:             schemaVersion,
		MarkdownDescription: schemaMarkdownDescription,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal identifier of the resource",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the snapshot repository to register or update.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"verify": schema.BoolAttribute{
				MarkdownDescription: "If true, the request verifies the repository is functional on all master and data nodes in the cluster.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
		},
		Blocks: map[string]schema.Block{
			"fs":    fsBlock(),
			"url":   urlBlock(),
			"gcs":   gcsBlock(),
			"azure": azureBlock(),
			"s3":    s3Block(),
			"hdfs":  hdfsBlock(),
		},
	}
}

// commonBlockAttributes returns the common settings shared across most repository types.
func commonBlockAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"chunk_size": schema.StringAttribute{
			MarkdownDescription: "Maximum size of files in snapshots.",
			Optional:            true,
			Computed:            true,
		},
		"compress": schema.BoolAttribute{
			MarkdownDescription: "If true, metadata files, such as index mappings and settings, are compressed in snapshots.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
		},
		"max_snapshot_bytes_per_sec": schema.StringAttribute{
			MarkdownDescription: "Maximum snapshot creation rate per node.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("40mb"),
		},
		"max_restore_bytes_per_sec": schema.StringAttribute{
			MarkdownDescription: "Maximum snapshot restore rate per node.",
			Optional:            true,
			Computed:            true,
		},
		"readonly": schema.BoolAttribute{
			MarkdownDescription: "If true, the repository is read-only.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
	}
}

// commonStdBlockAttributes returns attributes for standard (non-URL) repositories.
func commonStdBlockAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"max_number_of_snapshots": schema.Int64Attribute{
			MarkdownDescription: "Maximum number of snapshots the repository can contain.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(500),
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
			},
		},
	}
}

func fsBlock() schema.Block {
	attrs := mergeAttributes(commonBlockAttributes(), commonStdBlockAttributes(), map[string]schema.Attribute{
		"location": schema.StringAttribute{
			MarkdownDescription: "Location of the shared filesystem used to store and retrieve snapshots.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
	})
	return schema.SingleNestedBlock{
		MarkdownDescription: "Shared filesystem repository. Repositories of this type use a shared filesystem to store snapshots. " +
			"This filesystem must be accessible to all master and data nodes in the cluster.",
		Attributes: attrs,
		Validators: []validator.Object{
			requireBlockAttrs("location"),
		},
	}
}

func urlBlock() schema.Block {
	attrs := mergeAttributes(commonBlockAttributes(), commonStdBlockAttributes(), map[string]schema.Attribute{
		"url": schema.StringAttribute{
			MarkdownDescription: "URL location of the root of the shared filesystem repository.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
			Validators: []validator.String{
				stringvalidator.RegexMatches(urlProtocolRegex, "Url following protocols supported: file, ftp, http, https, jar"),
			},
		},
		"http_max_retries": schema.Int64Attribute{
			MarkdownDescription: "Maximum number of retries for http and https URLs.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(5),
			Validators: []validator.Int64{
				int64validator.AtLeast(0),
			},
		},
		"http_socket_timeout": schema.StringAttribute{
			MarkdownDescription: "Maximum wait time for data transfers over a connection.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("50s"),
		},
	})
	return schema.SingleNestedBlock{
		MarkdownDescription: "URL repository. Provides read-only access to a shared filesystem repository.",
		Attributes:          attrs,
		Validators: []validator.Object{
			requireBlockAttrs("url"),
		},
	}
}

func gcsBlock() schema.Block {
	attrs := mergeAttributes(commonBlockAttributes(), map[string]schema.Attribute{
		"bucket": schema.StringAttribute{
			MarkdownDescription: "The name of the bucket to be used for snapshots.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"client": schema.StringAttribute{
			MarkdownDescription: "The name of the client to use to connect to Google Cloud Storage.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("default"),
		},
		"base_path": schema.StringAttribute{
			MarkdownDescription: "Specifies the path within the bucket to the repository data. Defaults to the root of the bucket.",
			Optional:            true,
			Computed:            true,
		},
	})
	return schema.SingleNestedBlock{
		MarkdownDescription: "Google Cloud Storage repository. Stores snapshots in a Google Cloud Storage bucket.",
		Attributes:          attrs,
		Validators: []validator.Object{
			requireBlockAttrs("bucket"),
		},
	}
}

func azureBlock() schema.Block {
	attrs := mergeAttributes(commonBlockAttributes(), map[string]schema.Attribute{
		"container": schema.StringAttribute{
			MarkdownDescription: "Container name. You must create the Azure container before creating the repository.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"client": schema.StringAttribute{
			MarkdownDescription: "Azure named client to use.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("default"),
		},
		"base_path": schema.StringAttribute{
			MarkdownDescription: "Specifies the path within the container to the repository data.",
			Optional:            true,
			Computed:            true,
		},
		"location_mode": schema.StringAttribute{
			MarkdownDescription: "Location mode for the Azure repository. `primary_only` or `secondary_only`. " +
				"See the [Azure storage redundancy documentation](https://docs.microsoft.com/en-us/azure/storage/common/storage-redundancy) for more details.",
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString("primary_only"),
			Validators: []validator.String{
				stringvalidator.OneOf("primary_only", "secondary_only"),
			},
		},
	})
	return schema.SingleNestedBlock{
		MarkdownDescription: "Azure repository. Stores snapshots in Microsoft Azure Blob Storage.",
		Attributes:          attrs,
		Validators: []validator.Object{
			requireBlockAttrs("container"),
		},
	}
}

func s3Block() schema.Block {
	attrs := mergeAttributes(commonBlockAttributes(), map[string]schema.Attribute{
		"bucket": schema.StringAttribute{
			MarkdownDescription: "Name of the S3 bucket to use for snapshots.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"endpoint": schema.StringAttribute{
			MarkdownDescription: "Custom S3 service endpoint, useful when using VPC endpoints or non-default S3 URLs.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				s3EndpointValidator{},
			},
		},
		"client": schema.StringAttribute{
			MarkdownDescription: "The name of the S3 client to use to connect to S3.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("default"),
		},
		"base_path": schema.StringAttribute{
			MarkdownDescription: "Specifies the path to the repository data within its bucket.",
			Optional:            true,
			Computed:            true,
		},
		"server_side_encryption": schema.BoolAttribute{
			MarkdownDescription: "When true, files are encrypted server-side using AES-256 algorithm.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"buffer_size": schema.StringAttribute{
			MarkdownDescription: "Minimum threshold below which the chunk is uploaded using a single request.",
			Optional:            true,
			Computed:            true,
		},
		"canned_acl": schema.StringAttribute{
			MarkdownDescription: "The S3 repository supports all S3 canned ACLs.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("private"),
			Validators: []validator.String{
				stringvalidator.OneOf("private", "public-read", "public-read-write", "authenticated-read", "log-delivery-write", "bucket-owner-read", "bucket-owner-full-control"),
			},
		},
		"storage_class": schema.StringAttribute{
			MarkdownDescription: "Sets the S3 storage class for objects stored in the snapshot repository.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString("standard"),
			Validators: []validator.String{
				stringvalidator.OneOf("standard", "reduced_redundancy", "standard_ia", "onezone_ia", "intelligent_tiering"),
			},
		},
		"path_style_access": schema.BoolAttribute{
			MarkdownDescription: "If true, path style access pattern will be used.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
	})
	return schema.SingleNestedBlock{
		MarkdownDescription: "S3 repository. Stores snapshots in an Amazon S3 bucket.",
		Attributes:          attrs,
		Validators: []validator.Object{
			requireBlockAttrs("bucket"),
		},
	}
}

func hdfsBlock() schema.Block {
	attrs := mergeAttributes(commonBlockAttributes(), map[string]schema.Attribute{
		"uri": schema.StringAttribute{
			MarkdownDescription: `The uri address for hdfs. ex: "hdfs://<host>:<port>/".`,
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"path": schema.StringAttribute{
			MarkdownDescription: "The file path within the filesystem where data is stored/loaded.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"load_defaults": schema.BoolAttribute{
			MarkdownDescription: "Whether to load the default Hadoop configuration or not.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
		},
	})
	return schema.SingleNestedBlock{
		MarkdownDescription: "HDFS repository. Stores snapshots in Hadoop Distributed File System.",
		Attributes:          attrs,
		Validators: []validator.Object{
			requireBlockAttrs("uri", "path"),
		},
	}
}

// mergeAttributes merges multiple attribute maps into one.
func mergeAttributes(mps ...map[string]schema.Attribute) map[string]schema.Attribute {
	result := make(map[string]schema.Attribute)
	for _, m := range mps {
		maps.Copy(result, m)
	}
	return result
}
