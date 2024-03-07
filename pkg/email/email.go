package email

import (
	"context"
	"downloader_gochat/configs"
	"downloader_gochat/db/mongodb"
	"downloader_gochat/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func AddRegisterEmail(email string, emailVerifyToken string, rawUsername string, delaySec int) error {
	host := configs.GetConfigs().MainServerAddress
	emailDoc := bson.D{
		{"name", "registration email"},
		{"data", bson.D{
			{"rawUsername", rawUsername},
			{"email", email},
			{"emailVerifyToken", emailVerifyToken},
			{"host", host},
		}},
		{"priority", 0},
		{"shouldSaveResult", false},
		{"type", "normal"},
		{"nextRunAt", time.Now().Add(time.Duration(delaySec) * time.Second).UTC()},
		{"lastModifiedBy", nil},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := mongodb.MONGODB.Db.Collection(configs.GetConfigs().AgendaJobsCollection).InsertOne(ctx, emailDoc)
	if err != nil {
		return err
	}
	return nil
}

func AddLoginEmail(email string, deviceInfo *model.DeviceInfo, delaySec int) error {
	emailDoc := bson.D{
		{"name", "login email"},
		{"data", bson.D{
			{"deviceInfo", deviceInfo},
			{"email", email},
		}},
		{"priority", 0},
		{"shouldSaveResult", false},
		{"type", "normal"},
		{"nextRunAt", time.Now().Add(time.Duration(delaySec) * time.Second).UTC()},
		{"lastModifiedBy", nil},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := mongodb.MONGODB.Db.Collection(configs.GetConfigs().AgendaJobsCollection).InsertOne(ctx, emailDoc)
	if err != nil {
		return err
	}
	return nil
}

func AddUpdatePasswordEmail(email string, delaySec int) error {
	emailDoc := bson.D{
		{"name", "login email"},
		{"data", bson.D{
			{"email", email},
		}},
		{"priority", 0},
		{"shouldSaveResult", false},
		{"type", "normal"},
		{"nextRunAt", time.Now().Add(time.Duration(delaySec) * time.Second).UTC()},
		{"lastModifiedBy", nil},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := mongodb.MONGODB.Db.Collection(configs.GetConfigs().AgendaJobsCollection).InsertOne(ctx, emailDoc)
	if err != nil {
		return err
	}
	return nil
}

func AddVerifyEmail(userId int64, email string, emailVerifyToken string, rawUsername string, delaySec int) error {
	host := configs.GetConfigs().MainServerAddress
	emailDoc := bson.D{
		{"name", "verify email"},
		{"data", bson.D{
			{"userId", userId},
			{"rawUsername", rawUsername},
			{"email", email},
			{"emailVerifyToken", emailVerifyToken},
			{"host", host},
		}},
		{"priority", 0},
		{"shouldSaveResult", false},
		{"type", "normal"},
		{"nextRunAt", time.Now().Add(time.Duration(delaySec) * time.Second).UTC()},
		{"lastModifiedBy", nil},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := mongodb.MONGODB.Db.Collection(configs.GetConfigs().AgendaJobsCollection).InsertOne(ctx, emailDoc)
	if err != nil {
		return err
	}
	return nil
}
