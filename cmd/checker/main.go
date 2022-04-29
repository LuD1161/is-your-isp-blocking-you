package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/LuD1161/is-your-isp-blocking-you/internal/helpers"
)

func main() {
	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts")
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	// Download zip file
	// http://s3-us-west-1.amazonaws.com/umbrella-static/top-1m.csv.zip
	top_1M_csv_zip := "http://s3-us-west-1.amazonaws.com/umbrella-static/top-1m.csv.zip"
	if helpers.DownloadPackage(top_1M_csv_zip, "/tmp/top-1m.csv.zip") != nil {
		return
	}
	helpers.Unzip("/tmp/top-1m.csv.zip")
	//Convert the body to type string
	sb := string(body)
	log.Print(sb)
}
