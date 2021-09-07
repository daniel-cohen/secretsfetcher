package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the secretsfetcher service version",
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println(rootCmd.Use + " " + version)
		//fmt.Println(rootCmd.Use)
		fmt.Println(rootCmd.Use+" Version:", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
