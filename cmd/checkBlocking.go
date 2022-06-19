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

		// Get ISP first, to get the country
		proxyTransport := SetProxyTransport()
		ispResult, err := GetISP(proxyTransport)
		if err != nil {
			log.Fatal().Msgf("Error getting ISP data. Probably internet not connected or ifconfig.co is blocked ( unlikely, check this in your browser or terminal ).")
		}

		switch domainList {
		case "citizenlabs":
			filePath = "data/citizenlabs-lists/lists/global.csv"
			columnNumber = 0
		case "cisco":
			filePath = "data/cisco.csv"
			columnNumber = 1
		default:
			filePath = domainList
			// check if file exists
			if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
				// Get country specific citizenlabs' list
				countrySpecificList := fmt.Sprintf("data/citizenlabs-lists/lists/%s.csv", ispResult.CountryISO)
				log.Error().Msgf("User specified file does not exists : %s. Switching to country ( %s ) specific list : %s", domainList, ispResult.Country, countrySpecificList)
				filePath = countrySpecificList
				if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
					filePath = "data/citizenlabs-lists/lists/global.csv"
					log.Error().Msgf("Couldn't find country specific list ( %s ). Using global list - %s", countrySpecificList, filePath)
				}
			}
			columnNumber = 0
		}
		records, err := ReadCsvFile(filePath)
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

		log.Info().Msgf("\n✅ Domain List : %s\n✅ Unique URLs : %d\n✅ Threads %d\n", filePath, len(urls), threads)
		urlsChan := make(chan string, threads)
		resultsChan := make(chan Result)
		start := time.Now()
		log.Info().Msgf("Started scan with ID : %s", scanId)
		for i := 0; i < threads; i++ {
			go MakeRequest(urlsChan, resultsChan, proxyTransport)
		}

		go func() {
			for url := range urls {
				url := fmt.Sprintf("http://%s", url)
				log.Debug().Msgf("Working on URL : %s", url)
				urlsChan <- url
			}
		}()

		results := make([]Record, 0)
		accessible, inaccessible, blocked, unknownHost, timedOut := make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0)
		for i := 0; i < len(urls); i++ {
			result := <-resultsChan
			record := Record{
				Model:          gorm.Model{},
				ScanId:         scanId,
				Website:        result.URL,
				ISP:            ispResult.AsnOrg,
				Country:        ispResult.Country,
				Location:       ispResult.City,
				Latitude:       ispResult.Latitude,
				Longitude:      ispResult.Longitude,
				Accessible:     false,
				Data:           result.Data,
				ErrMsg:         "",
				HTTPStatusCode: result.HTTPStatusCode,
				HTMLTitle:      result.HTMLTitle,
				HTMLBodyLength: result.HTMLBodyLength,
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
		if err != nil {
			log.Fatal().Msgf("Error unmarshalling data from ifconfig : %s", err.Error())
		}
		evilISP := false
		if len(blocked) > 0 {
			evilISP = true
		}
		scanStats := ScanStats{
			Model:                gorm.Model{},
			ScanId:               scanId,
			ScanTime:             scanTime,
			UniqueDomainsScanned: len(urls),
			Accessible:           len(accessible),
			Inaccessible:         len(inaccessible),
			Blocked:              len(blocked),
			TimedOut:             len(timedOut),
			UnknownHost:          len(unknownHost),
			ISP:                  ispResult.AsnOrg,
			Country:              ispResult.Country,
			Location:             ispResult.City,
			Latitude:             ispResult.Latitude,
			Longitude:            ispResult.Longitude,
			DomainList:           filePath,
			EvilISP:              evilISP,
		}
		db, err := initialiseDB(storeInDB, scanId)
		if err != nil {
			log.Error().Stack().Err(err).Msgf("Error initialising DB : %s", err.Error())
		} else if db != nil {
			// if no err and db is initialized
			// this will save the results to DB if db_url is passed
			if err := saveToDB(db, results, scanStats); err != nil {
				log.Error().Stack().Err(err).Msgf("Error saving results in DB : %s", err.Error())
			}
		}
		if len(blocked) > 0 {
			scanStats.EvilISP = true
		}
		// ToDo : Change this to dependent on a CLI flag
		if os.Getenv("LogLevel") == "Debug" {
			fmt.Printf("blocked URLs : %s\n", blocked)
			fmt.Printf("Inaccessible URLs : %s\n", inaccessible)
		}
		close(urlsChan)
		close(resultsChan)
		// Print tabular output
		printTable(scanTime, ispResult, scanStats, filePath)
	},
}

func printTable(scanTime int, result IfConfigResponse, scanStats ScanStats, filePath string) {
	// Print ISP details
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.SetTitle("Current ISP Details")
	t.AppendHeader(table.Row{"Scan Time", "Country", "IP", "ISP", "Region", "City", "Lat", "Lon"})
	t.AppendRows([]table.Row{
		{fmt.Sprintf("%d seconds", scanTime), result.Country, result.IP, result.AsnOrg, result.RegionName, result.City, result.Latitude, result.Longitude},
	})
	t.AppendSeparator()
	t.Render()

	// Print scan stats
	t = table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.SetTitle(fmt.Sprintf("Current Scan Data | Domain List : %s", filePath))
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
