package models

import (
	"time"
	configs "zebra/configs/database"

	"github.com/jinzhu/gorm"
)

var db *gorm.DB

type Video struct {
	gorm.Model
	OriginalFilePath string     `json:"original_file_path"`
	FilePath         string     `json:"file_path"`
	Title            *string    `json:"title"`
	ProcessedAt      *time.Time `json:"processed_at"`
	Url              string     `json:"url"`
	Failed           bool       `json:"failed"`
}

func init() {
	configs.Connect()
	db = configs.GetDB()
	db.AutoMigrate(&Video{})
}

func (v *Video) Save() error {
	if err := db.Create(&v).Error; err != nil {
		return err
	}
	return nil
}

func GetVideoById(id string) (*Video, *gorm.DB) {
	var getVideo Video
	db := db.Where("id = ?", id).Find(&getVideo)
	return &getVideo, db
}
