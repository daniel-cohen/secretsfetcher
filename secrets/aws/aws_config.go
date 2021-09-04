package aws

type AWSConfig struct {
	PrefixFilter    string
	TagFilter       map[string]string
	Region          string
	PathTranslation string
}
