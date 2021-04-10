package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve live data",
	Long:  `serve live data in either bar or tick format`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			fmt.Println("check help for usage info")
			os.Exit(1)
		}
		reso := args[0]
		if reso != "bar" && reso != "tick" {
			fmt.Println("Please choose either bar or tick")
			os.Exit(1)
		}
		broadport, _ := cmd.Flags().GetInt("broadport")
		fmt.Println("serveing", reso, "of", args[1], "to 127.0.0.1:"+strconv.FormatInt(int64(broadport), 10))
		select {}
	},
}

func init() {
	serveCmd.Flags().IntP("broadport", "b", 8888, `port number to broadcast`)
}
