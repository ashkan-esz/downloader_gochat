package repository

import (
	"downloader_gochat/model"

	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type ICastRepository interface {
	SaveStaffCastImageBlurHash(staffId int64, blurHash string) error
	SaveCharacterCastImageBlurHash(characterId int64, blurHash string) error
}

type CastRepository struct {
	db      *gorm.DB
	mongodb *mongo.Database
}

func NewCastRepository(db *gorm.DB, mongodb *mongo.Database) *CastRepository {
	return &CastRepository{db: db, mongodb: mongodb}
}

//------------------------------------------
//------------------------------------------

func (c *CastRepository) SaveStaffCastImageBlurHash(staffId int64, blurHash string) error {
	err := c.db.
		Model(&model.CastImage{}).
		Where("\"staffId\" = ?", staffId).
		UpdateColumn("\"blurHash\"", blurHash).
		Error
	return err
}

func (c *CastRepository) SaveCharacterCastImageBlurHash(characterId int64, blurHash string) error {
	err := c.db.
		Model(&model.CastImage{}).
		Where("\"characterId\" = ?", characterId).
		UpdateColumn("\"blurHash\"", blurHash).
		Error
	return err
}
