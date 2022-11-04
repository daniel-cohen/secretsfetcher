package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/daniel-cohen/secretsfetcher/secrets/aws"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version defines the build version of the service, injected during build.
var (
	version        string
	cfgFile        string
	cfg            config
	consoleLogging bool
)

//
// rootCmd represents the base command when called without any subcommands
//
var rootCmd = &cobra.Command{
	Use:   "secretsfetcher",
	Short: "A tool to fetch aws secrets to local json files.",
	Run: func(cmd *cobra.Command, args []string) {

	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		return nil
	},
}

// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ver string) {
	version = ver

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	pf := rootCmd.PersistentFlags()
	pf.StringVar(&cfgFile, "config", "config.yaml", "config file (default is ./config.yaml)")
	//pf.StringVar(&cfgFile.loglevel, "loglevel", "info", "log level (default is info)")

	pf.String("loglevel", "info", "log level (default is info)")
	pf.BoolVar(&consoleLogging, "consolelog", false, "console logger for debugging instead of the json formatted logging")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("APP") // to make this more explicit

	// Note: only fields which are read from the config will be loaded from ENV: https://github.com/spf13/viper/issues/584
	// We can use viper.SetDefault("") https://github.com/spf13/viper#unmarshaling

	//viper.SetDefault("LogLevel", "warn")
	// Bind the cli flag so we can override it in ENV/Config:
	viper.BindPFlag("loglevel", rootCmd.Flags().Lookup("loglevel"))

	///-----------------------------------------------------------------

	//TODO: See if I can refactor this into the aws.go command:
	if awsCmd.Flags().Lookup("tagkeys") != nil {
		viper.BindPFlag("Aws.TagKeyFilters", awsCmd.Flags().Lookup("tagkeys"))
	}

	if awsCmd.Flags().Lookup("tagvalues") != nil {
		viper.BindPFlag("Aws.TagValueFilters", awsCmd.Flags().Lookup("tagvalues"))
	}

	if awsCmd.Flags().Lookup("prefix") != nil {
		viper.BindPFlag("Aws.PrefixFilter", awsCmd.Flags().Lookup("prefix"))
	}

	///-----------------------------------------------------------------

	// Set specific (even if empty) defaults so we can load them from ENV even if the config is not loaded:
	viper.SetDefault("Aws.PrefixFilter", "")
	viper.SetDefault("Aws.TagKeyFilters", []string{})
	viper.SetDefault("Aws.TagValueFilters", []string{})
	viper.SetDefault("Aws.PathTranslation", aws.DefaultPathTranslation)
	viper.SetDefault("Aws.Region", "")
	viper.SetDefault("Aws.TagFilter", map[string]string{})

	viper.AutomaticEnv()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Read config file:", viper.ConfigFileUsed())
	}

	//Put all the config in a common struct
	err := viper.Unmarshal(&cfg)
	if err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}
}
