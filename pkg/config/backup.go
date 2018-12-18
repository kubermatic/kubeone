package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// BackupConfig describes where and how to store Ark backups
type BackupConfig struct {
	// Provider is provider for buckets and volume snapshots.
	// Possible values are: AWS (includes compatible AWS S3 storages), Azure and GCP
	// TODO(xmudrii): By default uses specified control plane provider if compatible with Ark
	Provider string `json:"provider"`

	// S3AccessKey is Access Key used to access backups S3 bucket.
	// This variable is sourced from BACKUP_AWS_ACCESS_KEY_ID,
	// or if unset from AWS_ACCESS_KEY_ID environment variable
	S3AccessKey string `json:"s3_access_key"`
	// S3SecretAccessKey is secret key used to access backups S3 bucket.
	// This variable is sourced from BACKUP_AWS_SECRET_ACCESS_KEY environment variable,
	// or if unset from AWS_SECRET_ACCESS_KEY environment variable
	S3SecretAccessKey string `json:"s3_secret_access_key"`

	// BucketName is name of the S3 bucket where backups are stored
	BucketName string `json:"bucket_name"`

	// BackupStorageConfig is optional configuration depending on the provider specified
	// Details: https://heptio.github.io/ark/v0.10.0/api-types/backupstoragelocation.html
	BackupStorageConfig map[string]string `json:"backup_storage_config"`

	// VolumesSnapshotConfig is optional configuration depending on the provider specified
	// Details: https://heptio.github.io/ark/v0.10.0/api-types/volumesnapshotlocation.html
	VolumesSnapshotConfig map[string]string `json:"volumes_snapshot_region"`
}

// Enabled checks if a provider is set and Ark should be deployed.
func (m *BackupConfig) Enabled() bool {
	return m.Provider != ""
}

// Validate valides the BackupConfig structure, ensuring credentials and bucket name are provided
func (m *BackupConfig) Validate() error {
	// if the backup is not enabled, nothing else matters
	if !m.Enabled() {
		return nil
	}

	if len(m.S3AccessKey) == 0 {
		return errors.New("S3 access key must be given")
	}

	if len(m.S3SecretAccessKey) == 0 {
		return errors.New("S3 secret access key must be given")
	}

	if len(m.BucketName) == 0 {
		return errors.New("S3 bucket name must be given")
	}

	if m.Provider != "aws" && m.Provider != "azure" && m.Provider != "gcp" {
		return fmt.Errorf("invalid provider %s; supported values: \"aws\", \"azure\" or \"gcp\"", m.Provider)
	}

	return nil
}

// ApplyEnvironment reads credentials from environment variables,
// returning an error if a required variable is not set.
func (m *BackupConfig) ApplyEnvironment() error {
	const envPrefix = "env:"

	if strings.HasPrefix(m.S3AccessKey, envPrefix) {
		envName := strings.TrimPrefix(m.S3AccessKey, envPrefix)
		m.S3AccessKey = os.Getenv(envName)
	}

	if strings.HasPrefix(m.S3SecretAccessKey, envPrefix) {
		envName := strings.TrimPrefix(m.S3SecretAccessKey, envPrefix)
		m.S3SecretAccessKey = os.Getenv(envName)
	}

	return nil
}
