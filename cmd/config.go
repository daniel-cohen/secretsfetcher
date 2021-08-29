package cmd

// Config - config vars for the application
type config struct {
	LogLevel string

	Aws *AWSConfig
}

type AWSConfig struct {
	PrefixFilter    string
	TagFilter       map[string]string
	Region          string
	PathTranslation string
}
