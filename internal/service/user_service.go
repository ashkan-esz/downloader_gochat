package service

import (
	"context"
	"downloader_gochat/cloudStorage"
	"downloader_gochat/configs"
	"downloader_gochat/internal/repository"
	"downloader_gochat/model"
	"downloader_gochat/pkg/email"
	"downloader_gochat/pkg/geoip"
	"downloader_gochat/pkg/response"
	"downloader_gochat/rabbitmq"
	"downloader_gochat/util"
	"errors"
	"fmt"
	"mime/multipart"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type IUserService interface {
	SignUp(registerVM *model.RegisterViewModel, ip string) (*model.UserViewModel, error)
	LoginUser(loginVM *model.LoginViewModel, ip string) (*model.UserViewModel, error)
	GetToken(deviceVM *model.DeviceInfo, prevRefreshToken string, jwtUserData *util.MyJwtClaims, addProfileImages bool, ip string) (*model.UserViewModel, *util.TokenDetail, error)
	LogOut(c *fiber.Ctx, jwtUserData *util.MyJwtClaims, prevRefreshToken string) error
	ForceLogoutDevice(c *fiber.Ctx, jwtUserData *util.MyJwtClaims, refreshToken string, deviceId string) error
	ForceLogoutAll(c *fiber.Ctx, jwtUserData *util.MyJwtClaims, refreshToken string) error
	SetNotifToken(jwtUserData *util.MyJwtClaims, refreshToken string, notifToken string) error
	FollowUser(jwtUserData *util.MyJwtClaims, followId int64) error
	UnFollowUser(jwtUserData *util.MyJwtClaims, followId int64) error
	GetUserFollowers(userId int64, skip int, limit int) ([]model.FollowUserDataModel, error)
	GetUserFollowings(userId int64, skip int, limit int) ([]model.FollowUserDataModel, error)
	GetUserSettings(userId int64, settingName model.SettingName) (*model.UserSettingsRes, error)
	UpdateUserSettings(userId int64, settingName model.SettingName, settings *model.UserSettingsRes) error
	UpdateUserFavoriteGenres(userId int64, genresArray []string) error
	GetActiveSessions(userId int64, refreshToken string) (*model.ActiveSessionRes, error)
	GetUserProfile(requestParams *model.UserProfileReq) (*model.UserProfileRes, error)
	GetUserRolePermission(requestParams *model.UserProfileReq) (*model.UserRolePermissionRes, error)
	EditUserProfile(userId int64, editFields *model.EditProfileReq) (*model.UserDataModel, error)
	UpdateUserPassword(userId int64, passwords *model.UpdatePasswordReq) error
	SendVerifyEmail(userId int64) error
	VerifyEmail(userId int64, token string) error
	SendDeleteAccount(userId int64) error
	DeleteUserAccount(userId int64, token string) error
	GetProfileImagesCount(userId int64) (int64, error)
	UploadProfileImage(userId int64, contentType string, fileSize int64, fileBuffer multipart.File) (*[]model.ProfileImageDataModel, error)
	RemoveProfileImage(userId int64, fileName string) (*[]model.ProfileImageDataModel, error)
}

type UserService struct {
	userRepo     repository.IUserRepository
	rabbitmq     rabbitmq.RabbitMQ
	cloudStorage cloudStorage.IS3Storage
	timeout      time.Duration
}

func NewUserService(userRepo repository.IUserRepository, rabbit rabbitmq.RabbitMQ, cloudStorage cloudStorage.IS3Storage) *UserService {
	return &UserService{
		userRepo:     userRepo,
		rabbitmq:     rabbit,
		cloudStorage: cloudStorage,
		timeout:      time.Duration(2) * time.Second,
	}
}

//------------------------------------------
//------------------------------------------

func (s *UserService) SignUp(registerVM *model.RegisterViewModel, ip string) (*model.UserViewModel, error) {
	searchResult, err := s.userRepo.GetUserByUsernameEmail(registerVM.Username, registerVM.Email)
	if err != nil {
		return nil, err
	}
	if searchResult != nil {
		return &model.UserViewModel{
			UserId:   0,
			Username: searchResult.Username,
			Email:    searchResult.Email,
		}, nil
	}

	var user = model.User{
		Username:               strings.ToLower(registerVM.Username),
		RawUsername:            registerVM.Username,
		PublicName:             registerVM.Username,
		Email:                  strings.ToLower(registerVM.Email),
		EmailVerified:          false,
		EmailVerifyTokenExpire: time.Now().Add(6 * time.Hour).UnixMilli(),
		MbtiType:               model.ENTJ,
		DefaultProfile:         configs.GetConfigs().DefaultProfileImage,
	}

	err = user.EncryptPassword(registerVM.Password)
	if err != nil {
		return nil, err
	}
	err = user.EncryptEmailToken(uuid.NewString())
	if err != nil {
		return nil, err
	}

	result, err := s.userRepo.AddUser(&user, int64(model.DefaultUser))
	if err != nil {
		return nil, err
	}

	token, err := util.CreateJwtToken(result.UserId, result.Username, []int64{int64(model.DefaultUser)})
	if err != nil {
		return nil, err
	}

	deviceId := registerVM.DeviceInfo.Fingerprint
	if deviceId == "" {
		deviceId = uuid.NewString()
	}
	ipLocation := geoip.GetRequestLocation(ip)
	err = s.userRepo.AddSession(&registerVM.DeviceInfo, deviceId, result.UserId, token.RefreshToken, ipLocation)
	if err != nil {
		return nil, err
	}

	//-----------------------------------
	verifyEmailUrl := fmt.Sprintf("%v/v1/user/VerifyEmail/%v/%v",
		configs.GetConfigs().ServerAddress, user.UserId, user.EmailVerifyToken)
	queueConf := rabbitmq.NewConfigPublish(rabbitmq.EmailExchange, rabbitmq.EmailBindingKey)
	emailData := email.EmailQueueData{
		Type:        email.UserRegistration,
		UserId:      user.UserId,
		RawUsername: user.RawUsername,
		Email:       registerVM.Email,
		Token:       "",
		Host:        "",
		Url:         verifyEmailUrl,
		DeviceInfo:  nil,
		IpLocation:  ipLocation,
	}
	s.rabbitmq.Publish(context.TODO(), emailData, queueConf, user.UserId)
	//-----------------------------------

	userVM := model.UserViewModel{
		UserId:   result.UserId,
		Username: result.Username,
		Email:    result.Email,
		Token: model.TokenViewModel{
			AccessToken:       token.AccessToken,
			AccessTokenExpire: token.ExpiresAt,
			RefreshToken:      token.RefreshToken,
		},
		RoleIds:   []int64{int64(model.DefaultUser)},
		RoleNames: []string{string(model.DefaultUserRole)},
	}

	return &userVM, nil
}

func (s *UserService) LoginUser(loginVM *model.LoginViewModel, ip string) (*model.UserViewModel, error) {
	searchResult, err := s.userRepo.GetUserByUsernameEmail(loginVM.Email, loginVM.Email)
	if err != nil {
		return nil, err
	}
	if searchResult == nil {
		return nil, nil
	}

	err = loginVM.CheckPassword(loginVM.Password, searchResult.Password)
	if err != nil {
		return nil, err
	}

	roles, err := s.userRepo.GetUserRoles(searchResult.UserId)
	if err != nil {
		return nil, err
	}

	roleIds := []int64{}
	roleNames := []string{}
	for _, r := range roles {
		roleIds = append(roleIds, r.Id)
		roleNames = append(roleNames, r.Name)
	}

	token, err := util.CreateJwtToken(searchResult.UserId, searchResult.Username, roleIds)
	if err != nil {
		return nil, err
	}

	deviceId := loginVM.DeviceInfo.Fingerprint
	if deviceId == "" {
		deviceId = uuid.NewString() + "-" + strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	ipLocation := geoip.GetRequestLocation(ip)
	isNewDevice, err := s.userRepo.UpdateSession(&loginVM.DeviceInfo, deviceId, searchResult.UserId, token.RefreshToken, ipLocation)
	if err != nil {
		return nil, err
	}

	if isNewDevice {
		//-----------------------------------
		queueConf := rabbitmq.NewConfigPublish(rabbitmq.EmailExchange, rabbitmq.EmailBindingKey)
		emailData := email.EmailQueueData{
			Type:        email.UserLogin,
			UserId:      searchResult.UserId,
			RawUsername: searchResult.Username,
			Email:       loginVM.Email,
			Token:       "",
			Host:        "",
			Url:         "",
			DeviceInfo:  &loginVM.DeviceInfo,
			IpLocation:  ipLocation,
		}
		s.rabbitmq.Publish(context.TODO(), emailData, queueConf, searchResult.UserId)
		//-----------------------------------

		activeSessions, err := s.userRepo.GetUserActiveSessions(searchResult.UserId)
		if err != nil {
			err := s.userRepo.RemoveSession(searchResult.UserId, token.RefreshToken)
			return nil, err
		}
		if len(activeSessions) > configs.GetConfigs().ActiveSessionsLimit {
			lastUsedSessions := activeSessions[5:]
			tokens := make([]string, len(lastUsedSessions))
			for i, session := range lastUsedSessions {
				tokens[i] = session.RefreshToken
			}
			err := s.userRepo.RemoveSessions(searchResult.UserId, tokens)
			if err != nil {
				return nil, err
			}
		}
	}

	userVM := model.UserViewModel{
		UserId:   searchResult.UserId,
		Username: searchResult.Username,
		Email:    searchResult.Email,
		Token: model.TokenViewModel{
			AccessToken:       token.AccessToken,
			AccessTokenExpire: token.ExpiresAt,
			RefreshToken:      token.RefreshToken,
		},
		RoleIds:   roleIds,
		RoleNames: roleNames,
	}

	return &userVM, nil
}

func (s *UserService) GetToken(deviceVM *model.DeviceInfo, prevRefreshToken string, jwtUserData *util.MyJwtClaims, addProfileImages bool, ip string) (*model.UserViewModel, *util.TokenDetail, error) {
	roles, err := s.userRepo.GetUserRoles(jwtUserData.UserId)
	if err != nil {
		return nil, nil, err
	}

	roleIds := []int64{}
	roleNames := []string{}
	for _, r := range roles {
		roleIds = append(roleIds, r.Id)
		roleNames = append(roleNames, r.Name)
	}

	token, err := util.CreateJwtToken(jwtUserData.UserId, jwtUserData.Username, roleIds)
	if err != nil {
		return nil, nil, err
	}

	ipLocation := geoip.GetRequestLocation(ip)
	sessionData, err := s.userRepo.UpdateSessionRefreshToken(deviceVM, jwtUserData.UserId, token.RefreshToken, prevRefreshToken, ipLocation)
	if err != nil {
		return nil, nil, err
	}
	if sessionData == nil {
		return nil, nil, nil
	}

	userVM := model.UserViewModel{
		UserId: jwtUserData.UserId,
		Token: model.TokenViewModel{
			AccessToken:       token.AccessToken,
			AccessTokenExpire: token.ExpiresAt,
			RefreshToken:      token.RefreshToken,
		},
		RoleIds:   roleIds,
		RoleNames: roleNames,
	}

	if addProfileImages {
		userData, err := s.userRepo.GetDetailUser(jwtUserData.UserId)
		if err != nil {
			return nil, token, err
		}
		if sessionData == nil {
			return nil, token, nil
		}
		userVM.ProfileImages = userData.ProfileImages
		userVM.Username = userData.Username
	}

	return &userVM, token, nil
}

func (s *UserService) LogOut(c *fiber.Ctx, jwtUserData *util.MyJwtClaims, prevRefreshToken string) error {
	err := s.userRepo.RemoveSession(jwtUserData.UserId, prevRefreshToken)
	if err != nil {
		return err
	}

	remainingTime := jwtUserData.ExpiresAt - time.Now().UnixMilli()
	err = setJwtDataCache(prevRefreshToken, "logout", time.Duration(remainingTime)*time.Millisecond)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) ForceLogoutDevice(c *fiber.Ctx, jwtUserData *util.MyJwtClaims, refreshToken string, deviceId string) error {
	result, err := s.userRepo.RemoveAuthSession(jwtUserData.UserId, refreshToken, deviceId)
	if err != nil {
		return err
	}

	remainingTime := configs.GetConfigs().AccessTokenExpireHour
	err = setJwtDataCache(result.RefreshToken, "logout", time.Duration(remainingTime)*time.Hour)
	if err != nil {
		return err
	}

	if result.NotifToken != "" {
		_ = removeNotifTokenFromCachedUserData(jwtUserData.UserId, result.NotifToken)
	}

	return nil
}

func (s *UserService) ForceLogoutAll(c *fiber.Ctx, jwtUserData *util.MyJwtClaims, refreshToken string) error {
	result, err := s.userRepo.RemoveAllAuthSession(jwtUserData.UserId, refreshToken)
	if err != nil {
		return err
	}

	remainingTime := configs.GetConfigs().AccessTokenExpireHour
	for i := range result {
		_ = setJwtDataCache(result[i].RefreshToken, "logout", time.Duration(remainingTime)*time.Hour)
		if result[i].NotifToken != "" {
			_ = removeNotifTokenFromCachedUserData(jwtUserData.UserId, result[i].NotifToken)
		}
	}

	return nil
}

func (s *UserService) SetNotifToken(jwtUserData *util.MyJwtClaims, refreshToken string, notifToken string) error {
	err := s.userRepo.UpdateSessionNotifToken(jwtUserData.UserId, refreshToken, notifToken)
	if err == nil && notifToken != "" {
		_ = addNotifTokenToCachedUserData(jwtUserData.UserId, notifToken)
	}
	return err
}

//------------------------------------------
//------------------------------------------

func (s *UserService) FollowUser(jwtUserData *util.MyJwtClaims, followId int64) error {
	err := s.userRepo.AddUserFollow(jwtUserData.UserId, followId)
	if err == nil {
		// need to save the notification, show notification in app, send push-notification to followed user
		ctx, _ := context.WithCancel(context.Background())
		//defer cancel()
		readQueueConf := rabbitmq.NewConfigPublish(rabbitmq.NotificationExchange, rabbitmq.NotificationBindingKey)
		message := model.CreateFollowNotificationAction(jwtUserData.UserId, followId)
		s.rabbitmq.Publish(ctx, message, readQueueConf, followId)
	}
	return err
}

func (s *UserService) UnFollowUser(jwtUserData *util.MyJwtClaims, followId int64) error {
	err := s.userRepo.RemoveUserFollow(jwtUserData.UserId, followId)
	return err
}

func (s *UserService) GetUserFollowers(userId int64, skip int, limit int) ([]model.FollowUserDataModel, error) {
	result, err := s.userRepo.GetUserFollowers(userId, skip, limit)
	return result, err
}

func (s *UserService) GetUserFollowings(userId int64, skip int, limit int) ([]model.FollowUserDataModel, error) {
	result, err := s.userRepo.GetUserFollowings(userId, skip, limit)
	return result, err
}

//------------------------------------------
//------------------------------------------

func (s *UserService) GetUserSettings(userId int64, settingName model.SettingName) (*model.UserSettingsRes, error) {
	result := model.UserSettingsRes{
		DownloadLinksSettings: nil,
		NotificationSettings:  nil,
		MovieSettings:         nil,
	}
	if settingName == model.AllSettingsName {
		notif, err := s.userRepo.GetUserNotificationSettings(userId)
		if err != nil {
			return nil, err
		}
		result.NotificationSettings = notif

		download, err := s.userRepo.GetUserDownloadLinkSettings(userId)
		if err != nil {
			return nil, err
		}
		result.DownloadLinksSettings = download

		movie, err := s.userRepo.GetUserMovieSettings(userId)
		if err != nil {
			return nil, err
		}
		result.MovieSettings = movie
	} else {
		switch settingName {
		case model.NotificationSettingsName:
			notif, err := s.userRepo.GetUserNotificationSettings(userId)
			if err != nil {
				return nil, err
			}
			result.NotificationSettings = notif
		case model.DownloadSettingsName:
			download, err := s.userRepo.GetUserDownloadLinkSettings(userId)
			if err != nil {
				return nil, err
			}
			result.DownloadLinksSettings = download
		case model.MovieSettingsName:
			movie, err := s.userRepo.GetUserMovieSettings(userId)
			if err != nil {
				return nil, err
			}
			result.MovieSettings = movie
		default:
			return nil, errors.New("invalid settingName")
		}
	}
	return &result, nil
}

func (s *UserService) UpdateUserSettings(userId int64, settingName model.SettingName, settings *model.UserSettingsRes) error {
	switch settingName {
	case model.NotificationSettingsName:
		err := s.userRepo.UpdateUserNotificationSettings(userId, *settings.NotificationSettings)
		if err == nil {
			_ = updateNotificationSettingsOfCachedUserData(userId, *settings.NotificationSettings)
			client, ok := getClientFromHub(userId)
			if ok {
				client.Message <- model.CreateNotificationSettingsAction(settings.NotificationSettings)
			}
		}
		return err
	case model.DownloadSettingsName:
		err := s.userRepo.UpdateUserDownloadLinkSettings(userId, *settings.DownloadLinksSettings)
		return err
	case model.MovieSettingsName:
		err := s.userRepo.UpdateUserMovieSettings(userId, *settings.MovieSettings)
		return err
	default:
		return errors.New("invalid settingName")
	}
}

//------------------------------------------
//------------------------------------------

func (s *UserService) UpdateUserFavoriteGenres(userId int64, genresArray []string) error {
	err := s.userRepo.UpdateUserFavoriteGenres(userId, genresArray)
	return err
}

//------------------------------------------
//------------------------------------------

func (s *UserService) GetActiveSessions(userId int64, refreshToken string) (*model.ActiveSessionRes, error) {
	sessions, err := s.userRepo.GetActiveSessions(userId)
	if err != nil {
		return nil, err
	}

	thisDeviceIndex := slices.IndexFunc(sessions, func(s model.ActiveSessionDataModel) bool {
		return s.RefreshToken == refreshToken
	})
	thisDevice := sessions[thisDeviceIndex]
	activeSessions := slices.DeleteFunc(sessions, func(s model.ActiveSessionDataModel) bool {
		return s.RefreshToken == refreshToken
	})
	result := model.ActiveSessionRes{
		ThisDevice:     &thisDevice,
		ActiveSessions: &activeSessions,
	}

	return &result, err
}

//------------------------------------------
//------------------------------------------

func (s *UserService) GetUserProfile(requestParams *model.UserProfileReq) (*model.UserProfileRes, error) {
	result, err := s.userRepo.GetUserProfile(requestParams)
	return result, err
}

func (s *UserService) GetUserRolePermission(requestParams *model.UserProfileReq) (*model.UserRolePermissionRes, error) {
	res := &model.UserRolePermissionRes{
		RolesWithPermissions: nil,
		Roles:                nil,
	}

	if requestParams.LoadRolesWithPermissions {
		rolesWithPermissions, err := s.userRepo.GetUserRolesWithPermissions(requestParams.UserId)
		if err != nil {
			return nil, err
		}
		res.RolesWithPermissions = rolesWithPermissions
	} else {
		roles, err := s.userRepo.GetUserRoles(requestParams.UserId)
		if err != nil {
			return nil, err
		}
		res.Roles = roles
	}

	return res, nil
}

func (s *UserService) EditUserProfile(userId int64, editFields *model.EditProfileReq) (*model.UserDataModel, error) {
	searchResult, err := s.userRepo.GetUserByUsernameEmailAndUserId(userId, editFields.Username, editFields.Email)
	if err != nil {
		return nil, err
	}
	if searchResult != nil {
		// email or username already taken by another user
		return &model.UserDataModel{
			UserId:   0,
			Username: searchResult.Username,
			Email:    searchResult.Email,
		}, nil
	} else {
		searchResult, err = s.userRepo.GetUserMetaData(userId)
		if err != nil {
			return nil, err
		}
	}

	updateFields := map[string]interface{}{
		"username":   editFields.Username,
		"publicName": editFields.PublicName,
		"email":      editFields.Email,
		"bio":        editFields.Bio,
		"mbtiType":   editFields.MbtiType,
	}
	if searchResult.Username != editFields.Username {
		updateFields["username"] = strings.ToLower(editFields.Username)
		updateFields["rawUsername"] = editFields.Username
		editFields.Username = strings.ToLower(editFields.Username)
	}
	if searchResult.Email != editFields.Email {
		updateFields["email"] = editFields.Email
		updateFields["emailVerified"] = false
		updateFields["emailVerifyToken"] = ""
		updateFields["emailVerifyToken_expire"] = 0
	}

	err = s.userRepo.EditUserProfile(userId, editFields, updateFields)

	if err == nil {
		_ = updateProfileDataOfCachedUserData(userId, strings.ToLower(editFields.Username), editFields.PublicName)
		client, ok := getClientFromHub(userId)
		if ok {
			client.Message <- model.CreateUpdateProfileAction(editFields)
		}
	}

	return searchResult, err
}

func (s *UserService) UpdateUserPassword(userId int64, passwords *model.UpdatePasswordReq) error {
	searchResult, err := s.userRepo.GetUserMetaData(userId)
	if err != nil {
		return err
	}

	err = util.CheckPassword(passwords.NewPassword, searchResult.Password)
	if err != nil {
		return errors.New(response.OldPassNotMatch)
	}

	hashedNewPassword, err := util.HashPassword(passwords.NewPassword)
	if err != nil {
		return err
	}
	passwords.NewPassword = hashedNewPassword

	err = s.userRepo.UpdateUserPassword(userId, passwords)

	if err == nil && searchResult.Email != "" {
		// send email
		//-----------------------------------
		queueConf := rabbitmq.NewConfigPublish(rabbitmq.EmailExchange, rabbitmq.EmailBindingKey)
		emailData := email.EmailQueueData{
			Type:        email.PasswordUpdated,
			UserId:      searchResult.UserId,
			RawUsername: searchResult.Username,
			Email:       searchResult.Email,
			Token:       "",
			Host:        "",
			Url:         "",
			DeviceInfo:  nil,
			IpLocation:  "",
		}
		s.rabbitmq.Publish(context.TODO(), emailData, queueConf, searchResult.UserId)
		//-----------------------------------
	}

	return err
}

//------------------------------------------
//------------------------------------------

func (s *UserService) SendVerifyEmail(userId int64) error {
	searchResult, err := s.userRepo.GetUserMetaData(userId)
	if err != nil {
		return err
	}

	hashToken, err := util.HashPassword(uuid.NewString())
	if err != nil {
		return err
	}
	verifyToken := strings.ReplaceAll(hashToken, "/", "")
	verifyTokenExpire := time.Now().Add(6 * time.Hour).UnixMilli()

	err = s.userRepo.SaveUserEmailToken(userId, verifyToken, verifyTokenExpire)
	if err != nil {
		return err
	}

	//-----------------------------------
	verifyEmailUrl := fmt.Sprintf("%v/v1/user/VerifyEmail/%v/%v",
		configs.GetConfigs().ServerAddress, searchResult.UserId, verifyToken)
	queueConf := rabbitmq.NewConfigPublish(rabbitmq.EmailExchange, rabbitmq.EmailBindingKey)
	queueConf.Expiration = strconv.FormatInt(verifyTokenExpire, 10)
	emailData := email.EmailQueueData{
		Type:        email.VerifyEmail,
		UserId:      searchResult.UserId,
		RawUsername: searchResult.Username,
		Email:       searchResult.Email,
		Token:       verifyToken,
		Host:        "",
		Url:         verifyEmailUrl,
		DeviceInfo:  nil,
		IpLocation:  "",
	}
	s.rabbitmq.Publish(context.TODO(), emailData, queueConf, searchResult.UserId)
	//-----------------------------------
	return err
}

func (s *UserService) VerifyEmail(userId int64, token string) error {
	err := s.userRepo.VerifyUserEmailToken(userId, token)
	return err
}

func (s *UserService) SendDeleteAccount(userId int64) error {
	searchResult, err := s.userRepo.GetUserMetaData(userId)
	if err != nil {
		return err
	}

	hashToken, err := util.HashPassword(uuid.NewString())
	if err != nil {
		return err
	}
	verifyToken := strings.ReplaceAll(hashToken, "/", "")
	verifyTokenExpire := time.Now().Add(10 * time.Minute).UnixMilli()

	err = s.userRepo.SaveDeleteAccountToken(userId, verifyToken, verifyTokenExpire)
	if err != nil {
		return err
	}

	//-----------------------------------
	deleteUrl := fmt.Sprintf("%v/v1/user/deleteAccount/%v/%v",
		configs.GetConfigs().ServerAddress, searchResult.UserId, verifyToken)
	queueConf := rabbitmq.NewConfigPublish(rabbitmq.EmailExchange, rabbitmq.EmailBindingKey)
	queueConf.Expiration = strconv.FormatInt(verifyTokenExpire, 10)
	emailData := email.EmailQueueData{
		Type:        email.DeleteAccount,
		UserId:      searchResult.UserId,
		RawUsername: searchResult.Username,
		Email:       searchResult.Email,
		Token:       verifyToken,
		Host:        "",
		Url:         deleteUrl,
		DeviceInfo:  nil,
		IpLocation:  "",
	}
	s.rabbitmq.Publish(context.TODO(), emailData, queueConf, searchResult.UserId)
	//-----------------------------------
	return err
}

func (s *UserService) DeleteUserAccount(userId int64, token string) error {
	err := s.userRepo.VerifyDeleteAccountToken(userId, token)
	if err != nil {
		return err
	}

	roles, err := s.userRepo.GetUserRoles(userId)
	if err != nil {
		return err
	}
	for _, r := range roles {
		if r.Id == int64(model.MainAdmin) || r.Name == string(model.MainAdminRole) {
			return errors.New("forbidden, cannot remove user with main-admin-role")
		}
	}

	profileImages, err := s.userRepo.RemoveAllProfileImageData(userId)
	if err != nil {
		return err
	}
	for i := range profileImages {
		temp := strings.Split(profileImages[i].Url, "/")
		filename := temp[len(temp)-1]
		_ = s.cloudStorage.RemoveFile(cloudStorage.ProfileImageBucketName, filename)
	}

	activeSessions, err := s.userRepo.GetActiveSessions(userId)
	if err != nil {
		return err
	}

	err = s.userRepo.DeleteUserAndRelatedData(userId)
	if err != nil {
		return err
	}

	accessTokenExpireHour := configs.GetConfigs().AccessTokenExpireHour
	for i := range activeSessions {
		_ = setJwtDataCache(activeSessions[i].RefreshToken, "deleteAccount", time.Duration(accessTokenExpireHour)*time.Hour)
	}

	return nil
}

//------------------------------------------
//------------------------------------------

func (s *UserService) GetProfileImagesCount(userId int64) (int64, error) {
	result, err := s.userRepo.GetProfileImagesCount(userId)
	return result, err
}

func (s *UserService) UploadProfileImage(userId int64, contentType string, fileSize int64, fileBuffer multipart.File) (*[]model.ProfileImageDataModel, error) {
	saveType := "jpg"
	if contentType == "image/png" {
		saveType = "png"
	}
	savingFileName := fmt.Sprintf("user-%v-%v.%v", userId, time.Now().UnixMilli(), saveType)
	result, err := s.cloudStorage.UploadLargeFile(cloudStorage.ProfileImageBucketName, savingFileName, fileBuffer)
	if err != nil {
		return nil, err
	}

	profileImage := model.ProfileImage{
		UserId:       userId,
		AddDate:      time.Now().UTC(),
		Url:          result.Location,
		OriginalSize: fileSize,
		Size:         fileSize,
		Thumbnail:    "",
		BlurHash:     "",
	}
	profileImage.Thumbnail, profileImage.BlurHash = createThumbnailAndBlurHash(contentType, fileBuffer)

	err = s.userRepo.SaveProfileImageData(&profileImage)
	if err != nil {
		_ = s.cloudStorage.RemoveFile(cloudStorage.ProfileImageBucketName, savingFileName)
		return nil, err
	}

	images, err := s.userRepo.GetProfileImages(userId)

	if err == nil {
		_ = updateProfileImageOfCachedUserData(userId, images)
		client, ok := getClientFromHub(userId)
		if ok {
			client.Message <- model.CreateUpdateProfileImagesAction(images)
		}
	}

	return images, err
}

func (s *UserService) RemoveProfileImage(userId int64, fileName string) (*[]model.ProfileImageDataModel, error) {
	err := s.userRepo.RemoveProfileImageData(userId, fileName)
	if err != nil {
		return nil, err
	}

	err = s.cloudStorage.RemoveFile(cloudStorage.ProfileImageBucketName, fileName)
	if err != nil {
		return nil, err
	}

	images, err := s.userRepo.GetProfileImages(userId)

	if err == nil {
		_ = updateProfileImageOfCachedUserData(userId, images)
		client, ok := getClientFromHub(userId)
		if ok {
			client.Message <- model.CreateUpdateProfileImagesAction(images)
		}
	}

	return images, err
}
