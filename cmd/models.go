package cmd

import (
	"net/http"

	"gorm.io/gorm"
)

type Record struct {
	gorm.Model
	ScanId         string  `gorm:"size:255;not null;" json:"scan_id"`
	Website        string  `gorm:"size:255;not null;" json:"website"`
	ISP            string  `gorm:"size:255;not null;" json:"isp"`
	Country        string  `gorm:"size:255;not null;" json:"country"`
	Location       string  `gorm:"size:255;" json:"location"`
	Latitude       float64 `gorm:"size:255;" json:"latitude"`
	Longitude      float64 `gorm:"size:255;" json:"longitude"`
	Accessible     bool    `gorm:"not null;" json:"accessible"`
	Data           string  `gorm:"type:text" json:"data"`
	ErrMsg         string  `gorm:"size:1024;" json:"err"`
	HTTPStatusCode int     `json:"http_status_code"`
	HTMLTitle      string  `gorm:"type:text" json:"html_title"`
	HTMLBodyLength int     `json:"html_body_length"`
	FilteringType  string  `gorm:"size:50;" json:"filtering_type"`
	ValidatorMsg   string  `gorm:"type:text" json:"validator_message"`
}

type Result struct {
	Code           string
	URL            string
	Msg            string // Any extra message. Like : what blocked string in HTTP Filtering
	Data           string // base64 encoded response body; default disabled
	HTTPStatusCode int
	HTMLTitle      string
	HTMLBodyLength int
	Error          error
}

type ScanStats struct {
	gorm.Model
	ScanId               string  `gorm:"size:255;not null;" json:"scan_id"`
	ScanTime             int     `gorm:"column:scan_time;int;not null;" json:"scan_time"`
	UniqueDomainsScanned int     `gorm:"int;not null;" json:"unique_domains_scanned"`
	Accessible           int     `gorm:"int;not null;" json:"accessible"`
	ConnectionReset      int     `gorm:"int;not null;" json:"connection_reset"`
	Inaccessible         int     `gorm:"int;not null;" json:"inaccessible"`
	Blocked              int     `gorm:"int;not null;" json:"blocked"`
	HTTPFiltered         int     `gorm:"int;not null;" json:"http_filtered"`
	DNSFiltered          int     `gorm:"int;not null;" json:"dns_filtered"`
	SNIFiltered          int     `gorm:"int;not null;" json:"sni_filtered"`
	TimedOut             int     `gorm:"int;not null;" json:"timed_out"`
	UnknownHost          int     `gorm:"int;not null;" json:"unknown_host"`
	ISP                  string  `gorm:"size:255;not null;" json:"isp"`
	Country              string  `gorm:"size:255;not null;" json:"country"`
	Location             string  `json:"location"`
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	DomainList           string  `gorm:"size:255;not null;" json:"domain_list"` // Filepath that was used to scan
	EvilISP              bool    `gorm:"not null;" json:"evil_isp"`
}

type IfConfigResponse struct {
	IP         string  `json:"ip"`
	Country    string  `json:"country"`
	CountryISO string  `json:"country_iso"`
	RegionName string  `json:"region_name"`
	ZipCode    string  `json:"zip_code"`
	City       string  `json:"city"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Asn        string  `json:"asn"`
	AsnOrg     string  `json:"asn_org"`
}

type ValidatorData struct {
	URL      string
	Response http.Response
	Err      error
}

type FilterObj struct {
	Value     string `yaml:"Value"`
	Country   string `yaml:"country"`
	ISP       string `yaml:"ISP"`
	ASN       string `yaml:"ASN"`
	Reference string `yaml:"Reference"`
	Comment   string `yaml:"Comment"`
}
type FilteringYAML struct {
	DNSFILTERING struct {
		CNAME []FilterObj `yaml:"CNAME"`
		IP    []FilterObj `yaml:"IP"`
	} `yaml:"DNS_FILTERING"`
	HTTPFILTERING struct {
		Body []FilterObj `yaml:"Body"`
		URL  []FilterObj `yaml:"URL"`
	} `yaml:"HTTP_FILTERING"`
	SNIFILTERING struct {
		Error []FilterObj `yaml:"Error"`
	} `yaml:"SNI_FILTERING"`
}
