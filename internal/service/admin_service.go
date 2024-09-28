package service

import (
	"downloader_gochat/cloudStorage"
	"downloader_gochat/internal/repository"
	"downloader_gochat/model"
	"downloader_gochat/rabbitmq"
	"time"
)

type IAdminService interface {
	GetServerStatus() *model.Status
}

type AdminService struct {
	userRepo     repository.IUserRepository
	adminRepo    repository.IAdminRepository
	rabbitmq     rabbitmq.RabbitMQ
	cloudStorage cloudStorage.IS3Storage
	timeout      time.Duration
	status       *model.Status
}

func NewAdminService(userRepo repository.IUserRepository, adminRepo repository.IAdminRepository, rabbit rabbitmq.RabbitMQ, cloudStorage cloudStorage.IS3Storage) *AdminService {
	svc := &AdminService{
		userRepo:     userRepo,
		adminRepo:    adminRepo,
		rabbitmq:     rabbit,
		cloudStorage: cloudStorage,
		timeout:      time.Duration(2) * time.Second,
		status: &model.Status{
			Tasks: &model.Tasks{},
		},
	}

	AdminSvc = svc

	return svc
}

var AdminSvc *AdminService

//------------------------------------------
//------------------------------------------

func (a *AdminService) GetServerStatus() *model.Status {
	return a.status
}
