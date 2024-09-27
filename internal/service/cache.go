package service

import (
	"context"
	"downloader_gochat/db/redis"
	"downloader_gochat/model"
	errorHandler "downloader_gochat/pkg/error"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

type ICacheService interface {
	GetJwtDataCache(key string) (string, error)
	setJwtDataCache(key string, value string, duration time.Duration) error
	getRolePermissionsCache(roleIds []int64) ([]string, error)
	setRolePermissionsCache(roleIds []int64, permissions []string, duration time.Duration) error
	removeNotifTokenFromcachedUserData(userId int64, notifToken string) error
	addNotifTokenToCachedUserData(userId int64, notifToken string) error
	updateProfileImageOfCachedUserData(userId int64, profileImages *[]model.ProfileImageDataModel) error
	updateNotificationSettingsOfCachedUserData(userId int64, settings model.NotificationSettings) error
	updateProfileDataOfCachedUserData(userId int64, username string, publicName string) error
	getCachedUserData(userId int64) (*model.CachedUserData, error)
	setUserDataCache(userId int64, userData *model.CachedUserData) error
	getCachedMovieData(movieId string) (*model.CachedMovieData, error)
	getCachedMultiMovieData(movieIds []string) ([]model.CachedMovieData, error)
	setMovieDataCache(movieId string, movieData *model.CachedMovieData) error
}

const (
	jwtDataCachePrefix         = "jwtKey:"
	userDataCachePrefix        = "user:"
	movieDataCachePrefix       = "movie:"
	botDataCachePrefix         = "bot:"
	rolePermissionsCachePrefix = "roleIds:"
)

//------------------------------------------
//------------------------------------------

func GetJwtDataCache(key string) (string, error) {
	result, err := redis.GetRedis(context.Background(), jwtDataCachePrefix+key)
	return result, err
}

func setJwtDataCache(key string, value string, duration time.Duration) error {
	err := redis.SetRedis(context.Background(), jwtDataCachePrefix+key, value, duration)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving jwt: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}
	return err
}

//------------------------------------------
//------------------------------------------

func getRolePermissionsCache(roleIds []int64) ([]string, error) {
	key := int64SliceToString(roleIds, ",")
	result, err := redis.GetRedis(context.Background(), rolePermissionsCachePrefix+key)
	if err != nil && err.Error() != "redis: nil" {
		return nil, nil
	}
	if result != "" {
		var jsonData []string
		err = json.Unmarshal([]byte(result), &jsonData)
		if err != nil {
			return nil, err
		}
		return jsonData, nil
	}
	return nil, err
}

func setRolePermissionsCache(roleIds []int64, permissions []string, duration time.Duration) error {
	jsonData, err := json.Marshal(permissions)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving permissions: %v", err)
		errorHandler.SaveError(errorMessage, err)
		return err
	}

	key := int64SliceToString(roleIds, ",")
	err = redis.SetRedis(context.Background(), rolePermissionsCachePrefix+key, jsonData, duration)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving permissions: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}
	return err
}

//------------------------------------------
//------------------------------------------

func removeNotifTokenFromCachedUserData(userId int64, notifToken string) error {
	cacheData, err := getCachedUserData(userId)
	if err != nil {
		return err
	}
	if cacheData == nil {
		return nil
	}

	cacheData.NotifTokens = slices.DeleteFunc(cacheData.NotifTokens, func(val string) bool {
		return val == notifToken
	})
	err = setUserDataCache(userId, cacheData)

	return err
}

func addNotifTokenToCachedUserData(userId int64, notifToken string) error {
	cacheData, err := getCachedUserData(userId)
	if err != nil {
		return err
	}
	if cacheData == nil {
		return nil
	}

	cacheData.NotifTokens = append(cacheData.NotifTokens, notifToken)
	cacheData.NotifTokens = slices.Compact(cacheData.NotifTokens)
	err = setUserDataCache(userId, cacheData)

	return err
}

func updateProfileImageOfCachedUserData(userId int64, profileImages *[]model.ProfileImageDataModel) error {
	cacheData, err := getCachedUserData(userId)
	if err != nil {
		return err
	}
	if cacheData == nil {
		return nil
	}

	cacheData.ProfileImages = *profileImages
	err = setUserDataCache(userId, cacheData)

	return err
}

func updateNotificationSettingsOfCachedUserData(userId int64, settings model.NotificationSettings) error {
	cacheData, err := getCachedUserData(userId)
	if err != nil {
		return err
	}
	if cacheData == nil {
		return nil
	}

	cacheData.NotificationSettings = settings
	err = setUserDataCache(userId, cacheData)

	return err
}

func updateProfileDataOfCachedUserData(userId int64, username string, publicName string) error {
	cacheData, err := getCachedUserData(userId)
	if err != nil {
		return err
	}
	if cacheData == nil {
		return nil
	}

	cacheData.Username = username
	cacheData.PublicName = publicName
	err = setUserDataCache(userId, cacheData)

	return err
}

//------------------------------------------
//------------------------------------------

func getCachedUserData(userId int64) (*model.CachedUserData, error) {
	result, err := redis.GetRedis(context.Background(), userDataCachePrefix+strconv.FormatInt(userId, 10))
	if err != nil && err.Error() != "redis: nil" {
		return nil, nil
	}
	if result != "" {
		var jsonData model.CachedUserData
		err = json.Unmarshal([]byte(result), &jsonData)
		if err != nil {
			return nil, err
		}
		return &jsonData, nil
	}
	return nil, err
}

func getCachedMultiUserData(userIds []int64) ([]model.CachedUserData, error) {
	keys := make([]string, len(userIds))
	for i, id := range userIds {
		keys[i] = userDataCachePrefix + strconv.FormatInt(id, 10)
	}

	result, err := redis.MGetRedis(context.Background(), keys)
	if err != nil && err.Error() != "redis: nil" {
		return nil, nil
	}
	if result != nil {
		cachedData := []model.CachedUserData{}
		for i := range result {
			if result[i] == nil {
				//not found
				continue
			}
			var jsonData model.CachedUserData
			err = json.Unmarshal([]byte(result[i].(string)), &jsonData)
			if err != nil {
				return nil, err
			}
			cachedData = append(cachedData, jsonData)
		}
		return cachedData, nil
	}
	return nil, err
}

func setUserDataCache(userId int64, userData *model.CachedUserData) error {
	jsonData, err := json.Marshal(userData)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving userData: %v", err)
		errorHandler.SaveError(errorMessage, err)
		return err
	}
	err = redis.SetRedis(context.Background(), userDataCachePrefix+strconv.FormatInt(userId, 10), jsonData, 24*time.Hour)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving userData: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}
	return err
}

//------------------------------------------
//------------------------------------------

func getCachedMovieData(movieId string) (*model.CachedMovieData, error) {
	result, err := redis.GetRedis(context.Background(), movieDataCachePrefix+movieId)
	if err != nil && err.Error() != "redis: nil" {
		return nil, nil
	}
	if result != "" {
		var jsonData model.CachedMovieData
		err = json.Unmarshal([]byte(result), &jsonData)
		if err != nil {
			return nil, err
		}
		return &jsonData, nil
	}
	return nil, err
}

func getCachedMultiMovieData(movieIds []string) ([]model.CachedMovieData, error) {
	keys := make([]string, len(movieIds))
	for i, id := range movieIds {
		keys[i] = movieDataCachePrefix + id
	}

	result, err := redis.MGetRedis(context.Background(), keys)
	if err != nil && err.Error() != "redis: nil" {
		return nil, nil
	}
	if result != nil {
		cachedData := []model.CachedMovieData{}
		for i := range result {
			if result[i] == nil {
				//not found
				continue
			}
			var jsonData model.CachedMovieData
			err = json.Unmarshal([]byte(result[i].(string)), &jsonData)
			if err != nil {
				return nil, err
			}
			cachedData = append(cachedData, jsonData)
		}
		return cachedData, nil
	}
	return nil, err
}

func setMovieDataCache(movieId string, movieData *model.CachedMovieData) error {
	jsonData, err := json.Marshal(movieData)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving movieData: %v", err)
		errorHandler.SaveError(errorMessage, err)
		return err
	}
	err = redis.SetRedis(context.Background(), movieDataCachePrefix+movieId, jsonData, 1*time.Hour)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving movieData: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}
	return err
}

//------------------------------------------
//------------------------------------------

func getCachedBotData(botId string) (*model.Bot, error) {
	result, err := redis.GetRedis(context.Background(), botDataCachePrefix+botId)
	if err != nil && err.Error() != "redis: nil" {
		return nil, nil
	}
	if result != "" {
		var jsonData model.Bot
		err = json.Unmarshal([]byte(result), &jsonData)
		if err != nil {
			return nil, err
		}
		return &jsonData, nil
	}
	return nil, err
}

func getCachedMultiBotData(botIds []string) ([]model.Bot, error) {
	keys := make([]string, len(botIds))
	for i, id := range botIds {
		keys[i] = botDataCachePrefix + id
	}

	result, err := redis.MGetRedis(context.Background(), keys)
	if err != nil && err.Error() != "redis: nil" {
		return nil, nil
	}
	if result != nil {
		cachedData := []model.Bot{}
		for i := range result {
			if result[i] == nil {
				//not found
				continue
			}
			var jsonData model.Bot
			err = json.Unmarshal([]byte(result[i].(string)), &jsonData)
			if err != nil {
				return nil, err
			}
			cachedData = append(cachedData, jsonData)
		}
		return cachedData, nil
	}
	return nil, err
}

func setBotDataCache(botId string, botData *model.Bot) error {
	jsonData, err := json.Marshal(botData)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving botData: %v", err)
		errorHandler.SaveError(errorMessage, err)
		return err
	}
	err = redis.SetRedis(context.Background(), botDataCachePrefix+botId, jsonData, 1*time.Hour)
	if err != nil {
		errorMessage := fmt.Sprintf("Redis Error on saving botData: %v", err)
		errorHandler.SaveError(errorMessage, err)
	}
	return err
}

//------------------------------------------
//------------------------------------------

func int64SliceToString(nums []int64, delimiter string) string {
	// Create a string slice to hold the converted numbers
	strNums := make([]string, len(nums))

	// Convert each int64 to string
	for i, num := range nums {
		strNums[i] = strconv.FormatInt(num, 10)
	}

	// Join the strings with the specified delimiter
	return strings.Join(strNums, delimiter)
}
