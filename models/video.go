package models

import (
	"time"
	configs "zebra/configs/database"

	"github.com/jinzhu/gorm"
)

var db *gorm.DB

type Video struct {
	gorm.Model
	TempFilePath   string     `json:"temp_file_path"`
	TranscodedUrl  *string    `json:"transcoded_url"`
	TranscodedPath *string    `json:"transcoded_path"`
	Title          *string    `json:"title"`
	ProcessedAt    *time.Time `json:"processed_at"`
	Url            *string    `json:"url"`
	Failed         bool       `json:"failed"`
	Original_id    *string    `json:"original_id"`
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
	db := configs.GetDB() // Open the database connection
	db = db.Where("id = ?", id).Find(&getVideo)
	return &getVideo, db
}

func UpdateVideo(video Video) (*Video, error) {
	db.Save(&video)
	return &video, nil
}
