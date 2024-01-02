package service

import (
	"downloader_gochat/internal/repository"
	"downloader_gochat/model"
	"time"
)

type IUserService interface {
	CreateUser(registerVM *model.RegisterViewModel) (*model.UserViewModel, error)
	LoginUser(loginVM *model.LoginViewModel) (*model.UserViewModel, error)
	GetListUser() (*[]model.UserViewModel, error)
	GetDetailUser(id int) (*model.UserViewModel, error)
	UpdateUser(userVM *model.User) (*model.UserViewModel, error)
	DeleteUser(id int) error
}

// todo : add/handle
//	_, cancel := context.WithTimeout(c, s.timeout)
//	defer cancel()

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

func (s *UserService) CreateUser(userVM *model.RegisterViewModel) (*model.UserViewModel, error) {
	var user = model.User{
		Username: userVM.Username,
		Email:    userVM.Email,
	}

	searchResult, err := s.userRepo.GetUserByUsernameEmail(userVM.Username, userVM.Email)
	if err != nil {
		return nil, err
	}
	if searchResult != nil {
		return &model.UserViewModel{
			ID:       0,
			Username: searchResult.Username,
			Email:    searchResult.Email,
		}, nil
	}

	password, err := user.EncryptPassword(userVM.Password)
	if err != nil {
		return nil, err
	}

	user.Password = password

	result, err := s.userRepo.CreateUser(&user)
	if err != nil {
		return nil, err
	}

	var afterRegVM model.UserViewModel

	if result != nil {
		afterRegVM = model.UserViewModel{
			ID:       result.ID,
			Username: result.Username,
			Email:    result.Email,
		}
	}

	return &afterRegVM, nil
}

func (s *UserService) LoginUser(loginVM *model.LoginViewModel) (*model.UserViewModel, error) {
	u, err := s.userRepo.GetDetailUserByEmail(loginVM.Email)
	if err != nil {
		return &model.UserViewModel{}, err
	}

	err = loginVM.CheckPassword(loginVM.Password, u.Password)
	if err != nil {
		return &model.UserViewModel{}, err
	}

	return &model.UserViewModel{Username: u.Username, ID: u.ID}, nil
}

func (s *UserService) GetListUser() (*[]model.UserViewModel, error) {
	result, err := s.userRepo.GetAllUser()
	if err != nil {
		return nil, err
	}

	var users []model.UserViewModel
	for _, item := range result {
		user := model.UserViewModel{ID: item.ID, Username: item.Email, Email: item.Email}
		users = append(users, user)
	}

	return &users, nil
}

func (s *UserService) GetDetailUser(id int) (*model.UserViewModel, error) {
	var viewModel model.UserViewModel

	result, err := s.userRepo.GetDetailUser(id)
	if err != nil {
		return nil, err
	}

	if result != nil {
		viewModel = model.UserViewModel{
			ID:       result.ID,
			Username: result.Username,
			Email:    result.Email,
		}
	}

	return &viewModel, nil
}

func (s *UserService) UpdateUser(userVM *model.User) (*model.UserViewModel, error) {
	password, err := userVM.EncryptPassword(userVM.Password)
	if err != nil {
		return nil, err
	}

	userVM.Password = password

	result, err := s.userRepo.UpdateUser(userVM)
	if err != nil {
		return nil, err
	}

	userAfterUpdate := model.UserViewModel{
		ID:       result.ID,
		Username: result.Username,
		Email:    result.Email,
	}

	return &userAfterUpdate, err
}

func (s *UserService) DeleteUser(id int) error {
	err := s.userRepo.DeleteUser(id)
	if err != nil {
		return err
	}

	return nil
}
