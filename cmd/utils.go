package cmd

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// https://gist.github.com/ptrelford/3d132b9169e2cde21181
func DownloadPackage(url, dstFilePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, _ := os.Create(dstFilePath)
	defer out.Close()
	io.Copy(out, resp.Body)
	return nil
}

func Unzip(filePath string) error {
	reader, _ := zip.OpenReader(filePath)
	defer reader.Close()
	for _, file := range reader.File {
		in, err := file.Open()
		if err != nil {
			return err
		}
		defer in.Close()
		relname := path.Join("/tmp", file.Name)
		dir := path.Dir(relname)
		os.MkdirAll(dir, 0777)
		out, err := os.Create(relname)
		if err != nil {
			return err
		}
		defer out.Close()
		io.Copy(out, in)
	}
	return nil
}

func MakeRequest(urlsChan <-chan string, resultsChan chan<- Result, customTransport *http.Transport) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = MAX_RETRIES
	retryClient.Logger = nil // Don't want to log anything here

	client := retryClient.StandardClient() // *http.Client
	client.Timeout = time.Duration(timeout) * time.Second
	client.Transport = customTransport
	for url := range urlsChan {
		var result Result
		result.URL = url
		result.Error = nil
		resp, err := client.Get(url)
		if err != nil {
			result.Error = err
			result.Code = CONN_UNKNOWN
			if strings.Contains(err.Error(), "connection reset by peer") {
				log.Debug().Msgf("URL : %s | Error : %+v", url, err)
				result.Code = CONN_RESET
			}
			if strings.Contains(err.Error(), "no such hostError") {
				result.Code = NO_SUCH_HOST
			}
			if err, ok := err.(net.Error); ok && err.Timeout() {
				result.Code = CONN_TIMEOUT
			}
			resultsChan <- result
			continue
		}
		finalURL := resp.Request.URL.String()
		if resp.StatusCode == 200 {
			result.URL = finalURL
			// check body message or length
			// Sometimes the website loads as "200 OK" but contains "(b|B)locked" as response text
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				result.Code = CONN_UNKNOWN
				result.Error = err
				resultsChan <- result
				continue
			}
			resp.Body.Close()
			if strings.Contains("blocked", string(body)) || len(body) < 600 {
				result.Code = CONN_BLOCKED
				resultsChan <- result
				continue
			}
			log.Debug().Msgf("URL : %s | finalURL : %s | len(body) : %d\n", url, finalURL, len(body))
			result.Code = CONN_OK
			resultsChan <- result
			continue
		} else {
			log.Debug().Msgf("URL : %s | resp.StatusCode : %d", url, resp.StatusCode)
			result.Code = CONN_UNKNOWN
			resultsChan <- result
			continue
		}
	}
}

func ReadCsvFile(filePath string) ([][]string, error) {
	err := error(nil)
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal().Msgf("Unable to read input file %s : %s", filePath, err.Error())
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	// Don't read all at once, memory safe
	// records, err := csvReader.ReadAll()
	rows := make([][]string, 0)
	for {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return rows, err
		}
		rows = append(rows, row)
	}
}

func GetISP(customTransport *http.Transport) (IfConfigResponse, error) {
	client := &http.Client{Transport: customTransport}
	req, err := http.NewRequest("GET", "https://ifconfig.co/json", nil)
	if err != nil {
		log.Error().Msgf("Error : %s ", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Msgf("Error : %s ", err.Error())
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error().Msgf("Error : %s ", err.Error())
	}
	var result IfConfigResponse
	err = json.Unmarshal(bodyText, &result)
	return result, err
}

func SetProxyTransport() *http.Transport {
	// create proxy transport
	proxyURL := os.Getenv("PROXY_URL")
	proxyTransport := &http.Transport{}
	if runThroughProxy {
		if proxyURL != "" {
			proxy, err := url.Parse(proxyURL)
			if err != nil {
				log.Error().Msgf("Error in parsing PROXY_URL from env vars, going with no proxy : %s", err.Error())
			}
			proxyTransport.Proxy = http.ProxyURL(proxy)
			log.Info().Msgf("Proxy set in http.transport : %s", proxyURL)
		}
	} else {
		log.Info().Msgf("No PROXY_URL specified. Hence, no proxy set in http.transport created.")
	}
	return proxyTransport
}

func GenerateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func saveInDB(results []Record, scanStats ScanStats) error {
	switch storeInDB {
	case "postgres":
		postgresDSN := os.Getenv("POSTGRES_DSN")
		if postgresDSN == "" {
			return errors.New("POSTGRES_DSN not specified")
		}
		db, err := gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
		if err != nil {
			log.Error().Msgf("Can't open DB connection : %s", err.Error())
			panic(err.Error())
		}
		dbn, err := db.DB()
		if err != nil {
			panic(err.Error())
		}
		defer dbn.Close()
		db.AutoMigrate(Record{}, ScanStats{})
		if err := db.CreateInBatches(results, 1000).Error; err != nil {
			log.Error().Stack().Err(err).Msgf("Error saving results in DB [CreateInBatches] : %s", err.Error())
			return err
		}
		// Create scan stats
		db.Create(&scanStats)
	case "sqlite":
	case "mysql":
	default:
		if storeInDB != "" {
			fmt.Println("sqlite & mysql are WIP. Please try with 'postgres'")
		}

	}
	return nil
}
