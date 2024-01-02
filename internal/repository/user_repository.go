package repository

import (
	"downloader_gochat/model"
	"errors"

	"gorm.io/gorm"
)

type IUserRepository interface {
	CreateUser(user *model.User) (*model.User, error)
	GetDetailUser(int) (*model.User, error)
	GetDetailUserByEmail(email string) (*model.User, error)
	GetUserByUsernameEmail(username string, email string) (*model.User, error)
	GetAllUser() ([]model.User, error)
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

// todo : handle field projection

func (r *UserRepository) CreateUser(user *model.User) (*model.User, error) {
	err := r.db.Create(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetDetailUser(id int) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).Take(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetDetailUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).Take(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByUsernameEmail(username string, email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ? OR email = ?", username, email).Take(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetAllUser() ([]model.User, error) {
	var users []model.User
	err := r.db.Order("id desc").Find(&users).Error
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
