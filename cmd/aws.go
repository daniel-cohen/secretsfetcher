package cmd

import (
	"os"

	"github.com/daniel-cohen/secretsfetcher/secrets"
	"github.com/daniel-cohen/secretsfetcher/secrets/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "fetched secretes from aws secrets manager",
	Run: func(cmd *cobra.Command, args []string) {

		// Init logging:
		zl := initLog(cfg.LogLevel, consoleLogging)
		defer zl.Sync() // flushes buffer, if any

		outputFolder, err := cmd.Flags().GetString("output")
		if err != nil {
			zl.Fatal("failed to get the output flag")
		}

		manifestFile, err := cmd.Flags().GetString("manifest")
		if err != nil {
			zl.Fatal("failed to get the manifest flag")
		}

		region := cfg.Aws.Region // we set it to default to empty string
		pathTranslationChar := aws.DefaultPathTranslation
		//var secretRes []*secrets.Secret

		var sf secrets.SecretsFetcher

		var manifestCfg *aws.SecretManifest
		// We're loading the manifest as  a viper config file:
		if manifestFile != "" {
			// viper instance for the manifest
			v := viper.New()

			// Use config file from the flag.
			v.SetConfigFile(manifestFile)

			// If a config file is found, read it in.
			if err := v.ReadInConfig(); err != nil {
				zl.Fatal("Failed to load manifest file", zap.String("manifestPath", manifestFile), zap.Error(err))

			}
			zl.Info("Read manifest file", zap.String("manifestPath", manifestFile))

			//Put all the config in a common struct
			manifestCfg = &aws.SecretManifest{}
			if err := v.Unmarshal(&manifestCfg); err != nil {
				zl.Fatal("Unable to decode into struct", zap.Error(err))
			}

			zl.Info("Loaded manifest config", zap.Any("manifestCfg", manifestCfg))

			// the manifest will take precedence over the region in the main config
			if manifestCfg.Region != "" {
				region = manifestCfg.Region
			}

			if manifestCfg.PathTranslation != "" {
				pathTranslationChar = manifestCfg.PathTranslation
			}
		} else {
			zl.Info("no manifest set")
			if cfg.Aws == nil || cfg.Aws.PrefixFilter == "" {
				zl.Fatal("no manifest and aws prefix filter not set")
			}
		}

		provider, err := aws.NewAWSSecretsManagerProvider(region, zl)
		if err != nil {
			zl.Fatal("failed to setup aws secrets provider", zap.Error(err))
		}

		if manifestCfg != nil {
			sf = aws.NewManifestSecretFetcher(provider, manifestCfg, zl)
		} else {
			sf = aws.NewListSecretFetcher(provider, cfg.Aws.PrefixFilter, cfg.Aws.TagKeyFilters, cfg.Aws.TagValueFilters, zl)
		}

		secretRes, err := sf.Fetch()
		if err != nil {
			zl.Fatal("failed to fetch secrets", zap.Error(err))
		}

		sw := secrets.NewFileSecretWriter(outputFolder, pathTranslationChar, zl).StopOnError()
		err = sw.WriteSecrets(secretRes)
		if err != nil {
			zl.Fatal("failed to write secrets", zap.Error(err))
		}

		os.Exit(0)
	},
}

func init() {
	awsCmd.Flags().StringP("manifest", "m", "", "secrets manifest file")
	awsCmd.Flags().StringP("output", "o", "", "output folder. Will default to the current working folder")

	awsCmd.Flags().StringSlice("tagkeys", []string{}, "an array of tag key prefixes of filters to find secerts by. Example: --tagkeys=app,secret-type")
	awsCmd.Flags().StringSlice("tagvalues", []string{}, "an array of tag value prefixes of filters to find secerts by. Example: --tagvalues=my-app-name,b44c6886-96c4-4b4d-b267-30d7c5787b1a")

	awsCmd.Flags().String("prefix", "", "a prefix for all secrets to fetch")

	rootCmd.AddCommand(awsCmd)
}
