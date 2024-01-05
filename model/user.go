package model

import (
	"downloader_gochat/util"
	"time"

	"github.com/badoux/checkmail"
)

type User struct {
	Bio                            string    `gorm:"column:bio;type:text;not null;default:\"\";"`
	DefaultProfile                 string    `gorm:"column:defaultProfile;type:text;not null;default:\"\";"`
	EmailVerified                  bool      `gorm:"column:emailVerified;type:boolean;not null;default:false;"`
	FavoriteGenres                 []string  `gorm:"column:favoriteGenres;type:text[];"`
	Password                       string    `gorm:"column:password;type:text;not null;"`
	PublicName                     string    `gorm:"column:publicName;type:text;not null;"`
	RawUsername                    string    `gorm:"column:rawUsername;type:text;not null;"`
	RegistrationDate               time.Time `gorm:"column:registrationDate;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	Role                           UserRole  `gorm:"column:role;type:\"userRole\";not null;default:\"user\";"`
	MbtiType                       MbtiType  `gorm:"column:mbtiType;type:\"MbtiType\";"`
	ComputedStatsLastUpdate        int64     `gorm:"column:ComputedStatsLastUpdate;type:bigint;not null;default:0;"`
	EmailVerifyToken               string    `gorm:"column:emailVerifyToken;type:text;not null;default:\"\";"`
	EmailVerifyTokenExpire         int64     `gorm:"column:emailVerifyToken_expire;type:bigint;not null;default:0;"`
	DeleteAccountVerifyToken       string    `gorm:"column:deleteAccountVerifyToken;type:text;not null;default:'';"`
	DeleteAccountVerifyTokenExpire int64     `gorm:"column:deleteAccountVerifyToken_expire;type:bigint;not null;default:0;"`
	//-----------------------------------
	//-----------------------------------
	UserId   int64  `gorm:"column:userId;type:serial;autoIncrement;primaryKey;uniqueIndex:User_userId_key;"`
	Username string `gorm:"column:username;type:text;not null;uniqueIndex:User_username_key;"`
	Email    string `gorm:"column:email;type:text;not null;uniqueIndex:User_email_key;"`
	//-----------------------------------
	//-----------------------------------
	Followers              []Follow                 `gorm:"foreignKey:FollowerId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Following              []Follow                 `gorm:"foreignKey:FollowingId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ProfileImages          []ProfileImage           `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ActiveSessions         []ActiveSession          `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ComputedFavoriteGenres []ComputedFavoriteGenres `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	DownloadLinksSettings  DownloadLinksSettings    `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	NotificationSettings   NotificationSettings     `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	MovieSettings          MovieSettings            `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	FavoriteCharacters     []FavoriteCharacter      `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	LikeDislikeCharacter   []LikeDislikeCharacter   `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	FollowStaff            []FollowStaff            `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	LikeDislikeStaff       []LikeDislikeStaff       `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	FollowMovies           []FollowMovie            `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	LikeDislikeMovies      []LikeDislikeMovie       `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	WatchedMovies          []WatchedMovie           `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	WatchListGroup         []WatchListGroup         `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	WatchListMovies        []WatchListMovie         `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserCollection         []UserCollection         `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserCollectionMovie    []UserCollectionMovie    `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (User) TableName() string {
	return "User"
}

//---------------------------------------
//---------------------------------------

type UserViewModel struct {
	UserId   int64  `json:"UserId" db:"UserId"`
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

//---------------------------------------
//---------------------------------------

type UserDataModel struct {
	UserId   int64
	Username string
	Email    string
	Password string
}

//---------------------------------------
//---------------------------------------

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
