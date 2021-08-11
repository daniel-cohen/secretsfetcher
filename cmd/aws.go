package cmd

import (
	"fmt"
	"io/ioutil"
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
		//fmt.Println(rootCmd.Use + " " + version)
		//fmt.Println(rootCmd.Use)
		//fmt.Println(rootCmd.Use+" Version:", version)

		v := viper.New()

		if cfgFile != "" {
			// Use config file from the flag.
			v.SetConfigFile(manifestFile)
		}

		// If a config file is found, read it in.
		if err := v.ReadInConfig(); err == nil {
			fmt.Println("Read manifest file:", v.ConfigFileUsed())
		} else {
			fmt.Printf("Failed to load manifest file. Error: %v\n", err)
		}

		//Put all the config in a common struct
		manifestCfg := &secrets.SecretProviderClass{}
		if err := v.Unmarshal(&manifestCfg); err != nil {
			fmt.Printf("Unable to decode into struct, %v", err)
			os.Exit(1)
		}

		fmt.Printf("Loaded manifest :, %v", manifestCfg)

		zlCfg := zap.NewDevelopmentConfig()
		zlCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zl, _ := zlCfg.Build()
		defer zl.Sync() // flushes buffer, if any

		//TODO: pass as a command flag/ENV:
		//outputFolderPath := "/secretsfetcher/secrets"
		//outputFolderPath := "/mnt/c/temp/secrets"

		provider, err := secrets.NewAWSSecretsManagerProvider(manifestCfg, zl)
		if err != nil {
			zl.Fatal("failed to setup aws secrets provider", zap.Error(err))
		}

		secrets, err := provider.FetchSecrets(manifestCfg.SecretObjects)
		for _, v := range secrets {
			zl.Debug("secret fetched",
				zap.String("secret_name", v.Name),
				zap.String("secret_name", v.Content),
			)

			outputFileName := v.Name
			if manifestCfg.PathTranslation != "" {
				if manifestCfg.PathTranslation != pathTranslationFalse {
					outputFileName = strings.ReplaceAll(outputFileName, "/", manifestCfg.PathTranslation)
				}
			} else {
				outputFileName = strings.ReplaceAll(outputFileName, "/", defaultPathTranslation)

			}

			outputFilePath := path.Join(outputFolder, outputFileName)
			//outputDirPath := path.Dir(outputFilePath)

			// Make sure the directorys are created if necessary (e.g if the secret name is: "a/b/c", we'll create the "a/b" folder structure )
			// if err := os.MkdirAll(outputDirPath, os.ModeDir|744); err != nil {
			// 	zl.Error("failed to create folder structure",
			// 		zap.String("file_path", outputFilePath),
			// 		zap.String("dir_path", outputDirPath),
			// 		zap.Error(err))
			// 	// Skip it and move on
			// 	continue
			// }

			zl.Debug("writing secret to file",
				zap.String("file_path", outputFilePath),
				zap.String("secret_name", v.Name),
				zap.String("secret_name", v.Content),
			)
			if err := ioutil.WriteFile(outputFilePath, []byte(v.Content), os.ModePerm); err != nil {
				zl.Error("failed to write file", zap.String("file_path", outputFilePath), zap.Error(err))
				// Skip it and move on
				continue
			}

		}

		if err != nil {
			zl.Fatal("failed get secrets from aws secrets provider", zap.Error(err))
		}

		os.Exit(0)
	},
}

func init() {
	awsCmd.Flags().StringVarP(&manifestFile, "manifest", "m", "", "secrets manifest file")
	awsCmd.Flags().StringVarP(&outputFolder, "output", "o", "", "output folder")
	awsCmd.MarkFlagRequired("manifest")

	rootCmd.AddCommand(awsCmd)
}
