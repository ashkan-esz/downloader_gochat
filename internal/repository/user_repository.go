package repository

import (
	"downloader_gochat/configs"
	"downloader_gochat/model"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IUserRepository interface {
	AddUser(user *model.User, roleId int64) (*model.User, error)
	GetDetailUser(int64) (*model.UserWithImageDataModel, error)
	GetUserByUsernameEmail(username string, email string) (*model.UserDataModel, error)
	GetUserByUsernameEmailAndUserId(userId int64, username string, email string) (*model.UserDataModel, error)
	GetUserProfile(requestParams *model.UserProfileReq) (*model.UserProfileRes, error)
	GetUserMetaData(id int64) (*model.UserDataModel, error)
	EditUserProfile(userId int64, editFields *model.EditProfileReq, updateFields map[string]interface{}) error
	UpdateUserPassword(userId int64, passwords *model.UpdatePasswordReq) error
	SaveUserEmailToken(userId int64, token string, expire int64) error
	SaveDeleteAccountToken(userId int64, token string, expire int64) error
	VerifyUserEmailToken(userId int64, token string) error
	VerifyDeleteAccountToken(userId int64, token string) error
	DeleteUserAndRelatedData(userId int64) error
	GetProfileImagesCount(userId int64) (int64, error)
	GetProfileImages(userId int64) (*[]model.ProfileImageDataModel, error)
	SaveProfileImageData(profileImage *model.ProfileImage) error
	RemoveProfileImageData(userId int64, fileName string) error
	RemoveAllProfileImageData(userId int64) ([]model.ProfileImage, error)
	AddSession(device *model.DeviceInfo, deviceId string, userId int64, refreshToken string, ipLocation string) error
	UpdateSession(device *model.DeviceInfo, deviceId string, userId int64, refreshToken string, ipLocation string) (bool, error)
	UpdateSessionRefreshToken(device *model.DeviceInfo, userId int64, refreshToken string, prevRefreshToken string, ipLocation string) (*model.ActiveSession, error)
	UpdateSessionNotifToken(userId int64, refreshToken string, notifToken string) error
	GetUserActiveSessions(userId int64) ([]model.ActiveSession, error)
	RemoveSession(userId int64, prevRefreshToken string) error
	RemoveSessions(userId int64, prevRefreshTokens []string) error
	RemoveAuthSession(userId int64, refreshToken string, deviceId string) (*model.ActiveSession, error)
	RemoveAllAuthSession(userId int64, refreshToken string) ([]model.ActiveSession, error)
	AddUserFollow(userId int64, followId int64) error
	RemoveUserFollow(userId int64, followId int64) error
	GetUserFollowers(userId int64, skip int, limit int) ([]model.FollowUserDataModel, error)
	GetUserFollowings(userId int64, skip int, limit int) ([]model.FollowUserDataModel, error)
	GetUserMetaDataAndNotificationSettings(id int64, imageLimit int) (*model.UserMetaWithNotificationSettings, error)
	GetUserDownloadLinkSettings(userId int64) (*model.DownloadLinksSettings, error)
	GetUserNotificationSettings(userId int64) (*model.NotificationSettings, error)
	GetUserMovieSettings(userId int64) (*model.MovieSettings, error)
	UpdateUserDownloadLinkSettings(userId int64, settings model.DownloadLinksSettings) error
	UpdateUserNotificationSettings(userId int64, settings model.NotificationSettings) error
	UpdateUserMovieSettings(userId int64, settings model.MovieSettings) error
	UpdateUserFavoriteGenres(userId int64, genresArray []string) error
	GetActiveSessions(userId int64) ([]model.ActiveSessionDataModel, error)
	GetUserBots(userId int64) ([]model.UserBotDataModel, error)
	GetBotData(botId string) (*model.Bot, error)
	GetUserRoles(userId int64) ([]model.Role, error)
	GetUserRolesWithPermissions(userId int64) ([]model.RoleWithPermissions, error)
	GetUserPermissionsByRoleIds(roleIds []int64) ([]model.Permission, error)
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

func (r *UserRepository) AddUser(user *model.User, roleId int64) (*model.User, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// do some database operations in the transaction (use 'tx' from this point, not 'db')
		if err := tx.Create(&user).Error; err != nil {
			// return any error will roll back
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
			FollowMovieBetterQuality:  false,
			FollowMovieSubtitle:       false,
			FutureList:                false,
			FutureListSerialSeasonEnd: true,
			FutureListSubtitle:        false,
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

		userRole := model.UserToRole{
			UserId: user.UserId,
			RoleId: roleId,
		}
		if err := tx.Create(&userRole).Error; err != nil {
			return err
		}

		userTorrent := model.UserTorrent{
			UserId:         user.UserId,
			TorrentLeachGb: 0,
			TorrentSearch:  0,
			FirstUseAt:     time.Now(),
		}
		if err := tx.Create(&userTorrent).Error; err != nil {
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

func (r *UserRepository) GetUserByUsernameEmailAndUserId(userId int64, username string, email string) (*model.UserDataModel, error) {
	var userDataModel model.UserDataModel
	err := r.db.Where("\"userId\" != ? AND ((username != '' AND username = ?) OR (email != '' AND email = ?))", userId, username, email).Model(&model.User{}).Limit(1).Find(&userDataModel).Error
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

func (r *UserRepository) GetUserProfile(requestParams *model.UserProfileReq) (*model.UserProfileRes, error) {
	var result model.UserProfileRes

	query := r.db.Model(&model.User{}).
		Where("\"userId\" = ?", requestParams.UserId).
		Limit(1)

	if requestParams.LoadProfileImages {
		query = query.Preload("ProfileImages", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"addDate\" DESC")
		})
	}

	if requestParams.IsSelfProfile {
		if requestParams.LoadComputedFavoriteGenres {
			query = query.Preload("ComputedFavoriteGenres", func(db *gorm.DB) *gorm.DB {
				return db.Order("\"percent\" DESC")
			})
		}

		if requestParams.LoadSettings {
			query = query.Preload("NotificationSettings").
				Preload("DownloadLinksSettings").
				Preload("MovieSettings")
		}
		if requestParams.RefreshToken != "" {
			query = query.Preload("ThisDevice", "\"refreshToken\" = ?", requestParams.RefreshToken)
		}
	}

	if requestParams.LoadTorrentUsage {
		query = query.Preload("UserTorrent")
	}

	err := query.Find(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if result.UserId == 0 {
		return nil, nil
	}

	if requestParams.LoadRolesWithPermissions {
		result.RolesWithPermissions, err = r.GetUserRolesWithPermissions(requestParams.UserId)
		if err != nil {
			return nil, err
		}
	} else if requestParams.LoadRoles {
		result.Roles, err = r.GetUserRoles(requestParams.UserId)
		if err != nil {
			return nil, err
		}
	}

	if requestParams.LoadFollowersCount {
		type countRes struct {
			FollowersCount  int64 `gorm:"column:f1;"`
			FollowingsCount int64 `gorm:"column:f2;"`
		}
		var res countRes

		err := r.db.
			Raw("SELECT count(1) filter ( where \"followingId\" = ? ) as f1, count(1) filter ( where \"followerId\" = ? ) as f2 FROM \"Follow\"",
				requestParams.UserId, requestParams.UserId).
			Scan(&res).
			Error

		if err == nil {
			result.FollowersCount = res.FollowersCount
			result.FollowingsCount = res.FollowingsCount
		}
	}
	if !requestParams.IsSelfProfile {
		result.ComputedStatsLastUpdate = 0
		result.Email = ""
	}

	return &result, nil
}

func (r *UserRepository) GetUserMetaData(id int64) (*model.UserDataModel, error) {
	var userDataModel model.UserDataModel
	err := r.db.Where("\"userId\" = ?", id).
		Model(&model.User{}).
		Limit(1).
		Find(&userDataModel).
		Error
	if err != nil {
		return nil, err
	}

	return &userDataModel, nil
}

func (r *UserRepository) EditUserProfile(userId int64, editFields *model.EditProfileReq, updateFields map[string]interface{}) error {
	res := r.db.
		Model(&model.User{}).
		Where("\"userId\" = ?", userId).
		Limit(1).
		Updates(updateFields)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepository) UpdateUserPassword(userId int64, passwords *model.UpdatePasswordReq) error {
	res := r.db.
		Model(&model.User{}).
		Where("\"userId\" = ?", userId).
		Limit(1).
		UpdateColumn("password", passwords.NewPassword)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) SaveUserEmailToken(userId int64, token string, expire int64) error {
	res := r.db.
		Model(&model.User{}).
		Where("\"userId\" = ?", userId).
		Limit(1).
		Updates(map[string]interface{}{
			"emailVerifyToken":        token,
			"emailVerifyToken_expire": expire,
		})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepository) SaveDeleteAccountToken(userId int64, token string, expire int64) error {
	res := r.db.
		Model(&model.User{}).
		Where("\"userId\" = ?", userId).
		Limit(1).
		Updates(map[string]interface{}{
			"deleteAccountVerifyToken":        token,
			"deleteAccountVerifyToken_expire": expire,
		})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepository) VerifyUserEmailToken(userId int64, token string) error {
	res := r.db.
		Model(&model.User{}).
		Where("\"userId\" = ? AND \"emailVerifyToken\" = ? AND \"emailVerifyToken_expire\" >= ? ",
			userId, token, time.Now().UnixMilli()).
		Limit(1).
		Updates(map[string]interface{}{
			"emailVerifyToken":        "",
			"emailVerifyToken_expire": 0,
			"emailVerified":           true,
		})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepository) VerifyDeleteAccountToken(userId int64, token string) error {
	res := r.db.
		Model(&model.User{}).
		Where("\"userId\" = ? AND \"deleteAccountVerifyToken\" = ? AND \"deleteAccountVerifyToken_expire\" >= ? ",
			userId, token, time.Now().UnixMilli()).
		Limit(1).
		Updates(map[string]interface{}{
			"deleteAccountVerifyToken":        "",
			"deleteAccountVerifyToken_expire": 0,
		})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) DeleteUserAndRelatedData(userId int64) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		//handle movie counts decrement
		err := deleteUserAndRelatedData_LikeDislikeMovies(tx, userId)
		if err != nil {
			return err
		}
		err = deleteUserAndRelatedData_WatchedMovies(tx, userId)
		if err != nil {
			return err
		}
		err = deleteUserAndRelatedData_FollowMovies(tx, userId)
		if err != nil {
			return err
		}
		err = deleteUserAndRelatedData_WatchListMovies(tx, userId)
		//----------------------------
		if err != nil {
			return err
		}
		err = deleteUserAndRelatedData_LikeDislikeStaff(tx, userId)
		if err != nil {
			return err
		}
		err = deleteUserAndRelatedData_FollowStaff(tx, userId)
		if err != nil {
			return err
		}
		//----------------------------
		err = deleteUserAndRelatedData_LikeDislikeCharacter(tx, userId)
		if err != nil {
			return err
		}
		err = deleteUserAndRelatedData_FavoriteCharacter(tx, userId)
		if err != nil {
			return err
		}
		//----------------------------

		// delete user data from User table
		err = tx.
			Where("\"userId\" = ?", userId).
			Delete(&model.User{}).
			Error
		if err != nil {
			// rollback
			return err
		}

		// return nil will commit the whole transaction
		return nil
	})

	return err
}

func deleteUserAndRelatedData_LikeDislikeMovies(tx *gorm.DB, userId int64) error {
	var rows []model.LikeDislikeMovie
	err := tx.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "movieId"}, {Name: "type"}}}).
		Where("\"userId\" = ?", userId).
		Delete(&rows).
		Error
	if err != nil {
		return err
	}

	likes := []string{}
	dislikes := []string{}
	for _, row := range rows {
		if row.Type == model.LIKE {
			likes = append(likes, row.MovieId)
		} else {
			dislikes = append(dislikes, row.MovieId)
		}
	}

	err = tx.Exec("update \"Movie\" set likes_count = likes_count - 1 where \"movieId\" IN ?", likes).Error
	if err != nil {
		return err
	}

	err = tx.Exec("update \"Movie\" set dislikes_count = dislikes_count - 1 where \"movieId\" IN ?", dislikes).Error
	if err != nil {
		return err
	}

	return nil
}

func deleteUserAndRelatedData_WatchedMovies(tx *gorm.DB, userId int64) error {
	var rows []model.WatchedMovie
	err := tx.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "movieId"}, {Name: "dropped"}, {Name: "favorite"}}}).
		Where("\"userId\" = ?", userId).
		Delete(&rows).
		Error
	if err != nil {
		return err
	}

	dropps := []string{}
	finishes := []string{}
	favorites := []string{}
	for _, row := range rows {
		if row.Dropped {
			dropps = append(dropps, row.MovieId)
		} else {
			finishes = append(finishes, row.MovieId)
		}
		if row.Favorite {
			favorites = append(favorites, row.MovieId)
		}
	}

	err = tx.Exec("update \"Movie\" set dropped_count = dropped_count - 1 where \"movieId\" IN ?", dropps).Error
	if err != nil {
		return err
	}

	err = tx.Exec("update \"Movie\" set finished_count = finished_count - 1 where \"movieId\" IN ?", finishes).Error
	if err != nil {
		return err
	}

	err = tx.Exec("update \"Movie\" set favorite_count = favorite_count - 1 where \"movieId\" IN ?", favorites).Error
	if err != nil {
		return err
	}

	return nil
}

func deleteUserAndRelatedData_FollowMovies(tx *gorm.DB, userId int64) error {
	var rows []model.FollowMovie
	err := tx.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "movieId"}}}).
		Where("\"userId\" = ?", userId).
		Delete(&rows).
		Error
	if err != nil {
		return err
	}

	follows := []string{}
	for _, row := range rows {
		follows = append(follows, row.MovieId)
	}

	err = tx.Exec("update \"Movie\" set follow_count = follow_count - 1 where \"movieId\" IN ?", follows).Error
	if err != nil {
		return err
	}

	return nil
}

func deleteUserAndRelatedData_WatchListMovies(tx *gorm.DB, userId int64) error {
	var rows []model.WatchListMovie
	err := tx.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "movieId"}}}).
		Where("\"userId\" = ?", userId).
		Delete(&rows).
		Error
	if err != nil {
		return err
	}

	watchLists := []string{}
	for _, row := range rows {
		watchLists = append(watchLists, row.MovieId)
	}

	err = tx.Exec("update \"Movie\" set watchlist_count = watchlist_count - 1 where \"movieId\" IN ?", watchLists).Error
	if err != nil {
		return err
	}

	return nil
}

func deleteUserAndRelatedData_LikeDislikeStaff(tx *gorm.DB, userId int64) error {
	var rows []model.LikeDislikeStaff
	err := tx.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "staffId"}, {Name: "type"}}}).
		Where("\"userId\" = ?", userId).
		Delete(&rows).
		Error
	if err != nil {
		return err
	}

	likes := []int64{}
	dislikes := []int64{}
	for _, row := range rows {
		if row.Type == model.LIKE {
			likes = append(likes, row.StaffId)
		} else {
			dislikes = append(dislikes, row.StaffId)
		}
	}

	err = tx.Exec("update \"Staff\" set likes_count = likes_count - 1 where \"id\" IN ?", likes).Error
	if err != nil {
		return err
	}

	err = tx.Exec("update \"Staff\" set dislikes_count = dislikes_count - 1 where \"id\" IN ?", dislikes).Error
	if err != nil {
		return err
	}

	return nil
}

func deleteUserAndRelatedData_FollowStaff(tx *gorm.DB, userId int64) error {
	var rows []model.FollowStaff
	err := tx.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "staffId"}}}).
		Where("\"userId\" = ?", userId).
		Delete(&rows).
		Error
	if err != nil {
		return err
	}

	follows := []int64{}
	for _, row := range rows {
		follows = append(follows, row.StaffId)
	}

	err = tx.Exec("update \"Staff\" set follow_count = follow_count - 1 where \"id\" IN ?", follows).Error
	if err != nil {
		return err
	}

	return nil
}

func deleteUserAndRelatedData_LikeDislikeCharacter(tx *gorm.DB, userId int64) error {
	var rows []model.LikeDislikeCharacter
	err := tx.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "characterId"}, {Name: "type"}}}).
		Where("\"userId\" = ?", userId).
		Delete(&rows).
		Error
	if err != nil {
		return err
	}

	likes := []int64{}
	dislikes := []int64{}
	for _, row := range rows {
		if row.Type == model.LIKE {
			likes = append(likes, row.CharacterId)
		} else {
			dislikes = append(dislikes, row.CharacterId)
		}
	}

	err = tx.Exec("update \"Character\" set likes_count = likes_count - 1 where \"id\" IN ?", likes).Error
	if err != nil {
		return err
	}

	err = tx.Exec("update \"Character\" set dislikes_count = dislikes_count - 1 where \"id\" IN ?", dislikes).Error
	if err != nil {
		return err
	}

	return nil
}

func deleteUserAndRelatedData_FavoriteCharacter(tx *gorm.DB, userId int64) error {
	var rows []model.FavoriteCharacter
	err := tx.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "characterId"}}}).
		Where("\"userId\" = ?", userId).
		Delete(&rows).
		Error
	if err != nil {
		return err
	}

	follows := []int64{}
	for _, row := range rows {
		follows = append(follows, row.CharacterId)
	}

	err = tx.Exec("update \"Character\" set favorite_count = favorite_count - 1 where \"id\" IN ?", follows).Error
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) GetProfileImagesCount(userId int64) (int64, error) {
	var result int64
	err := r.db.
		Model(&model.ProfileImage{}).
		Where("\"userId\" = ?", userId).
		Count(&result).
		Error

	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *UserRepository) GetProfileImages(userId int64) (*[]model.ProfileImageDataModel, error) {
	var result []model.ProfileImageDataModel
	err := r.db.
		Model(&model.ProfileImage{}).
		Where("\"userId\" = ?", userId).
		Order("\"addDate\" DESC").
		Find(&result).
		Error

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *UserRepository) SaveProfileImageData(profileImage *model.ProfileImage) error {
	err := r.db.Create(profileImage).Error
	return err
}

func (r *UserRepository) RemoveProfileImageData(userId int64, fileName string) error {
	queryStr := fmt.Sprintf("\"userId\" = %v AND url ~ '/%v$' ", userId, fileName)
	res := r.db.Where(queryStr).Delete(&model.ProfileImage{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *UserRepository) RemoveAllProfileImageData(userId int64) ([]model.ProfileImage, error) {
	var result []model.ProfileImage
	res := r.db.Where("\"userId\" = ?", userId).Delete(&result)
	if res.Error != nil {
		return nil, res.Error
	}
	return result, nil
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
		NotifToken:   device.NotifToken,
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
		NotifToken:   device.NotifToken,
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

func (r *UserRepository) UpdateSessionNotifToken(userId int64, refreshToken string, notifToken string) error {
	activeSession := model.ActiveSession{}
	result := r.db.Model(&activeSession).
		Where("\"userId\" = ? AND \"refreshToken\" = ?", userId, refreshToken).
		UpdateColumn("\"notifToken\"", notifToken)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
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

func (r *UserRepository) RemoveAuthSession(userId int64, refreshToken string, deviceId string) (*model.ActiveSession, error) {
	var searchResult model.ActiveSession
	err := r.db.
		Model(&model.ActiveSession{}).
		Where("\"userId\" = ? AND \"refreshToken\" = ?", userId, refreshToken).
		Limit(1).
		Find(&searchResult).
		Error
	if err != nil {
		return nil, err
	}

	var result []model.ActiveSession
	err = r.db.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "refreshToken"}}}).
		Where("\"userId\" = ? AND \"deviceId\" = ?", userId, deviceId).
		Delete(&result).
		Error

	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &result[0], nil
}

func (r *UserRepository) RemoveAllAuthSession(userId int64, refreshToken string) ([]model.ActiveSession, error) {
	var searchResult model.ActiveSession
	err := r.db.
		Model(&model.ActiveSession{}).
		Where("\"userId\" = ? AND \"refreshToken\" = ?", userId, refreshToken).
		Limit(1).
		Find(&searchResult).
		Error
	if err != nil {
		return nil, err
	}

	var result []model.ActiveSession
	err = r.db.
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "refreshToken"}}}).
		Where("\"userId\" = ? AND \"refreshToken\" != ?", userId, refreshToken).
		Delete(&result).
		Error
	if err != nil {
		return nil, err
	}

	return result, nil
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
	err := r.db.Where("\"followerId\" = ? AND \"followingId\" = ?", userId, followId).Delete(&model.Follow{}).Error
	return err
}

func (r *UserRepository) GetUserFollowers(userId int64, skip int, limit int) ([]model.FollowUserDataModel, error) {
	var result []model.FollowUserDataModel
	err := r.db.Model(&model.User{}).Joins("join \"Follow\" on \"userId\" = \"followerId\" AND \"followingId\" = ? ", userId).
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
		Preload("ActiveSessions", func(db *gorm.DB) *gorm.DB {
			return db.Select("userId", "notifToken")
		}).
		Find(&result).
		Error

	if err != nil {
		return nil, err
	}
	return &result, nil
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) GetUserDownloadLinkSettings(userId int64) (*model.DownloadLinksSettings, error) {
	var result model.DownloadLinksSettings
	err := r.db.
		Model(&model.DownloadLinksSettings{}).
		Where("\"userId\" = ?", userId).
		Limit(1).
		Find(&result).
		Error

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *UserRepository) GetUserNotificationSettings(userId int64) (*model.NotificationSettings, error) {
	var result model.NotificationSettings
	err := r.db.
		Model(&model.NotificationSettings{}).
		Where("\"userId\" = ?", userId).
		Limit(1).
		Find(&result).
		Error

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *UserRepository) GetUserMovieSettings(userId int64) (*model.MovieSettings, error) {
	var result model.MovieSettings
	err := r.db.
		Model(&model.MovieSettings{}).
		Where("\"userId\" = ?", userId).
		Limit(1).
		Find(&result).
		Error

	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *UserRepository) UpdateUserDownloadLinkSettings(userId int64, settings model.DownloadLinksSettings) error {
	err := r.db.
		Model(&model.DownloadLinksSettings{}).
		Where("\"userId\" = ?", userId).
		Updates(map[string]interface{}{
			"includeCensored":    settings.IncludeCensored,
			"includeDubbed":      settings.IncludeDubbed,
			"includeHardSub":     settings.IncludeHardSub,
			"preferredQualities": settings.PreferredQualities,
		}).
		Error
	return err
}

func (r *UserRepository) UpdateUserNotificationSettings(userId int64, settings model.NotificationSettings) error {
	err := r.db.
		Model(&model.NotificationSettings{}).
		Where("\"userId\" = ?", userId).
		Updates(map[string]interface{}{
			"newFollower":                settings.NewFollower,
			"newMessage":                 settings.NewMessage,
			"finishedList_spinOffSequel": settings.FinishedListSpinOffSequel,
			"followMovie":                settings.FollowMovie,
			"followMovie_betterQuality":  settings.FollowMovieBetterQuality,
			"followMovie_subtitle":       settings.FollowMovieSubtitle,
			"futureList":                 settings.FutureList,
			"futureList_serialSeasonEnd": settings.FutureListSerialSeasonEnd,
			"futureList_subtitle":        settings.FutureListSubtitle,
		}).
		Error
	return err
}

func (r *UserRepository) UpdateUserMovieSettings(userId int64, settings model.MovieSettings) error {
	err := r.db.
		Model(&model.MovieSettings{}).
		Where("\"userId\" = ?", userId).
		Updates(map[string]interface{}{
			"includeAnime":  settings.IncludeAnime,
			"includeHentai": settings.IncludeHentai,
		}).
		Error
	return err
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) UpdateUserFavoriteGenres(userId int64, genresArray []string) error {
	err := r.db.
		Model(&model.User{}).
		Where("\"userId\" = ?", userId).
		UpdateColumn("favoriteGenres", pq.StringArray(genresArray)).
		Error
	return err
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) GetActiveSessions(userId int64) ([]model.ActiveSessionDataModel, error) {
	var result []model.ActiveSessionDataModel
	err := r.db.
		Model(&model.ActiveSessionDataModel{}).
		Where("\"userId\" = ?", userId).
		Find(&result).
		Error
	return result, err
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) GetUserBots(userId int64) ([]model.UserBotDataModel, error) {
	var result []model.UserBotDataModel
	err := r.db.
		Model(&model.UserBotDataModel{}).
		Where("\"userId\" = ? AND notification = true", userId).
		Find(&result).
		Error
	return result, err
}

func (r *UserRepository) GetBotData(botId string) (*model.Bot, error) {
	var result model.Bot
	err := r.db.
		Model(&model.Bot{}).
		Where("\"botId\" = ?", botId).
		Find(&result).
		Error
	return &result, err
}

//------------------------------------------
//------------------------------------------

func (r *UserRepository) GetUserRoles(userId int64) ([]model.Role, error) {
	var roles []model.Role

	// Fetch the roles for this user
	if err := r.db.Model(&model.Role{}).
		Joins("JOIN \"UserToRole\" ON \"UserToRole\".\"roleId\" = \"Role\".id").
		Where("\"UserToRole\".\"userId\" = ?", userId).
		Find(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *UserRepository) GetUserRolesWithPermissions(userId int64) ([]model.RoleWithPermissions, error) {
	roles := []model.RoleWithPermissions{}

	type resType struct {
		model.Role
		model.Permission
	}
	var res []resType

	queryStr := `
		SELECT *
		FROM "UserToRole" ur
        	JOIN "Role" r ON ur."roleId" = r.id
        	JOIN "RoleToPermission" rp ON r.id = rp."roleId"
        	JOIN "Permission" p ON rp."permissionId" = p.id
		WHERE
			ur."userId" = @uid;`

	err := r.db.Raw(queryStr,
		map[string]interface{}{
			"uid": userId,
		}).
		Scan(&res).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			t := make([]model.RoleWithPermissions, 0)
			return t, nil
		}
		return nil, err
	}

	for _, item := range res {
		exist := false
		for i2 := range roles {
			if roles[i2].Id == item.Role.Id {
				roles[i2].Permissions = append(roles[i2].Permissions, item.Permission)
				exist = true
			}
		}
		if exist {
			continue
		}
		newRole := model.RoleWithPermissions{
			Id:                  item.Role.Id,
			Name:                item.Role.Name,
			Description:         item.Role.Description,
			TorrentLeachLimitGb: item.Role.TorrentLeachLimitGb,
			TorrentSearchLimit:  item.Role.TorrentSearchLimit,
			BotsNotification:    item.Role.BotsNotification,
			CreatedAt:           item.Role.CreatedAt,
			UpdatedAt:           item.Role.UpdatedAt,
			Permissions:         []model.Permission{item.Permission},
		}
		roles = append(roles, newRole)
	}

	return roles, nil
}

func (r *UserRepository) GetUserPermissionsByRoleIds(roleIds []int64) ([]model.Permission, error) {
	var permissions []model.Permission

	if err := r.db.Model(&model.Permission{}).
		Joins("JOIN \"RoleToPermission\" ON \"RoleToPermission\".\"permissionId\" = \"Permission\".id").
		Where("\"RoleToPermission\".\"roleId\" in ?", roleIds).
		Find(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

//------------------------------------------
//------------------------------------------
