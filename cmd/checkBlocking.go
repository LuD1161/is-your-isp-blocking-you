/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/joeguo/tldextract"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// checkBlockingCmd represents the checkBlocking command
var checkBlockingCmd = &cobra.Command{
	Use:   "checkBlocking",
	Short: "Check blocking for your ISP.",
	Run: func(cmd *cobra.Command, args []string) {
		// Create a TLD extractor
		cache := "tld.cache"
		extract, _ := tldextract.New(cache, false)

		var filePath string
		var columnNumber int // column number where domain exists in csv file
		switch domainList {
		case "citizenlabs":
			filePath = "data/citizenlabs-lists/lists/global.csv"
			columnNumber = 0
		case "cisco":
			filePath = "data/cisco.csv"
			columnNumber = 1
		case "others":
			filePath = domainList
			// check if file exists
			if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
				log.Fatal().Msgf("File does not exists : %s", domainList)
			}
			columnNumber = 0
		default:
			filePath = "data/citizenlabs-lists/lists/global.csv"
		}
		records, err := ReadCsvFile(filePath)
		// records = records // TODO: Change this
		if err != nil {
			log.Fatal().Msg("Error in reading csv file")
		}

		urls := make(map[string]bool, 0)
		for _, record := range records {
			result := extract.Extract(record[columnNumber])
			url := result.Root + "." + result.Tld
			log.Debug().Msgf("url : %s", url)
			urls[url] = true
		}

		log.Info().Msgf("\n✅ Domain List : %s\n✅ Unique URLs : %d\n✅ Threads %d\n", domainList, len(urls), threads)
		urlsChan := make(chan string, threads)
		resultsChan := make(chan Result)
		start := time.Now()

		for i := 0; i < threads; i++ {
			go MakeRequest(urlsChan, resultsChan)
		}

		go func() {
			for url := range urls {
				url := fmt.Sprintf("https://%s", url)
				log.Debug().Msgf("Working on URL : %s", url)
				urlsChan <- url
			}
		}()

		results := make([]Record, 0)
		accessible, inaccessible, blocked, unknownHost, timedOut := make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0)
		for i := 0; i < len(urls); i++ {
			result := <-resultsChan
			record := Record{
				Website:    result.URL,
				ISP:        "",
				Country:    "",
				Location:   "",
				Accessible: false,
				ErrMsg:     "",
			}
			if result.Error != nil {
				if i%10000 == 0 {
					log.Debug().Msgf("✅ URLs done : %d", i)
					log.Error().Msgf("Error : %s", result.Error.Error())
				}
				record.ErrMsg = result.Error.Error()
				// truncate error message to avoid sql errors
				if len(record.ErrMsg) > 1024 {
					record.ErrMsg = record.ErrMsg[:1024] // max length of the column
				}
			}
			switch result.Code {
			case CONN_RESET:
				inaccessible = append(inaccessible, result.URL)
			case CONN_BLOCKED:
				blocked = append(blocked, result.URL)
			case CONN_OK:
				// fmt.Printf("✅ All good. %s", result.URL)
				record.Accessible = true
				accessible = append(accessible, result.URL)
			case CONN_UNKNOWN:
				inaccessible = append(inaccessible, result.URL)
				// fmt.Printf("✅ All good. %s", record)
			case NO_SUCH_HOST:
				unknownHost = append(unknownHost, result.URL)
			case CONN_TIMEOUT:
				timedOut = append(timedOut, result.URL)
			default:
				log.Error().Msgf("default case : %+v", result)
			}
			results = append(results, record)
		}
		scanTime := int(time.Since(start).Seconds()) // total seconds to complete the scan
		result, err := GetISP()
		if err != nil {
			log.Fatal().Msgf("Error unmarshalling data from ifconfig : %s", err.Error())
		}
		scanStats := ScanStats{
			Model:                gorm.Model{},
			ScanTime:             scanTime,
			UniqueDomainsScanned: len(urls),
			Accessible:           len(accessible),
			Inaccessible:         len(inaccessible),
			Blocked:              len(blocked),
			TimedOut:             len(timedOut),
			UnknownHost:          len(unknownHost),
			ISP:                  result.AsnOrg,
			Country:              result.Country,
			Location:             result.City,
			EvilISP:              false,
		}

		if len(blocked) > 0 {
			scanStats.EvilISP = true
		}
		close(urlsChan)
		close(resultsChan)
		// Print tabular output
		printTable(scanTime, result, scanStats)
	},
}

func printTable(scanTime int, result IfConfigResponse, scanStats ScanStats) {
	// Print ISP details
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.SetTitle("Current ISP Details")
	t.AppendHeader(table.Row{"Scan Time", "Country", "IP", "ISP", "Region", "City"})
	t.AppendRows([]table.Row{
		{fmt.Sprintf("%d seconds", scanTime), result.Country, result.IP, result.AsnOrg, result.RegionName, result.City},
	})
	t.AppendSeparator()
	t.Render()

	// Print scan stats
	t = table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.SetTitle(fmt.Sprintf("Current Scan Data | Domain List : %s", domainList))
	t.AppendHeader(table.Row{"Unique Domains", "Accessible", "Inaccessible", "Blocked", "Timed Out", "Unknown Host", "Good ISP"})
	evilISP := "✅"
	if scanStats.EvilISP {
		evilISP = "❌"
	}
	t.AppendRows([]table.Row{
		{scanStats.UniqueDomainsScanned, scanStats.Accessible, scanStats.Inaccessible, scanStats.Blocked, scanStats.TimedOut, scanStats.UnknownHost, evilISP},
	})
	t.AppendSeparator()
	t.Render()
}

func init() {
	rootCmd.AddCommand(checkBlockingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkBlockingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkBlockingCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
