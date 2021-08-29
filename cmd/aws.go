package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/daniel-cohen/secretsfetcher/secrets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultPathTranslation = "_"
	pathTranslationFalse   = "False"
)

var (
	manifestFile string
	outputFolder string
)

// func initConfig() {
// 	//pf.StringVarP(&manifestFile, "manifest", "m", "", "secrets manifest file")
// 	//cobra.MarkFlagRequired(pf, "manifest")
// }

// versionCmd represents the version command
var awsCmd = &cobra.Command{
	Use:   "aws",
	Short: "Shows the secretsfetcher service version",
	Run: func(cmd *cobra.Command, args []string) {
		v := viper.New()

		//zlCfg := zap.NewDevelopmentConfig()
		//zlCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zl := initLog(cfg.LogLevel)
		defer zl.Sync() // flushes buffer, if any

		region := ""
		pathTranslationChar := defaultPathTranslation
		var secretRes []*secrets.Secret

		// We're loading the manifest as  a viper config file:
		if manifestFile != "" {
			// Use config file from the flag.
			v.SetConfigFile(manifestFile)

			// If a config file is found, read it in.
			if err := v.ReadInConfig(); err == nil {
				zl.Info("Read manifest file", zap.String("manifestPath", manifestFile))
			} else {
				zl.Error("Failed to load manifest file", zap.Error(err))
			}

			//Put all the config in a common struct
			manifestCfg := &secrets.SecretProviderClass{}
			if err := v.Unmarshal(&manifestCfg); err != nil {
				zl.Fatal("Unable to decode into struct", zap.Error(err))

			}

			zl.Info("Loaded manifest config", zap.Any("manifestCfg", manifestCfg))

			region = manifestCfg.Region
			pathTranslationChar = manifestCfg.PathTranslation

			//TODO: refactor:
			provider, err := secrets.NewAWSSecretsManagerProvider(region, zl)
			if err != nil {
				zl.Fatal("failed to setup aws secrets provider", zap.Error(err))
			}

			secretRes, err = provider.FetchSecrets(manifestCfg.SecretObjects)
			if err != nil {
				zl.Fatal("failed to fetch secrets from aws secrets provider",
					zap.Any("secretObjects", manifestCfg.SecretObjects),
					zap.Error(err))
			}

		} else {
			zl.Info("no manifest set")
			if cfg.Aws == nil || cfg.Aws.PrefixFilter == "" {
				zl.Fatal("no manifest and aws prefix filter not set")
			}

			region = cfg.Aws.Region
			pathTranslationChar = cfg.Aws.PathTranslation

			//TODO: refactor:
			provider, err := secrets.NewAWSSecretsManagerProvider(region, zl)
			if err != nil {
				zl.Fatal("failed to setup aws secrets provider", zap.Error(err))
			}

			secretRes, err = provider.FetchAllSecrets(cfg.Aws.PrefixFilter, cfg.Aws.TagFilter)

			if err != nil {
				zl.Fatal("failed to fetch all secrets from aws secrets provider",
					zap.String("prefixFilter", cfg.Aws.PrefixFilter),
					zap.Any("tagFilter", cfg.Aws.TagFilter),
					zap.Error(err))
			}
		}

		for _, v := range secretRes {
			zl.Debug("secret fetched", zap.String("secretName", v.Name))

			outputFileName := v.Name
			if pathTranslationChar != "" {
				if pathTranslationChar != pathTranslationFalse {
					outputFileName = strings.ReplaceAll(outputFileName, "/", pathTranslationChar)
				}
			} else {
				outputFileName = strings.ReplaceAll(outputFileName, "/", defaultPathTranslation)

			}

			outputFilePath := path.Join(outputFolder, outputFileName)

			zl.Debug("writing secret to file",
				zap.String("file_path", outputFilePath),
				zap.String("secret_name", v.Name),
				//zap.String("secret_name", v.Content),
			)
			if err := ioutil.WriteFile(outputFilePath, []byte(v.Content), os.ModePerm); err != nil {
				zl.Error("failed to write file", zap.String("file_path", outputFilePath), zap.Error(err))
				// Skip it and move on
				continue
			}

		}

		os.Exit(0)
	},
}

func init() {
	awsCmd.Flags().StringVarP(&manifestFile, "manifest", "m", "", "secrets manifest file")
	awsCmd.Flags().StringVarP(&outputFolder, "output", "o", "", "output folder")

	// it's no longer required (we can do a wild card search)
	//awsCmd.MarkFlagRequired("manifest")

	rootCmd.AddCommand(awsCmd)
}

func initLog(logLevel string) *zap.Logger {
	level := zapcore.InfoLevel
	if err := level.Set(logLevel); err != nil {
		log.Fatalf("could not set zap log level to: \"%s\" \n", logLevel)
	}

	config := &zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(level),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "msg",

			LevelKey:    "lvl",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "ts",
			EncodeTime: zapcore.RFC3339NanoTimeEncoder,

			CallerKey:      "src",
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,

			StacktraceKey: "stk",
		},
	}

	zl, err := config.Build()
	if err != nil {
		log.Fatalf("failed to build zap logger. Error: %s", err)
	}
	return zl

}
