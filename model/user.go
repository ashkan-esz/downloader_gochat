package model

import (
	"downloader_gochat/util"

	"github.com/badoux/checkmail"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID       int64  `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type UserViewModel struct {
	ID       int64  `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
}

type RegisterViewModel struct {
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

type LoginViewModel struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"password" db:"password"`
}

//----------------------------------------

func (u *User) EncryptPassword(password string) (string, error) {
	hashPassword, err := util.HashPassword(password)
	if err != nil {
		return "", err
	}

	return hashPassword, nil
}

func (u *LoginViewModel) CheckPassword(password string, hashedPassword string) error {
	err := util.CheckPassword(password, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

func (u *User) Validate() map[string]string {
	var errorMessages = make(map[string]string)
	var err error

	if u.Email == "" {
		errorMessages["email_required"] = "email required"
	}
	if u.Email != "" {
		if err = checkmail.ValidateFormat(u.Email); err != nil {
			errorMessages["invalid_email"] = "email email"
		}
	}

	return errorMessages
}

func (u *LoginViewModel) Validate() map[string]string {
	var errorMessages = make(map[string]string)
	var err error

	if u.Password == "" {
		errorMessages["password_required"] = "password is required"
	}
	if u.Email == "" {
		errorMessages["email_required"] = "email is required"
	}
	if u.Email != "" {
		if err = checkmail.ValidateFormat(u.Email); err != nil {
			errorMessages["invalid_email"] = "please provide a valid email"
		}
	}

	return errorMessages
}

func (u *RegisterViewModel) Validate() map[string]string {
	var errorMessages = make(map[string]string)
	var err error

	if u.Username == "" {
		errorMessages["username_required"] = "username is required"
	}
	if u.Username != "" && len(u.Username) < 4 {
		errorMessages["username_password"] = "username should be at least 4 characters"
	}
	if u.Password == "" {
		errorMessages["password_required"] = "password is required"
	}
	if u.Password != "" && len(u.Password) < 6 {
		errorMessages["invalid_password"] = "password should be at least 6 characters"
	}
	if u.Email == "" {
		errorMessages["email_required"] = "email is required"
	}
	if u.Email != "" {
		if err = checkmail.ValidateFormat(u.Email); err != nil {
			errorMessages["invalid_email"] = "please provide a valid email"
		}
	}

	return errorMessages
}
