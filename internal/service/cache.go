package service

import (
	"context"
	"downloader_gochat/db/redis"
	"downloader_gochat/model"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"time"
)

type ICacheService interface {
	removeNotifTokenFromcachedUserData(userId int64, notifToken string) error
	addNotifTokenToCachedUserData(userId int64, notifToken string) error
	updateProfileImageOfCachedUserData(userId int64, profileImages *[]model.ProfileImageDataModel) error
	updateNotificationSettingsOfCachedUserData(userId int64, settings model.NotificationSettings) error
	updateProfileDataOfCachedUserData(userId int64, username string, publicName string) error
	getCachedUserData(userId int64) (*model.CachedUserData, error)
	setUserDataCache(userId int64, userData *model.CachedUserData) error
}

const (
	userDataCachePrefix = "user:"
)

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
		fmt.Println("Redis Error on saving userData: ", err)
		return err
	}
	err = redis.SetRedis(context.Background(), userDataCachePrefix+strconv.FormatInt(userId, 10), jsonData, 24*time.Hour)
	if err != nil {
		fmt.Println("Redis Error on saving userData: ", err)
	}
	return err
}
