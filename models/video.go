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
	FaleidMessage  *string    `json:"failed_message"`
}

func init() {
	configs.Connect()
	db = configs.GetDB()
	db.AutoMigrate(&Video{})
}

func (v *Video) Save() error {
	db := configs.GetDB()
	if err := db.Create(&v).Error; err != nil {
		return err
	}
	return nil
}

func GetVideoById(id string) (*Video, error) {
	var getVideo Video
	db := configs.GetDB()

	err := db.Where("id = ?", id).Find(&getVideo).Error
	if err != nil {
		return nil, err
	}
	return &getVideo, nil
}

func UpdateVideo(video Video) (*Video, error) {
	db := configs.GetDB()
	if err := db.Save(&video).Error; err != nil {
		return nil, err
	}
	return &video, nil
}
