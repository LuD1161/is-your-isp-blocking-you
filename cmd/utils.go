package cmd

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	b64 "encoding/base64"
	"encoding/csv"
	"encoding/json"
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
	"golang.org/x/net/html"
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

func MakeRequest(urlsChan <-chan string, responseChan chan<- ValidatorData, customTransport *http.Transport) {
	retryClient := retryablehttp.NewClient()
	// http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	retryClient.HTTPClient.Transport = customTransport
	retryClient.HTTPClient.Timeout = time.Duration(timeout) * time.Second
	retryClient.RetryMax = MAX_RETRIES
	retryClient.Logger = nil // Don't want to log anything here

	client := retryClient.StandardClient() // *http.Client

	for url := range urlsChan {
		result := ValidatorData{
			URL:         url,
			Response:    http.Response{},
			Err:         nil,
			ResolvedIPs: "",
		}
		ResolvedIPs, err := GetIPs(url)
		// if no IPs are resolved then skip sending requests, retrying and waiting for timeout
		if err != nil {
			log.Error().Msgf("Error resolving IPs : %s", err.Error())
			result.Err = err
			responseChan <- result
			continue
		}
		result.ResolvedIPs = ResolvedIPs
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			result.Err = err
			log.Debug().Msgf("URL : %s | Error : %+v", url, err)
			responseChan <- result
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:101.0) Gecko/20100101 Firefox/101.0")
		response, err := client.Do(req)
		result.Err = err
		if err == nil {
			result.Response = *response
			log.Debug().Msgf("result.Response : %s", result.Response.Request.URL.String())
			log.Debug().Msgf("result.Response.StatusCode : %d", result.Response.StatusCode)
		} else {
			log.Debug().Msgf("result.Err : %s", result.Err.Error())
		}
		responseChan <- result
	}
}

func ValidateResponse(responseChan <-chan ValidatorData, resultsChan chan<- Result, validator Validator) {
	for response := range responseChan {
		result := Result{
			Code:           "",
			URL:            response.URL,
			Msg:            "",
			Data:           "",
			HTTPStatusCode: 0,
			HTMLTitle:      "",
			HTMLBodyLength: 0,
			Error:          response.Err,
			ResolvedIPs:    response.ResolvedIPs,
		}
		var body []byte
		err := response.Err
		if err != nil {
			if strings.Contains(err.Error(), "no such host") {
				result.Code = NO_SUCH_HOST
			} else if strings.Contains(err.Error(), "connect: network is unreachable") || strings.Contains(err.Error(), "Client.Timeout exceeded") {
				// giving up after 4 attempt(s): Get "http://blog.com": dial tcp 195.170.168.2:80: connect: network is unreachable
				// context deadline exceeded (Client.Timeout exceeded while awaiting headers)
				result.Code = CONN_TIMEOUT
			} else {
				// Running this through validator to evaluate PR_CONNECTION_RESET , whether it's actual SNI block or not
				code, msg, err := validator.Validate(response)
				if err != nil {
					log.Error().Msgf("Error in validating response : %s", err.Error())
				}
				result.Code = code
				result.Error = err
				result.Msg = msg
			}
		} else {
			result.URL = response.Response.Request.URL.String()
			body, err = ioutil.ReadAll(response.Response.Body)
			if err == nil {
				result.HTMLTitle = GetHTMLTitle(string(body))
				result.HTMLBodyLength = len(body)
				// if saveResponses is set, only then save response body to DB
				if saveResponses {
					// base64 encode body
					result.Data = b64.StdEncoding.EncodeToString(body)
				}
			} else { // err != nil
				log.Error().Msgf("URL : %s | Error in reading body : %s", result.URL, err.Error())
			}
			result.HTTPStatusCode = response.Response.StatusCode

			// Restore the io.ReadCloser to it's original state
			response.Response.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			code, msg, err := validator.Validate(response)
			if err != nil {
				log.Error().Msgf("Error in validating response : %s", err.Error())
			}
			result.Code = code
			result.Error = err
			result.Msg = msg
		}
		resultsChan <- result
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
	// Allow insecure, expired certificates
	proxyTransport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	if runThroughProxy {
		if proxyURL != "" {
			proxy, err := url.Parse(proxyURL)
			if err != nil {
				log.Error().Msgf("Error in parsing PROXY_URL from env vars, going with no proxy : %s", err.Error())
			}
			proxyTransport.Proxy = http.ProxyURL(proxy)
			log.Info().Msgf("Proxy set in http.transport : %s", proxyURL)
		} else {
			log.Error().Msgf("run_through_proxy set but No PROXY_URL specified. Hence, no proxy set in http.transport.")
		}
	} else {
		log.Info().Msgf("No PROXY_URL specified. Hence, no proxy set in http.transport.")
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

func GetHTMLTitle(body string) string {
	tkn := html.NewTokenizer(strings.NewReader(body))
	var isTitle bool
	for {
		tt := tkn.Next()
		switch {
		case tt == html.ErrorToken:
			return ""
		case tt == html.StartTagToken:
			t := tkn.Token()
			isTitle = t.Data == "title"
		case tt == html.TextToken:
			t := tkn.Token()
			if isTitle {
				log.Info().Msgf("Finding title : %s", t.Data)
				return t.Data
			}
		}
	}
}

func GetIPs(url string) (string, error) {
	url = strings.Replace(url, "http://", "", -1)
	url = strings.Replace(url, "https://", "", -1)
	var dataArr []string
	var data string
	ips, err := net.LookupIP(url)
	if err != nil {
		return data, err
	}
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			dataArr = append(dataArr, ipv4.String())
		}
	}
	data = strings.Join(dataArr, ",")
	return data, err
}
