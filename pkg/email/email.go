package email

import (
	"downloader_gochat/model"
)

type EmailType string

const (
	UserRegistration EmailType = "registration email"
	UserLogin        EmailType = "login email"
	PasswordUpdated  EmailType = "password updated"
	ResetPassword    EmailType = "reset password"
	VerifyEmail      EmailType = "verify email"
	DeleteAccount    EmailType = "delete account"
)

type EmailQueueData struct {
	Type        EmailType         `json:"type"`
	UserId      int64             `json:"userId"`
	RawUsername string            `json:"rawUsername"`
	Email       string            `json:"email"`
	Token       string            `json:"token"`
	Host        string            `json:"host"`
	Url         string            `json:"url"`
	DeviceInfo  *model.DeviceInfo `json:"deviceInfo"`
	IpLocation  string            `json:"ipLocation"`
}
