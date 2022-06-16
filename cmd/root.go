/*
Copyright © 2022 Aseem Shrey

*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var (
	threads    int
	timeout    int
	domainList string
	proxyURL   string
	storeInDB  string
	scanId     = fmt.Sprintf("%d-%s", time.Now().Unix(), GenerateRandomString(10))
	rootCmd    = &cobra.Command{
		Use:     "is-your-isp-blocking-you",
		Short:   "A tool to test if your ISP is blocking your access to some parts of the Internet.",
		Long:    "This tool tries to get website content for a large number of websites and checks, whether it's accessible or not.",
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
	rootCmd.PersistentFlags().StringVarP(&domainList, "domain_list", "l", "citizenlabs", "Domain list to choose from. Valid options : 'citizenlabs','cisco', 'others'. Choosing 'others' you need to specify the full path of the list.")
	rootCmd.PersistentFlags().StringVarP(&proxyURL, "proxy_url", "p", "", "Proxy URL to pass traffic through. The URL format : http(s)://<username>:<password>@proxy.website.com:<proxy_port> e.g. - http://localhost:8080 .The tool will try to fetch all the domains through it. This is useful, when you want to test blocking on another ISP/country etc.")
	rootCmd.PersistentFlags().StringVarP(&storeInDB, "store_in_db", "d", "", "If you want to save the results to db pass in the DB type. Valid choices : 'postgres', 'sqlite', 'mysql'. Also make sure to populate the `set-env-vars.sh` file with the respective env vars for the db.")

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	switch os.Getenv("LogLevel") {
	case "Debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.With().Timestamp().Stack().Caller().Logger()
		log.Logger = log.With().Str("version", rootCmd.Version).Logger()
	case "Error":
		log.Logger = log.With().Timestamp().Stack().Caller().Logger()
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Handle exit gracefully
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Info().Msg("Got signal to close the program")
		os.Exit(0)
	}()

}