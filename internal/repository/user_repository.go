package repository

import (
	"downloader_gochat/configs"
	"downloader_gochat/model"
	"errors"
	"time"

	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IUserRepository interface {
	AddUser(user *model.User) (*model.User, error)
	GetDetailUser(int64) (*model.UserWithImageDataModel, error)
	GetUserByUsernameEmail(username string, email string) (*model.UserDataModel, error)
	UpdateUser(*model.User) (*model.User, error)
	AddSession(device *model.DeviceInfo, deviceId string, userId int64, refreshToken string, ipLocation string) error
	UpdateSession(device *model.DeviceInfo, deviceId string, userId int64, refreshToken string, ipLocation string) (bool, error)
	UpdateSessionRefreshToken(device *model.DeviceInfo, userId int64, refreshToken string, prevRefreshToken string, ipLocation string) (*model.ActiveSession, error)
	GetUserActiveSessions(userId int64) ([]model.ActiveSession, error)
	RemoveSession(userId int64, prevRefreshToken string) error
	RemoveSessions(userId int64, prevRefreshTokens []string) error
	AddUserFollow(userId int64, followId int64) error
	RemoveUserFollow(userId int64, followId int64) error
	GetUserFollowers(userId int64, skip int, limit int) ([]model.FollowUserDataModel, error)
	GetUserFollowings(userId int64, skip int, limit int) ([]model.FollowUserDataModel, error)
	GetUserMetaDataAndNotificationSettings(id int64, imageLimit int) (*model.UserMetaWithNotificationSettings, error)
}

type UserRepository struct {
	db      *gorm.DB
	mongodb *mongo.Database
}

func NewUserRepository(db *gorm.DB, mongodb *mongo.Database) *UserRepository {
	return &UserRepository{db: db, mongodb: mongodb}
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) AddUser(user *model.User) (*model.User, error) {
	//todo : need optimization
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// do some database operations in the transaction (use 'tx' from this point, not 'db')
		if err := tx.Create(&user).Error; err != nil {
			// return any error will rollback
			return err
		}

		movieSettings := model.MovieSettings{
			UserId:        user.UserId,
			IncludeAnime:  true,
			IncludeHentai: false,
		}
		if err := tx.Create(&movieSettings).Error; err != nil {
			return err
		}

		downloadLinksSettings := model.DownloadLinksSettings{
			UserId:             user.UserId,
			IncludeCensored:    true,
			IncludeDubbed:      true,
			IncludeHardSub:     true,
			PreferredQualities: pq.StringArray{"720p", "1080p", "2160p"},
		}

		if err := tx.Create(&downloadLinksSettings).Error; err != nil {
			return err
		}

		notificationSettings := model.NotificationSettings{
			UserId:                    user.UserId,
			FinishedListSpinOffSequel: true,
			FollowMovie:               true,
			FollowMovieBetterQuality:  true,
			FollowMovieSubtitle:       true,
			FutureList:                true,
			FutureListSerialSeasonEnd: true,
			FutureListSubtitle:        true,
			NewFollower:               true,
			NewMessage:                false,
		}
		if err := tx.Create(&notificationSettings).Error; err != nil {
			return err
		}

		userMessageRead := model.UserMessageRead{
			UserId: user.UserId,
		}
		if err := tx.Create(&userMessageRead).Error; err != nil {
			return err
		}

		// return nil will commit the whole transaction
		return nil
	})

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetDetailUser(id int64) (*model.UserWithImageDataModel, error) {
	var userDataModel model.UserWithImageDataModel
	err := r.db.Where("\"userId\" = ?", id).
		Model(&model.User{}).
		Limit(1).
		Preload("ProfileImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"addDate\" DESC")
		}).
		Find(&userDataModel).
		Error
	if err != nil {
		return nil, err
	}

	return &userDataModel, nil
}

func (r *UserRepository) GetUserByUsernameEmail(username string, email string) (*model.UserDataModel, error) {
	var userDataModel model.UserDataModel
	err := r.db.Where("(username != '' AND username = ?) OR (email != '' AND email = ?)", username, email).Model(&model.User{}).Limit(1).Find(&userDataModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if userDataModel.UserId == 0 {
		return nil, nil
	}
	return &userDataModel, nil
}

func (r *UserRepository) UpdateUser(user *model.User) (*model.User, error) {
	err := r.db.Save(&user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) AddSession(device *model.DeviceInfo, deviceId string, userId int64, refreshToken string, ipLocation string) error {
	newDevice := model.ActiveSession{
		UserId:       userId,
		RefreshToken: refreshToken,
		DeviceId:     deviceId,
		AppName:      device.AppName,
		AppVersion:   device.AppVersion,
		DeviceModel:  device.DeviceModel,
		DeviceOs:     device.Os,
		IpLocation:   ipLocation,
	}
	err := r.db.Create(&newDevice).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) UpdateSession(device *model.DeviceInfo, deviceId string, userId int64, refreshToken string, ipLocation string) (bool, error) {
	now := time.Now().UTC()
	newDevice := model.ActiveSession{
		UserId:       userId,
		RefreshToken: refreshToken,
		DeviceId:     deviceId,
		AppName:      device.AppName,
		AppVersion:   device.AppVersion,
		DeviceModel:  device.DeviceModel,
		DeviceOs:     device.Os,
		IpLocation:   ipLocation,
		LoginDate:    now,
	}

	err := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "userId"}, {Name: "deviceId"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"appName", "appVersion", "deviceOs", "deviceModel", "ipLocation", "lastUseDate", "refreshToken"}),
	}).Create(&newDevice).Error

	if err != nil {
		return false, err
	}
	isNewDevice := newDevice.LoginDate.UnixMilli()-now.UnixMilli() < 3000
	return isNewDevice, nil
}

func (r *UserRepository) UpdateSessionRefreshToken(device *model.DeviceInfo, userId int64, refreshToken string, prevRefreshToken string, ipLocation string) (*model.ActiveSession, error) {
	activeSession := model.ActiveSession{}
	result := r.db.Model(&activeSession).Clauses(clause.Returning{}).Where("\"userId\" = ? AND \"refreshToken\" = ?", userId, prevRefreshToken).Updates(map[string]interface{}{
		"refreshToken": refreshToken,
		"appName":      device.AppName,
		"appVersion":   device.AppVersion,
		"deviceModel":  device.DeviceModel,
		"deviceOs":     device.Os,
		"ipLocation":   ipLocation,
		"lastUseDate":  time.Now().UTC(),
	})

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &activeSession, nil
}

func (r *UserRepository) GetUserActiveSessions(userId int64) ([]model.ActiveSession, error) {
	var activeSessions []model.ActiveSession
	err := r.db.Model(&model.ActiveSession{}).Where("\"userId\" = ?", userId).Order("\"lastUseDate\" desc").Limit(2 * configs.GetConfigs().ActiveSessionsLimit).Find(&activeSessions).Error
	if err != nil {
		return nil, err
	}
	return activeSessions, nil
}

func (r *UserRepository) RemoveSession(userId int64, prevRefreshToken string) error {
	err := r.db.Where("\"userId\" = ? AND \"refreshToken\" = ?", userId, prevRefreshToken).Delete(&model.ActiveSession{}).Error
	return err
}

func (r *UserRepository) RemoveSessions(userId int64, prevRefreshTokens []string) error {
	err := r.db.Where("\"userId\" = ? AND \"refreshToken\" IN ?", userId, prevRefreshTokens).Delete(&model.ActiveSession{}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return nil
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) AddUserFollow(userId int64, followId int64) error {
	follow := model.Follow{
		FollowerId:  userId,
		FollowingId: followId,
		AddDate:     time.Now().UTC(),
	}

	err := r.db.Create(&follow).Error
	return err
}

func (r *UserRepository) RemoveUserFollow(userId int64, followId int64) error {
	err := r.db.Debug().Where("\"followerId\" = ? AND \"followingId\" = ?", userId, followId).Delete(&model.Follow{}).Error
	return err
}

func (r *UserRepository) GetUserFollowers(userId int64, skip int, limit int) ([]model.FollowUserDataModel, error) {
	var result []model.FollowUserDataModel
	err := r.db.Debug().Model(&model.User{}).Joins("join \"Follow\" on \"userId\" = \"followerId\" AND \"followingId\" = ? ", userId).
		Order("\"addDate\" desc").
		Offset(skip).
		Limit(limit).
		Preload("ProfileImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"addDate\" DESC")
		}).
		Find(&result).Error

	return result, err
}

func (r *UserRepository) GetUserFollowings(userId int64, skip int, limit int) ([]model.FollowUserDataModel, error) {
	var result []model.FollowUserDataModel
	err := r.db.Model(&model.User{}).Joins("join \"Follow\" on \"userId\" = \"followingId\" AND \"followerId\" = ? ", userId).
		Order("\"addDate\" desc").
		Offset(skip).
		Limit(limit).
		Preload("ProfileImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"addDate\" DESC")
		}).
		Find(&result).Error

	return result, err
}

func (r *UserRepository) GetUserMetaDataAndNotificationSettings(id int64, imageLimit int) (*model.UserMetaWithNotificationSettings, error) {
	var result model.UserMetaWithNotificationSettings
	err := r.db.
		Model(&model.User{}).
		Select("\"User\".*, \"NotificationSettings\".*").
		Where("\"User\".\"userId\" = ?", id).
		Limit(1).
		Joins("JOIN \"NotificationSettings\" ON \"User\".\"userId\" = \"NotificationSettings\".\"userId\" ").
		Preload("ProfileImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"addDate\" DESC").Limit(imageLimit)
		}).
		Find(&result).
		Error

	if err != nil {
		return nil, err
	}
	return &result, nil
}
