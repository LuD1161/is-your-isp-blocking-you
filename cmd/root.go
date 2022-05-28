/*
Copyright © 2022 Aseem Shrey

*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	threads int
	timeout int
	rootCmd = &cobra.Command{
		Use:   "is-your-isp-blocking-you",
		Short: "A tool to test if your ISP is blocking your access to some parts of the Internet.",
		Long:  "This tool tries to get website content for a large number of websites and checks, whether it's accessible or not.",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		Version: "0.1",
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	b, err := ioutil.ReadFile("./cmd/banner.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().IntVarP(&threads, "threads", "t", 100, "No of threads")
	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "", 15, "Timeout for requests")

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.With().Str("version", rootCmd.Version).Logger()
	switch os.Getenv("LogLevel") {
	case "Debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

}
