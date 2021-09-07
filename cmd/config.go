package cmd

import (
	"github.com/daniel-cohen/secretsfetcher/secrets/aws"
)

// Config - config vars for the application
type config struct {
	LogLevel string

	Aws *aws.AWSConfig
}
