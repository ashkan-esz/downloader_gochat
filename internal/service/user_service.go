package service

import (
	"downloader_gochat/configs"
	"downloader_gochat/db/redis"
	"downloader_gochat/internal/repository"
	"downloader_gochat/model"
	"downloader_gochat/pkg/geoip"
	"downloader_gochat/util"
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
}

type UserService struct {
	userRepo repository.IUserRepository
	timeout  time.Duration
}

func NewUserService(userRepo repository.IUserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
		timeout:  time.Duration(2) * time.Second,
	}
}

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
		Role:                   model.USER,
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

	result, err := s.userRepo.AddUser(&user)
	if err != nil {
		return nil, err
	}

	token, err := util.CreateJwtToken(result.UserId, result.Username, "user")
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

	userVM := model.UserViewModel{
		UserId:   result.UserId,
		Username: result.Username,
		Email:    result.Email,
		Token: model.TokenViewModel{
			AccessToken:       token.AccessToken,
			AccessTokenExpire: token.ExpiresAt,
			RefreshToken:      token.RefreshToken,
		},
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

	token, err := util.CreateJwtToken(searchResult.UserId, searchResult.Username, "user")
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
	}

	return &userVM, nil
}

func (s *UserService) GetToken(deviceVM *model.DeviceInfo, prevRefreshToken string, jwtUserData *util.MyJwtClaims, addProfileImages bool, ip string) (*model.UserViewModel, *util.TokenDetail, error) {
	token, err := util.CreateJwtToken(jwtUserData.UserId, jwtUserData.Username, "user")
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
	err = redis.SetRedis(c.Context(), "jwtKey:"+prevRefreshToken, "logout", time.Duration(remainingTime)*time.Millisecond)
	if err != nil {
		return err
	}

	return nil
}
