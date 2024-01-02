package repository

import (
	"downloader_gochat/model"
	"errors"

	"gorm.io/gorm"
)

type IUserRepository interface {
	CreateUser(user *model.User) (*model.User, error)
	GetDetailUser(int) (*model.UserDataModel, error)
	GetDetailUserByEmail(email string) (*model.UserDataModel, error)
	GetUserByUsernameEmail(username string, email string) (*model.UserDataModel, error)
	GetAllUser() ([]model.UserDataModel, error)
	UpdateUser(*model.User) (*model.User, error)
	DeleteUser(int) error
}

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

//------------------------------------------

func (r *UserRepository) CreateUser(user *model.User) (*model.User, error) {
	err := r.db.Create(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetDetailUser(id int) (*model.UserDataModel, error) {
	var userDataModel model.UserDataModel
	err := r.db.Where("id = ?", id).Model(&model.User{}).Limit(1).Find(&userDataModel).Error
	if err != nil {
		return nil, err
	}

	return &userDataModel, nil
}

func (r *UserRepository) GetDetailUserByEmail(email string) (*model.UserDataModel, error) {
	var userDataModel model.UserDataModel
	err := r.db.Where("email = ?", email).Model(&model.User{}).Limit(1).Find(&userDataModel).Error
	if err != nil {
		return nil, err
	}

	return &userDataModel, nil
}

func (r *UserRepository) GetUserByUsernameEmail(username string, email string) (*model.UserDataModel, error) {
	var userDataModel model.UserDataModel
	err := r.db.Where("username = ? OR email = ?", username, email).Model(&model.User{}).Limit(1).Find(&userDataModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &userDataModel, nil
}

func (r *UserRepository) GetAllUser() ([]model.UserDataModel, error) {
	var users []model.UserDataModel
	err := r.db.Order("id desc").Model(&model.User{}).Limit(100).Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) UpdateUser(user *model.User) (*model.User, error) {
	err := r.db.Save(&user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) DeleteUser(id int) error {
	var user model.User
	err := r.db.Where("id = ?", id).Delete(&user).Error
	if err != nil {
		return err
	}

	return nil
}
