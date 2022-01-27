package config

/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// copied from https://github.com/kubernetes/apiserver/blob/v0.23.1/pkg/apis/config/types.go
// because upstream struct do not have proper yaml definition for marshaling,
// it was impossible to get the output into the right format

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EncryptionConfiguration stores the complete configuration for encryption providers.
type EncryptionConfiguration struct {
	Kind       string `yaml:"kind,omitempty"`
	APIVersion string `yaml:"apiVersion,omitempty"`
	// resources is a list containing resources, and their corresponding encryption providers.
	Resources []ResourceConfiguration `yaml:"resources,omitempty"`
}

// ResourceConfiguration stores per resource configuration.
type ResourceConfiguration struct {
	// resources is a list of kubernetes resources which have to be encrypted.
	Resources []string `yaml:"resources,omitempty"`
	// providers is a list of transformers to be used for reading and writing the resources to disk.
	// eg: aesgcm, aescbc, secretbox, identity.
	Providers []ProviderConfiguration `yaml:"providers,omitempty"`
}

// ProviderConfiguration stores the provided configuration for an encryption provider.
type ProviderConfiguration struct {
	// aesgcm is the configuration for the AES-GCM transformer.
	AESGCM *AESConfiguration `yaml:"aesgcm,omitempty"`
	// aescbc is the configuration for the AES-CBC transformer.
	AESCBC *AESConfiguration `yaml:"aescbc,omitempty"`
	// secretbox is the configuration for the Secretbox based transformer.
	Secretbox *SecretboxConfiguration `yaml:"secretbox,omitempty"`
	// identity is the (empty) configuration for the identity transformer.
	Identity *IdentityConfiguration `yaml:"identity,omitempty"`
	// kms contains the name, cache size and path to configuration file for a KMS based envelope transformer.
	KMS *KMSConfiguration `yaml:"kms,omitempty"`
}

// AESConfiguration contains the API configuration for an AES transformer.
type AESConfiguration struct {
	// keys is a list of keys to be used for creating the AES transformer.
	// Each key has to be 32 bytes long for AES-CBC and 16, 24 or 32 bytes for AES-GCM.
	Keys []Key `yaml:"keys,omitempty"`
}

// SecretboxConfiguration contains the API configuration for an Secretbox transformer.
type SecretboxConfiguration struct {
	// keys is a list of keys to be used for creating the Secretbox transformer.
	// Each key has to be 32 bytes long.
	Keys []Key `yaml:"keys,omitempty"`
}

// Key contains name and secret of the provided key for a transformer.
type Key struct {
	// name is the name of the key to be used while storing data to disk.
	Name string `yaml:"name"`
	// secret is the actual key, encoded in base64.
	Secret string `yaml:"secret"`
}

// IdentityConfiguration is an empty struct to allow identity transformer in provider configuration.
type IdentityConfiguration struct{}

// KMSConfiguration contains the name, cache size and path to configuration file for a KMS based envelope transformer.
type KMSConfiguration struct {
	// name is the name of the KMS plugin to be used.
	Name string `yaml:"name"`
	// cacheSize is the maximum number of secrets which are cached in memory. The default value is 1000.
	// +optional
	CacheSize int32 `yaml:"cache_size"`
	// endpoint is the gRPC server listening address, for example "unix:///var/run/kms-provider.sock".
	Endpoint string `yaml:"endpoint"`
	// Timeout for gRPC calls to kms-plugin (ex. 5s). The default is 3 seconds.
	// +optional
	Timeout *metav1.Duration `yaml:"timeout,omitempty"`
}
