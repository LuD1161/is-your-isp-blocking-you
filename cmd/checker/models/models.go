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
	Err_msg    string `gorm:"size:255;" json:"err"`
}

type Result struct {
	Code  int
	URL   string
	Error error
}
