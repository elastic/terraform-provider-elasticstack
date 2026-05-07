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
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"maps"
)

func GetSchema(_ context.Context) schema.Schema {
	return schema.Schema{
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
			"fs":    fsAttribute(),
			"url":   urlAttribute(),
			"gcs":   gcsAttribute(),
			"azure": azureAttribute(),
			"s3":    s3Attribute(),
			"hdfs":  hdfsAttribute(),
		},
	}
}

// commonAttributes returns the common settings shared across most repository types.
func commonAttributes() map[string]schema.Attribute {
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

// commonStdAttributes returns attributes for standard (non-URL) repositories.
func commonStdAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"max_number_of_snapshots": schema.Int64Attribute{
			MarkdownDescription: "Maximum number of snapshots the repository can contain.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(500),
		},
	}
}

func fsAttribute() schema.Attribute {
	attrs := mergeAttributes(commonAttributes(), commonStdAttributes(), map[string]schema.Attribute{
		"location": schema.StringAttribute{
			MarkdownDescription: "Location of the shared filesystem used to store and retrieve snapshots.",
			Required:            true,
		},
	})
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Shared filesystem repository. Repositories of this type use a shared filesystem to store snapshots. " +
			"This filesystem must be accessible to all master and data nodes in the cluster.",
		Optional:   true,
		Attributes: attrs,
	}
}

func urlAttribute() schema.Attribute {
	attrs := mergeAttributes(commonAttributes(), commonStdAttributes(), map[string]schema.Attribute{
		"url": schema.StringAttribute{
			MarkdownDescription: "URL location of the root of the shared filesystem repository.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.RegexMatches(urlProtocolRegex, "Url following protocols supported: file, ftp, http, https, jar"),
			},
		},
		"http_max_retries": schema.Int64Attribute{
			MarkdownDescription: "Maximum number of retries for http and https URLs.",
			Optional:            true,
			Computed:            true,
			Default:             int64default.StaticInt64(5),
		},
		"http_socket_timeout": schema.StringAttribute{
			MarkdownDescription: "Maximum wait time for data transfers over a connection.",
			Optional:            true,
			Computed:            true,
		},
	})
	return schema.SingleNestedAttribute{
		MarkdownDescription: "URL repository. Provides read-only access to a shared filesystem repository.",
		Optional:            true,
		Attributes:          attrs,
	}
}

func gcsAttribute() schema.Attribute {
	attrs := mergeAttributes(commonAttributes(), map[string]schema.Attribute{
		"bucket": schema.StringAttribute{
			MarkdownDescription: "The name of the bucket to be used for snapshots.",
			Required:            true,
		},
		"client": schema.StringAttribute{
			MarkdownDescription: "The name of the client to use to connect to Google Cloud Storage.",
			Optional:            true,
			Computed:            true,
		},
		"base_path": schema.StringAttribute{
			MarkdownDescription: "Specifies the path within the bucket to the repository data. Defaults to the root of the bucket.",
			Optional:            true,
			Computed:            true,
		},
	})
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Google Cloud Storage repository. Stores snapshots in a Google Cloud Storage bucket.",
		Optional:            true,
		Attributes:          attrs,
	}
}

func azureAttribute() schema.Attribute {
	attrs := mergeAttributes(commonAttributes(), map[string]schema.Attribute{
		"container": schema.StringAttribute{
			MarkdownDescription: "Container name. You must create the Azure container before creating the repository.",
			Required:            true,
		},
		"client": schema.StringAttribute{
			MarkdownDescription: "Azure named client to use.",
			Optional:            true,
			Computed:            true,
		},
		"base_path": schema.StringAttribute{
			MarkdownDescription: "Specifies the path within the container to the repository data.",
			Optional:            true,
			Computed:            true,
		},
		"location_mode": schema.StringAttribute{
			MarkdownDescription: "Location mode for the Azure repository. Primary_only or secondary_only.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.OneOf("primary_only", "secondary_only"),
			},
		},
	})
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Azure repository. Stores snapshots in Microsoft Azure Blob Storage.",
		Optional:            true,
		Attributes:          attrs,
	}
}

func s3Attribute() schema.Attribute {
	attrs := mergeAttributes(commonAttributes(), map[string]schema.Attribute{
		"bucket": schema.StringAttribute{
			MarkdownDescription: "Name of the S3 bucket to use for snapshots.",
			Required:            true,
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
			Validators: []validator.String{
				stringvalidator.OneOf("private", "public-read", "public-read-write", "authenticated-read", "log-delivery-write", "bucket-owner-read", "bucket-owner-full-control"),
			},
		},
		"storage_class": schema.StringAttribute{
			MarkdownDescription: "Sets the S3 storage class for objects stored in the snapshot repository.",
			Optional:            true,
			Computed:            true,
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
	return schema.SingleNestedAttribute{
		MarkdownDescription: "S3 repository. Stores snapshots in an Amazon S3 bucket.",
		Optional:            true,
		Attributes:          attrs,
	}
}

func hdfsAttribute() schema.Attribute {
	attrs := mergeAttributes(commonAttributes(), map[string]schema.Attribute{
		"uri": schema.StringAttribute{
			MarkdownDescription: `The uri address for hdfs. ex: "hdfs://<host>:<port>/".`,
			Required:            true,
		},
		"path": schema.StringAttribute{
			MarkdownDescription: "The file path within the filesystem where data is stored/loaded.",
			Required:            true,
		},
		"load_defaults": schema.BoolAttribute{
			MarkdownDescription: "Whether to load the default Hadoop configuration or not.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
		},
	})
	return schema.SingleNestedAttribute{
		MarkdownDescription: "HDFS repository. Stores snapshots in Hadoop Distributed File System.",
		Optional:            true,
		Attributes:          attrs,
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
