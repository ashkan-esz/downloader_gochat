package db

import (
	"downloader_gochat/configs"
	"downloader_gochat/model"
	errorHandler "downloader_gochat/pkg/error"
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db *gorm.DB
}

func NewDatabase() (*Database, error) {
	db, err := gorm.Open(
		postgres.Open(configs.GetConfigs().DbUrl),
		&gorm.Config{
			SkipDefaultTransaction: true,
			PrepareStmt:            true,
			TranslateError:         true,
		},
	)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	return &Database{db: db}, nil
}

func (d *Database) AutoMigrate() {
	if !configs.GetConfigs().MigrateOnStart {
		return
	}

	////err := d.db.Exec("DROP TYPE IF EXISTS \"titleRelation\"").Error
	err := d.db.Exec("create type \"titleRelation\" as enum ('prequel', 'sequel', 'spin_off', 'side_story', 'full_story', 'summary', 'parent_story', 'other', 'alternative_setting', 'alternative_version');").Error
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		errorMessage := fmt.Sprintf("error on AutoMigrate: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}

	////err = d.db.Exec("DROP TYPE IF EXISTS \"userRole\"").Error
	err = d.db.Exec("create type \"userRole\" as enum ('test_user', 'user', 'dev', 'admin');").Error
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		errorMessage := fmt.Sprintf("error on AutoMigrate: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}

	////err = d.db.Exec("DROP TYPE IF EXISTS \"likeDislike\"").Error
	//err = d.db.Exec("create type \"likeDislike\" as enum ('like', 'dislike');").Error
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		errorMessage := fmt.Sprintf("error on AutoMigrate: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}

	////err = d.db.Exec("DROP TYPE IF EXISTS \"MbtiType\"").Error
	err = d.db.Exec("create type \"MbtiType\" as enum ('ISTJ', 'ISFJ', 'INFJ', 'INTJ', 'ISTP', 'ISFP', 'INFP', 'INTP', 'ESTP', 'ESFP', 'ENFP', 'ENTP', 'ESTJ', 'ESFJ', 'ENFJ', 'ENTJ');\n").Error
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		errorMessage := fmt.Sprintf("error on AutoMigrate: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}

	d.db.Exec("DROP TABLE IF EXISTS \"ProfileImage\"")
	d.db.Exec("DROP TABLE IF EXISTS \"ActiveSession\"")
	d.db.Exec("DROP TABLE IF EXISTS \"ComputedFavoriteGenres\"")
	d.db.Exec("DROP TABLE IF EXISTS \"MovieSettings\"")
	d.db.Exec("DROP TABLE IF EXISTS \"NotificationSettings\"")
	d.db.Exec("DROP TABLE IF EXISTS \"CastImage\"")
	d.db.Exec("DROP TABLE IF EXISTS \"Credit\"")
	err = d.db.AutoMigrate(
		&model.User{},
		&model.Movie{}, &model.RelatedMovie{},
		&model.Follow{},
		&model.ProfileImage{},
		&model.ActiveSession{},
		&model.ComputedFavoriteGenres{},
		&model.DownloadLinksSettings{}, &model.MovieSettings{}, &model.NotificationSettings{},
		&model.Staff{}, &model.Character{},
		&model.Credit{},
		&model.FavoriteCharacter{}, &model.LikeDislikeCharacter{},
		&model.FollowStaff{}, &model.LikeDislikeStaff{}, &model.CastImage{},
		&model.FollowMovie{}, &model.LikeDislikeMovie{}, &model.WatchedMovie{},
		&model.WatchListGroup{}, &model.WatchListMovie{},
		&model.UserCollection{}, &model.UserCollectionMovie{},
		&model.Room{}, &model.Message{}, &model.UserMessageRead{}, &model.MediaFile{},
		&model.Bot{}, &model.UserBot{},
	)
	if err != nil {
		errorMessage := fmt.Sprintf("error on AutoMigrate: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}

	err = d.db.Model(&model.NotificationEntityType{}).CreateInBatches(model.NotificationEntityTypesAndId, 10).Error
	if err != nil {
		errorMessage := fmt.Sprintf("error on Inserting Notification entity types: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}
}

func (d *Database) Close() {
	// try not to use it due to gorm connection pooling
	sqlDB, err := d.db.DB()
	if err != nil {
		log.Fatalln(err)
	}
	sqlDB.Close()
}

func (d *Database) GetDB() *gorm.DB {
	return d.db
}
