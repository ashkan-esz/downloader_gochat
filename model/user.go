package model

import (
	"downloader_gochat/util"
	"regexp"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/lib/pq"
)

type User struct {
	Bio                            string         `gorm:"column:bio;type:text;not null;default:\"\";"`
	DefaultProfile                 string         `gorm:"column:defaultProfile;type:text;not null;default:\"\";"`
	EmailVerified                  bool           `gorm:"column:emailVerified;type:boolean;not null;default:false;"`
	FavoriteGenres                 pq.StringArray `gorm:"column:favoriteGenres;type:text[];" swaggertype:"array,string"`
	Password                       string         `gorm:"column:password;type:text;not null;"`
	PublicName                     string         `gorm:"column:publicName;type:text;not null;"`
	RawUsername                    string         `gorm:"column:rawUsername;type:text;not null;"`
	RegistrationDate               time.Time      `gorm:"column:registrationDate;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	Role                           UserRole       `gorm:"column:role;type:\"userRole\";not null;default:\"user\";"`
	MbtiType                       MbtiType       `gorm:"column:mbtiType;type:\"MbtiType\";"`
	ComputedStatsLastUpdate        int64          `gorm:"column:ComputedStatsLastUpdate;type:bigint;not null;default:0;"`
	EmailVerifyToken               string         `gorm:"column:emailVerifyToken;type:text;not null;default:\"\";"`
	EmailVerifyTokenExpire         int64          `gorm:"column:emailVerifyToken_expire;type:bigint;not null;default:0;"`
	DeleteAccountVerifyToken       string         `gorm:"column:deleteAccountVerifyToken;type:text;not null;default:'';"`
	DeleteAccountVerifyTokenExpire int64          `gorm:"column:deleteAccountVerifyToken_expire;type:bigint;not null;default:0;"`
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
	CreatedRooms           []Room                   `gorm:"foreignKey:CreatorId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ReceiverRooms          []Room                   `gorm:"foreignKey:ReceiverId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SendedMessages         []Message                `gorm:"foreignKey:CreatorId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ReceivedMessages       []Message                `gorm:"foreignKey:ReceiverId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserMessageRead        UserMessageRead          `gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	createdNotifications   []Notification           `gorm:"foreignKey:CreatorId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	receivedNotifications  []Notification           `gorm:"foreignKey:ReceiverId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (User) TableName() string {
	return "User"
}

//---------------------------------------
//---------------------------------------

type UserViewModel struct {
	UserId        int64          `json:"userId"`
	Username      string         `json:"username"`
	Email         string         `json:"email,omitempty"`
	Token         TokenViewModel `json:"token"`
	ProfileImages []ProfileImage `json:"profileImages,omitempty"`
}

type TokenViewModel struct {
	AccessToken       string `json:"accessToken"`
	AccessTokenExpire int64  `json:"accessToken_expire"`
	RefreshToken      string `json:"refreshToken,omitempty"`
}

type RegisterViewModel struct {
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Password        string     `json:"password"`
	ConfirmPassword string     `json:"confirmPassword"`
	DeviceInfo      DeviceInfo `json:"deviceInfo"`
}

type LoginViewModel struct {
	Email      string     `json:"username_email"`
	Password   string     `json:"password"`
	DeviceInfo DeviceInfo `json:"deviceInfo"`
}

type DeviceInfo struct {
	AppName     string `json:"appName"`
	AppVersion  string `json:"appVersion"`
	Os          string `json:"os"`
	DeviceModel string `json:"deviceModel"`
	NotifToken  string `json:"notifToken"`
	Fingerprint string `json:"fingerprint"`
}

//---------------------------------------
//---------------------------------------

type UserDataModel struct {
	UserId   int64  `db:"userId" gorm:"column:userId" json:"userId"`
	Username string `db:"username" gorm:"column:username" json:"username"`
	Email    string `db:"email" gorm:"column:email" json:"email"`
	Password string `db:"password" gorm:"column:password" json:"-"`
}

type UserWithImageDataModel struct {
	UserId        int64          `db:"userId" gorm:"column:userId" json:"userId"`
	Username      string         `db:"username" gorm:"column:username" json:"username"`
	Email         string         `db:"email" gorm:"column:email" json:"email"`
	Password      string         `db:"password" gorm:"column:password" json:"-"`
	ProfileImages []ProfileImage `db:"profileImages" gorm:"foreignKey:UserId;references:UserId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"profileImages"`
}

type UserMetaWithNotificationSettings struct {
	UserId         int64           `gorm:"column:userId" json:"userId"`
	Username       string          `gorm:"column:username" json:"username"`
	PublicName     string          `gorm:"column:publicName" json:"publicName"`
	ProfileImages  []ProfileImage  `gorm:"foreignKey:UserId;references:UserId;" json:"profileImages"`
	ActiveSessions []ActiveSession `gorm:"foreignKey:UserId;references:UserId;" json:"activeSessions"`
	//NotificationSettings NotificationSettings `gorm:"foreignKey:UserId;references:UserId;" json:"notificationSettings"`
	NewFollower               bool `gorm:"column:newFollower;" json:"newFollower"`
	NewMessage                bool `gorm:"column:newMessage;" json:"newMessage"`
	FinishedListSpinOffSequel bool `gorm:"column:finishedList_spinOffSequel;" json:"finishedListSpinOffSequel"`
	FollowMovie               bool `gorm:"column:followMovie;" json:"followMovie"`
	FollowMovieBetterQuality  bool `gorm:"column:followMovie_betterQuality;type:boolean;not null;" json:"followMovieBetterQuality"`
	FollowMovieSubtitle       bool `gorm:"column:followMovie_subtitle;" json:"followMovieSubtitle"`
	FutureList                bool `gorm:"column:futureList;" json:"futureList"`
	FutureListSerialSeasonEnd bool `gorm:"column:futureList_serialSeasonEnd;" json:"futureListSerialSeasonEnd"`
	FutureListSubtitle        bool `gorm:"column:futureList_subtitle;" json:"futureListSubtitle"`
}

type UserMetaDataModel struct {
	UserId     int64  `db:"userId" gorm:"column:userId" json:"userId"`
	Username   string `db:"username" gorm:"column:username" json:"username"`
	PublicName string `db:"publicName" gorm:"column:publicName" json:"publicName"`
}

type UserMetaWithImageDataModel struct {
	UserId        int64          `gorm:"column:userId" json:"userId"`
	Username      string         `gorm:"column:username" json:"username"`
	PublicName    string         `gorm:"column:publicName" json:"publicName"`
	ProfileImages []ProfileImage `gorm:"foreignKey:UserId;references:UserId;" json:"profileImages"`
}

//---------------------------------------
//---------------------------------------

type UserProfileReq struct {
	UserId                     int64  `json:"userId"`
	IsSelfProfile              bool   `json:"isSelfProfile"`
	LoadSettings               bool   `json:"loadSettings"`
	LoadFollowersCount         bool   `json:"loadFollowersCount"`
	LoadProfileImages          bool   `json:"loadProfileImages"`
	LoadComputedFavoriteGenres bool   `json:"loadComputedFavoriteGenres"`
	RefreshToken               string `json:"refreshToken"`
}

type UserProfileRes struct {
	UserId                  int64                             `gorm:"column:userId" json:"userId"`
	Username                string                            `gorm:"column:username" json:"username"`
	RawUsername             string                            `gorm:"column:rawUsername" json:"rawUsername"`
	PublicName              string                            `gorm:"column:publicName" json:"publicName"`
	Email                   string                            `gorm:"column:email" json:"email"`
	EmailVerified           bool                              `gorm:"column:emailVerified" json:"emailVerified"`
	Bio                     string                            `gorm:"column:bio" json:"bio"`
	RegistrationDate        time.Time                         `gorm:"column:registrationDate;" json:"registrationDate"`
	DefaultProfile          string                            `gorm:"column:defaultProfile" json:"defaultProfile"`
	ComputedStatsLastUpdate int64                             `gorm:"column:ComputedStatsLastUpdate;" json:"computedStatsLastUpdate"`
	FavoriteGenres          pq.StringArray                    `gorm:"column:favoriteGenres;type:text[];" json:"favoriteGenres" swaggertype:"array,string"`
	Role                    UserRole                          `gorm:"column:role;" json:"role"`
	MbtiType                MbtiType                          `gorm:"column:mbtiType" json:"mbtiType"`
	ProfileImages           []FollowListProfileImageDataModel `gorm:"foreignKey:UserId;references:UserId;" json:"profileImages"`
	ComputedFavoriteGenres  []ComputedFavoriteGenres          `gorm:"foreignKey:UserId;references:UserId;" json:"computedFavoriteGenres"`
	NotificationSettings    *NotificationSettings             `gorm:"foreignKey:UserId;references:UserId;" json:"notificationSettings"`
	DownloadLinksSettings   *DownloadLinksSettings            `gorm:"foreignKey:UserId;references:UserId;" json:"downloadLinksSettings"`
	MovieSettings           *MovieSettings                    `gorm:"foreignKey:UserId;references:UserId;" json:"MovieSettings"`
	ThisDevice              *ActiveSessionDataModel           `gorm:"foreignKey:UserId;references:UserId;" json:"thisDevice"`
	FollowersCount          int64                             `gorm:"-" json:"followersCount"`
	FollowingsCount         int64                             `gorm:"-" json:"followingsCount"`
}

//---------------------------------------
//---------------------------------------

func (u *User) EncryptPassword(password string) error {
	hashPassword, err := util.HashPassword(password)
	if err != nil {
		return err
	}
	u.Password = hashPassword

	return nil
}

func (u *User) EncryptEmailToken(token string) error {
	hashToken, err := util.HashPassword(token)
	if err != nil {
		return err
	}
	u.EmailVerifyToken = strings.ReplaceAll(hashToken, "/", "")

	return nil
}

func (u *LoginViewModel) CheckPassword(password string, hashedPassword string) error {
	err := util.CheckPassword(password, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

//---------------------------------------
//---------------------------------------

func (u *RegisterViewModel) Validate() []string {
	errors := make([]string, 0)

	if u.Username == "" {
		errors = append(errors, "Username Is Empty")
	} else {
		if len(u.Username) < 6 {
			errors = append(errors, "Username Length Must Be More Than 6")
		} else if len(u.Username) > 50 {
			errors = append(errors, "Username Length Must Be Less Than 50")
		}
		if matched, _ := regexp.MatchString("(?i)^[a-z|\\d_-]+$", u.Username); !matched {
			errors = append(errors, "Only a-z, 0-9, and underscores are allowed")
		}
	}

	if u.Password == "" {
		errors = append(errors, "Password Is Empty")
	} else {
		if len(u.Password) < 8 {
			errors = append(errors, "Password Length Must Be More Than 8")
		} else if len(u.Password) > 50 {
			errors = append(errors, "Password Length Must Be Less Than 50")
		}
		if matched, _ := regexp.MatchString("[0-9]", u.Password); !matched {
			errors = append(errors, "Password Must Contain A Number")
		}
		if matched, _ := regexp.MatchString("[A-Z]", u.Password); !matched {
			errors = append(errors, "Password Must Contain An Uppercase")
		}
		if strings.Contains(u.Password, " ") {
			errors = append(errors, "Password Cannot Have Space")
		}
		if u.Username == u.Password {
			errors = append(errors, "Password Is Equal With Username")
		}
		if u.Password != u.ConfirmPassword {
			errors = append(errors, "Passwords Don't Match")
		}
	}

	if u.Email == "" {
		errors = append(errors, "Email Is Empty")
	} else {
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			errors = append(errors, "Email Is in Wrong Format")
		}
	}

	deviceInfoError := u.DeviceInfo.Validate()
	errors = append(errors, deviceInfoError...)

	return errors
}

func (u *RegisterViewModel) Normalize() *RegisterViewModel {
	u.Username = strings.TrimSpace(u.Username)
	u.Password = strings.TrimSpace(u.Password)
	u.Email = strings.TrimSpace(u.Email)
	u.ConfirmPassword = strings.TrimSpace(u.ConfirmPassword)
	u.DeviceInfo.Normalize()

	return u
}

//---------------------------------------
//---------------------------------------

func (u *LoginViewModel) Validate() []string {
	errors := make([]string, 0)

	if u.Email == "" {
		errors = append(errors, "Username Is Empty")
	} else {
		if len(u.Email) < 6 {
			errors = append(errors, "Username Length Must Be More Than 6")
		} else if len(u.Email) > 50 {
			errors = append(errors, "Username Length Must Be Less Than 50")
		}
	}

	if u.Password == "" {
		errors = append(errors, "Password Is Empty")
	} else {
		if len(u.Password) < 8 {
			errors = append(errors, "Password Length Must Be More Than 8")
		} else if len(u.Password) > 50 {
			errors = append(errors, "Password Length Must Be Less Than 50")
		}
		if u.Email == u.Password {
			errors = append(errors, "Password Is Equal With Username")
		}
	}

	deviceInfoError := u.DeviceInfo.Validate()
	errors = append(errors, deviceInfoError...)

	return errors
}

func (u *LoginViewModel) Normalize() *LoginViewModel {
	u.Password = strings.TrimSpace(u.Password)
	u.Email = strings.TrimSpace(u.Email)
	u.DeviceInfo.Normalize()

	return u
}

//---------------------------------------
//---------------------------------------

func (d *DeviceInfo) Validate() []string {
	errors := make([]string, 0)

	if d.AppName == "" {
		errors = append(errors, "Missed parameter deviceInfo.appName")
	}
	if d.AppVersion == "" {
		errors = append(errors, "Missed parameter deviceInfo.appVersion")
	} else if matched, _ := regexp.MatchString("^\\d\\d?\\.\\d\\d?\\.\\d\\d?$", d.AppVersion); !matched {
		errors = append(errors, "Invalid parameter deviceInfo.appVersion")
	}
	if d.Os == "" {
		errors = append(errors, "Missed parameter deviceInfo.os")
	}
	if d.DeviceModel == "" {
		errors = append(errors, "Missed parameter deviceInfo.deviceModel")
	}

	return errors
}

func (d *DeviceInfo) Normalize() *DeviceInfo {
	d.AppName = strings.TrimSpace(d.AppName)
	d.AppVersion = strings.TrimSpace(d.AppVersion)
	d.Os = strings.TrimSpace(d.Os)
	d.DeviceModel = strings.TrimSpace(d.DeviceModel)

	return d
}
