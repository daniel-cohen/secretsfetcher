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
	// region := a.awsCfg.Region
	// // the manifest will take precidence over the region in the main config
	// if msf.manifest.Region != "" {
	// 	region = msf.manifest.Region
	// }

	//TODO: refactor:
	// provider, err := NewAWSSecretsManagerProvider(region, msf.zl)
	// if err != nil {
	// 	msf.zl.Fatal("failed to setup aws secrets provider", zap.Error(err))
	// }

	// validate that the region on the provider is the same the region we're trying to query:
	// if region != msf.provider.Region() {
	// 	return nil, fmt.Errorf("aws provider region(%s) and manifest region(%s) do not  match", msf.provider.Region(), region)
	// }

	secretRes, err := msf.provider.FetchSecrets(msf.manifest.SecretObjects)
	if err != nil {
		msf.zl.Error("failed to fetch secrets from aws secrets provider",
			zap.Any("secretObjects", msf.manifest.SecretObjects),
			zap.Error(err))

		return nil, err
	}

	return secretRes, nil
}

//----------------------------------------------------------------
type ListSecretFetcher struct {
	// implements secrets.SecretsFetcher
	zl           *zap.Logger
	provider     *AWSSecretsManagerProvider
	prefixFilter string
	tagFilter    map[string]string
}

func NewListSecretFetcher(
	provider *AWSSecretsManagerProvider,
	prefixFilter string,
	tagFilter map[string]string,
	zl *zap.Logger) *ListSecretFetcher {
	return &ListSecretFetcher{
		zl:           zl,
		provider:     provider,
		prefixFilter: prefixFilter,
		tagFilter:    tagFilter,
	}
}

func (lsf *ListSecretFetcher) Fetch() ([]*secrets.Secret, error) {
	//TODO: Would we want allow a tag filter search only ?
	if lsf.prefixFilter == "" {
		lsf.zl.Error("prefix filter not set")
		return nil, fmt.Errorf("prefix filter cannot be empty ")
	}

	secretRes, err := lsf.provider.FetchAllSecrets(lsf.prefixFilter, lsf.tagFilter)

	if err != nil {
		lsf.zl.Error("failed to fetch all secrets from aws secrets provider",
			zap.String("prefixFilter", lsf.prefixFilter),
			zap.Any("tagFilter", lsf.tagFilter),
			zap.Error(err))
		return nil, err
	}

	return secretRes, nil
}
