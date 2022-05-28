package cmd

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog/log"
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

func MakeRequest(urlsChan <-chan string, resultsChan chan<- Result) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = MAX_RETRIES
	retryClient.Logger = nil // Don't want to log anything here

	client := retryClient.StandardClient() // *http.Client
	client.Timeout = time.Duration(timeout) * time.Second
	for url := range urlsChan {
		var result Result
		result.URL = url
		result.Error = nil
		resp, err := client.Get(url)
		if err != nil {
			result.Error = err
			result.Code = CONN_UNKNOWN
			if strings.Contains(err.Error(), "connection reset by peer") {
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

func InsertIntoDB(db *gorm.DB, records []Record) error {
	// db.Create
	// db.CreateInBatches(records, 1000)
	return nil
}

func GetISP() (IfConfigResponse, error) {
	client := &http.Client{}
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
