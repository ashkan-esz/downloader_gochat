package service

import (
	"context"
	"downloader_gochat/db/redis"
	"downloader_gochat/model"
	"encoding/json"
	"fmt"
	"strconv"
)

type ICacheService interface {
	getCachedUserData(userId int64) (*model.CachedUserData, error)
	setUserDataCache(userId int64, userData *model.CachedUserData) error
}

const (
	userDataCachePrefix = "user:"
)

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

func setUserDataCache(userId int64, userData *model.CachedUserData) error {
	jsonData, err := json.Marshal(userData)
	if err != nil {
		fmt.Println("Redis Error on saving userData: ", err)
		return err
	}
	err = redis.SetRedis(context.Background(), userDataCachePrefix+strconv.FormatInt(userId, 10), jsonData, 0)
	if err != nil {
		fmt.Println("Redis Error on saving userData: ", err)
	}
	return err
}
