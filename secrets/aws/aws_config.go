package aws

type AWSConfig struct {
	PrefixFilter string

	TagKeyFilters   []string
	TagValueFilters []string

	Region          string
	PathTranslation string
}
