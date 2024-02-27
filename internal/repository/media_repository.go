package repository

import (
	"downloader_gochat/model"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type IMediaRepository interface {
	SaveMediaData(mediaFile *model.MediaFile) error
}

type MediaRepository struct {
	db      *gorm.DB
	mongodb *mongo.Database
}

func NewMediaRepository(db *gorm.DB, mongodb *mongo.Database) *MediaRepository {
	return &MediaRepository{db: db, mongodb: mongodb}
}

//------------------------------------------
//------------------------------------------

func (m *MediaRepository) SaveMediaData(mediaFile *model.MediaFile) error {
	err := m.db.Create(mediaFile).Error
	return err
}
