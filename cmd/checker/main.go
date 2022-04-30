package main

import (
	"fmt"
	"log"
	"time"

	"github.com/LuD1161/is-your-isp-blocking-you/internal/helpers"
)

func main() {
	// Download zip file
	// http://s3-us-west-1.amazonaws.com/umbrella-static/top-1m.csv.zip
	top_1M_csv_zip := "http://s3-us-west-1.amazonaws.com/umbrella-static/top-1m.csv.zip"
	if helpers.DownloadPackage(top_1M_csv_zip, "/tmp/top-1m.csv.zip") != nil {
		return
	}
	helpers.Unzip("/tmp/top-1m.csv.zip")

	start := time.Now()

	records, err := helpers.ReadCsvFile("/tmp/top-1m.csv")
	if err != nil {
		log.Fatalln("Error in reading csv file")
	}

	ch := make(chan string)
	for i := 0; i < len(records); i++ {
		fmt.Printf("printing record : %v\n", records[i])
		url := fmt.Sprintf("https://%s", records[i][1])
		fmt.Printf("url : %v\n", records[i][1])
		go helpers.MakeRequest(url, ch)
	}

	for i := 0; i < len(records); i++ {
		fmt.Println(<-ch)
	}

	fmt.Printf("\n%.2fs elapsed\n", time.Since(start).Seconds())
}
