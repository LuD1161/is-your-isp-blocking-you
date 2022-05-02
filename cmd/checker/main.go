package main

import (
	"fmt"

	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LuD1161/is-your-isp-blocking-you/cmd/checker/models"
	"github.com/LuD1161/is-your-isp-blocking-you/internal/constants"
	"github.com/LuD1161/is-your-isp-blocking-you/internal/helpers"
	"github.com/joeguo/tldextract"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	// set log settings
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.With().Caller().Logger()
	log.Logger = log.With().Str("git_commit", GitCommit).Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

var GitCommit string // set using go build ldflags "-X main.GitCommit"

func main() {
	// Create a TLD extractor
	cache := "/tmp/tld.cache"
	extract, _ := tldextract.New(cache, false)
	// Download zip file
	// http://s3-us-west-1.amazonaws.com/umbrella-static/top-1m.csv.zip
	// top_1M_csv_zip := "http://s3-us-west-1.amazonaws.com/umbrella-static/top-1m.csv.zip"
	// if helpers.DownloadPackage(top_1M_csv_zip, "/tmp/top-1m.csv.zip") != nil {
	// 	return
	// }
	// helpers.Unzip("/tmp/top-1m.csv.zip")
	maxGoroutines := 1000
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Info().Msg("Got signal to close the program")
		os.Exit(0)
	}()

	start := time.Now()
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, "postgres", "postgres", "site_checker")
	db, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{})
	dbn, err := db.DB()
	if err != nil {
		panic(err.Error())
	}
	defer dbn.Close()
	db.AutoMigrate(&models.Record{}, &models.ScanStats{})

	records, err := helpers.ReadCsvFile("top-1m.csv")
	records = records[:10000] // TODO: Change this
	if err != nil {
		log.Fatal().Msg("Error in reading csv file")
	}

	urls := make(map[string]bool, 0)
	for _, record := range records {
		result := extract.Extract(record[1])
		url := result.Root + "." + result.Tld
		log.Info().Msgf("url : %s", url)
		urls[url] = true
		fmt.Printf("%+v;%s\n", result, record[1])
	}
	log.Info().Msgf("Unique URLs : %d", len(urls))
	urlsChan := make(chan string, maxGoroutines)
	resultsChan := make(chan models.Result)
	// records = [][]string{
	// 	{"1", "anonfile.com"}
	// }

	for i := 0; i < maxGoroutines; i++ {
		go helpers.MakeRequest(urlsChan, resultsChan)
	}

	go func() {
		for url, _ := range urls {
			url := fmt.Sprintf("https://%s", url)
			log.Debug().Msgf("Working on URL : %s", url)
			urlsChan <- url
		}
	}()

	results := make([]models.Record, 0)
	accessible, inaccessible, blocked, unknownHost, timedOut := make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0), make([]string, 0)
	for i := 0; i < len(urls); i++ {
		result := <-resultsChan
		record := models.Record{
			Website:    result.URL,
			ISP:        "airtel-in",
			Country:    "India",
			Location:   "Bangalore",
			Accessible: false,
			ErrMsg:     "",
		}
		if result.Error != nil {
			if i%10000 == 0 {
				log.Info().Msgf("✅ URLs done : %d", i)
				log.Error().Msgf("Error : %s", result.Error.Error())
			}
			record.ErrMsg = result.Error.Error()
		}
		switch result.Code {
		case constants.CONN_RESET:
			inaccessible = append(inaccessible, result.URL)
		case constants.CONN_BLOCKED:
			blocked = append(blocked, result.URL)
		case constants.CONN_OK:
			// fmt.Printf("✅ All good. %s", result.URL)
			record.Accessible = true
			accessible = append(accessible, result.URL)
		case constants.CONN_UNKNOWN:
			inaccessible = append(inaccessible, result.URL)
			// fmt.Printf("✅ All good. %s", record)
		case constants.NO_SUCH_HOST:
			unknownHost = append(unknownHost, result.URL)
		case constants.CONN_TIMEOUT:
			timedOut = append(timedOut, result.URL)
		default:
			log.Error().Msgf("default case : %+v", result)
		}
		results = append(results, record)
	}

	// log.Info().Msgf("results[0] : %+v\n", results[0])
	// err = helpers.InsertIntoDB(db, results)
	db.CreateInBatches(results, 1000)
	db.Commit()
	if err != nil {
		log.Error().Msgf("Error inserting into DB %s", err.Error())
	}

	scanTime := int(time.Since(start).Seconds()) // total seconds to complete the scan
	scanStats := &models.ScanStats{
		Model:                gorm.Model{},
		ScanTime:             scanTime,
		UniqueDomainsScanned: len(urls),
		Accessible:           len(accessible),
		Inaccessible:         len(inaccessible),
		Blocked:              len(blocked),
		TimedOut:             len(timedOut),
		UnknownHost:          len(unknownHost),
		ISP:                  "airtel",
		Country:              "India",
		Location:             "Bengaluru",
		EvilISP:              false,
	}

	if len(blocked) > 0 {
		scanStats.EvilISP = true
	}

	statsPSQLInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", "localhost", 5432, "postgres", "postgres", "scan_stats")
	scan_db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbn,
		DSN:  statsPSQLInfo,
	}), &gorm.Config{})
	scan_dbn, err := scan_db.DB()
	if err != nil {
		panic(err.Error())
	}
	scan_db.Create(&scanStats)
	defer scan_dbn.Close()
	// close channels
	close(urlsChan)
	close(resultsChan)

	if len(inaccessible) > 0 {
		log.Info().Msg("🚨 Evil ISP involved.")
		log.Info().Msgf("Websites :\nunique : %d\ninaccessible : %d\nblocked : %d\naccessible : %d\nunknownHost : %d\ntimedOut : %d\n", len(urls), len(inaccessible), len(blocked), len(accessible), len(unknownHost), len(timedOut))
		// fmt.Printf("Websites inaccessible : %v\n", inaccessible)
		// fmt.Printf("Websites blocked : %v\n", blocked)
		// fmt.Printf("Websites accessible : %v\n", accessible)
		// fmt.Printf("Websites unknown_host : %v\n", unknown_host)
	}

	log.Info().Msgf("\n%.2fs elapsed\n", time.Since(start).Seconds())
}
