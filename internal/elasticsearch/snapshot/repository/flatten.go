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

package repository

func flattenCommonSettings(settings map[string]any) (commonDataSourceModel, error) {
	var m commonDataSourceModel
	var err error
	m.ChunkSize = strSettingNull(settings, settingChunkSize)
	m.Compress, err = boolSettingNull(settings, settingCompress)
	if err != nil {
		return m, err
	}
	m.MaxSnapshotBytesPerSec = strSettingNull(settings, settingMaxSnapshotBytesPerSec)
	m.MaxRestoreBytesPerSec = strSettingNull(settings, settingMaxRestoreBytesPerSec)
	m.Readonly, err = boolSettingNull(settings, settingReadonly)
	if err != nil {
		return m, err
	}
	return m, nil
}

func flattenFsSettings(settings map[string]any) (fsDataSourceModel, error) {
	var m fsDataSourceModel
	var err error
	m.commonDataSourceModel, err = flattenCommonSettings(settings)
	if err != nil {
		return m, err
	}
	m.MaxNumberOfSnapshots, err = int64SettingNull(settings, settingMaxNumberOfSnapshots)
	if err != nil {
		return m, err
	}
	m.Location = strSettingNull(settings, settingLocation)
	return m, nil
}

func flattenURLSettings(settings map[string]any) (urlDataSourceModel, error) {
	var m urlDataSourceModel
	var err error
	m.commonDataSourceModel, err = flattenCommonSettings(settings)
	if err != nil {
		return m, err
	}
	m.MaxNumberOfSnapshots, err = int64SettingNull(settings, settingMaxNumberOfSnapshots)
	if err != nil {
		return m, err
	}
	m.URL = strSettingNull(settings, settingURL)
	m.HTTPMaxRetries, err = int64SettingNull(settings, "http_max_retries")
	if err != nil {
		return m, err
	}
	m.HTTPSocketTimeout = strSettingNull(settings, "http_socket_timeout")
	return m, nil
}

func flattenGCSSettings(settings map[string]any) (gcsDataSourceModel, error) {
	var m gcsDataSourceModel
	var err error
	m.commonDataSourceModel, err = flattenCommonSettings(settings)
	if err != nil {
		return m, err
	}
	m.Bucket = strSettingNull(settings, settingBucket)
	m.Client = strSettingNull(settings, settingClient)
	m.BasePath = strSettingNull(settings, settingBasePath)
	return m, nil
}

func flattenAzureSettings(settings map[string]any) (azureDataSourceModel, error) {
	var m azureDataSourceModel
	var err error
	m.commonDataSourceModel, err = flattenCommonSettings(settings)
	if err != nil {
		return m, err
	}
	m.Container = strSettingNull(settings, settingContainer)
	m.Client = strSettingNull(settings, settingClient)
	m.BasePath = strSettingNull(settings, settingBasePath)
	m.LocationMode = strSettingNull(settings, "location_mode")
	return m, nil
}

func flattenS3Settings(settings map[string]any) (s3DataSourceModel, error) {
	var m s3DataSourceModel
	var err error
	m.commonDataSourceModel, err = flattenCommonSettings(settings)
	if err != nil {
		return m, err
	}
	m.Bucket = strSettingNull(settings, settingBucket)
	m.Client = strSettingNull(settings, settingClient)
	m.BasePath = strSettingNull(settings, settingBasePath)
	m.ServerSideEncryption, err = boolSettingNull(settings, "server_side_encryption")
	if err != nil {
		return m, err
	}
	m.BufferSize = strSettingNull(settings, "buffer_size")
	m.CannedACL = strSettingNull(settings, "canned_acl")
	m.StorageClass = strSettingNull(settings, "storage_class")
	m.PathStyleAccess, err = boolSettingNull(settings, "path_style_access")
	if err != nil {
		return m, err
	}
	m.DisableChunkedEncoding, err = boolSettingNull(settings, settingDisableChunkedEncoding)
	if err != nil {
		return m, err
	}
	m.AlwaysSignRequests, err = boolSettingNull(settings, settingAlwaysSignRequests)
	if err != nil {
		return m, err
	}
	return m, nil
}

func flattenHDFSSettings(settings map[string]any) (hdfsDataSourceModel, error) {
	var m hdfsDataSourceModel
	var err error
	m.commonDataSourceModel, err = flattenCommonSettings(settings)
	if err != nil {
		return m, err
	}
	m.URI = strSettingNull(settings, settingURI)
	m.Path = strSettingNull(settings, settingPath)
	m.LoadDefaults, err = boolSettingNull(settings, "load_defaults")
	if err != nil {
		return m, err
	}
	return m, nil
}
