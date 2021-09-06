package aws

import (
	"fmt"

	"github.com/daniel-cohen/secretsfetcher/secrets"
	"go.uber.org/zap"
)

type ManifestSecretsFetcher struct {
	// implements secrets.SecretsFetcher
	zl       *zap.Logger
	provider *AWSSecretsManagerProvider

	manifest *SecretManifest
}

func NewManifestSecretFetcher(provider *AWSSecretsManagerProvider, manifest *SecretManifest, zl *zap.Logger) *ManifestSecretsFetcher {
	return &ManifestSecretsFetcher{
		provider: provider,
		manifest: manifest,
		zl:       zl,
	}
}

func (msf *ManifestSecretsFetcher) Fetch() ([]*secrets.Secret, error) {
	secretRes, err := msf.provider.FetchSecrets(msf.manifest.SecretObjects)
	if err != nil {
		msf.zl.Error("failed to fetch secrets from aws secrets provider",
			zap.Any("secretObjects", msf.manifest.SecretObjects),
			zap.Error(err))

		return nil, err
	}

	return secretRes, nil
}

type ListSecretFetcher struct {
	// implements secrets.SecretsFetcher
	zl              *zap.Logger
	provider        *AWSSecretsManagerProvider
	prefixFilter    string
	tagKeyFilters   []string
	tagValueFilters []string
}

func NewListSecretFetcher(
	provider *AWSSecretsManagerProvider,
	prefixFilter string,
	tagKeyFilters []string,
	tagValueFilters []string,
	zl *zap.Logger) *ListSecretFetcher {
	return &ListSecretFetcher{
		zl:              zl,
		provider:        provider,
		prefixFilter:    prefixFilter,
		tagKeyFilters:   tagKeyFilters,
		tagValueFilters: tagValueFilters,
	}
}

func (lsf *ListSecretFetcher) Fetch() ([]*secrets.Secret, error) {
	if lsf.prefixFilter == "" {
		lsf.zl.Error("prefix filter not set")
		return nil, fmt.Errorf("prefix filter cannot be empty ")
	}

	secretRes, err := lsf.provider.FetchAllSecrets(lsf.prefixFilter, lsf.tagKeyFilters, lsf.tagValueFilters)

	if err != nil {
		lsf.zl.Error("failed to fetch all secrets from aws secrets provider",
			zap.String("prefixFilter", lsf.prefixFilter),
			zap.Strings("tagKeyFilters", lsf.tagKeyFilters),
			zap.Strings("tagValueFilters", lsf.tagValueFilters),
			zap.Error(err))
		return nil, err
	}

	return secretRes, nil
}
