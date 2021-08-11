package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version defines the build version of the service, injected during build.
var (
	version string
	cfgFile string
	cfg     config
)

//
// rootCmd represents the base command when called without any subcommands
//
var rootCmd = &cobra.Command{
	Use:   "secretsfetcher",
	Short: "A service capable to perform segment related operations",
	Run: func(cmd *cobra.Command, args []string) {

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
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("APP") // to make this more explicit

	// Note: only fields which are read from the config will be loaded from ENV: https://github.com/spf13/viper/issues/584
	// We can use viper.SetDefault("") https://github.com/spf13/viper#unmarshaling
	viper.SetDefault("LogLevel", "warn")

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
