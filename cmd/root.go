package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Port int

// RootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ibgocli",
	Short: "ibgocli is cli built on top of TWS API",
	Long:  `ibgocli is a CLI to download, stream, and store market data through Interactive Brokers TWS API`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.PersistentFlags().IntVarP(&Port, "port", "p", 4002, "TWS/Gateway port")
}
