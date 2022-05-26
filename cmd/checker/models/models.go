package models

import (
	"gorm.io/gorm"
)

type Record struct {
	gorm.Model
	Website    string `gorm:"size:255;not null;" json:"website"`
	ISP        string `gorm:"size:255;not null;" json:"isp"`
	Country    string `gorm:"size:255;not null;" json:"country"`
	Location   string `gorm:"size:255;not null;" json:"location"`
	Accessible bool   `gorm:"not null;" json:"accessible"`
	ErrMsg     string `gorm:"size:1024;" json:"err"`
}

type Result struct {
	Code  int
	URL   string
	Error error
}

type ScanStats struct {
	gorm.Model
	ScanTime             int    `gorm:"column:scan_time;int;not null;" json:"scan_time"`
	UniqueDomainsScanned int    `gorm:"int;not null;" json:"unique_domains_scanned"`
	Accessible           int    `gorm:"int;not null;" json:"accessible"`
	Inaccessible         int    `gorm:"int;not null;" json:"inaccessible"`
	Blocked              int    `gorm:"int;not null;" json:"blocked"`
	TimedOut             int    `gorm:"int;not null;" json:"timed_out"`
	UnknownHost          int    `gorm:"int;not null;" json:"unknown_host"`
	ISP                  string `gorm:"size:255;not null;" json:"isp"`
	Country              string `gorm:"size:255;not null;" json:"country"`
	Location             string `gorm:"size:255;not null;" json:"location"`
	EvilISP              bool   `gorm:"not null;" json:"evil_isp"`
}
