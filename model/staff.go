package model

import "time"

type Staff struct {
	Id             int64    `gorm:"column:id;type:serial;autoIncrement;primaryKey;uniqueIndex:Staff_id_key;"`
	Name           string   `gorm:"column:name;type:text;not null;index:Staff_name_rawName_idx;uniqueIndex:Staff_name_key"`
	RawName        string   `gorm:"column:rawName;type:text;not null;index:Staff_name_rawName_idx;"`
	About          string   `gorm:"column:about;type:text;default:\"\";not null;"`
	Birthday       string   `gorm:"column:birthday;type:text;default:\"\";not null;"`
	Country        string   `gorm:"column:country;type:text;default:\"\";not null;"`
	Deathday       string   `gorm:"column:deathday;type:text;default:\"\";not null;"`
	EyeColor       string   `gorm:"column:eyeColor;type:text;default:\"\";not null;"`
	Gender         string   `gorm:"column:gender;type:text;default:\"\";not null;"`
	HairColor      string   `gorm:"column:hairColor;type:text;default:\"\";not null;"`
	Height         string   `gorm:"column:height;type:text;default:\"\";not null;"`
	Weight         string   `gorm:"column:weight;type:text;default:\"\";not null;"`
	Age            int      `gorm:"column:age;type:integer;default:0;not null;"`
	JikanPersonID  int      `gorm:"column:jikanPersonID;type:integer;default:0;not null;"`
	TvmazePersonID int      `gorm:"column:tvmazePersonID;type:integer;default:0;not null;"`
	OriginalImages []string `gorm:"column:originalImages;type:text[];default:ARRAY []::text[];"`
	//---------------------------------------
	InsertDate time.Time `gorm:"column:insert_date;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	UpdateDate time.Time `gorm:"column:update_date;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	//---------------------------------------
	DislikesCount int `gorm:"column:dislikes_count;type:integer;default:0;not null;"`
	FollowCount   int `gorm:"column:follow_count;type:integer;default:0;not null;"`
	LikesCount    int `gorm:"column:likes_count;type:integer;default:0;not null;"`
	//---------------------------------------
	//---------------------------------------
	Credits          []Credit           `gorm:"foreignKey:StaffId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	FollowStaff      []FollowStaff      `gorm:"foreignKey:StaffId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	LikeDislikeStaff []LikeDislikeStaff `gorm:"foreignKey:StaffId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ImageData        CastImage          `gorm:"foreignKey:StaffId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Staff) TableName() string {
	return "Staff"
}

//---------------------------------------
//---------------------------------------

type FollowStaff struct {
	Date    time.Time `gorm:"column:date;type:timestamp(3);not null;"`
	UserId  int64     `gorm:"column:userId;type:integer;not null;primaryKey;"`
	StaffId int64     `gorm:"column:staffId;type:integer;not null;primaryKey;"`
}

func (FollowStaff) TableName() string {
	return "FollowStaff"
}

//---------------------------------------
//---------------------------------------

type LikeDislikeStaff struct {
	Date    time.Time   `gorm:"column:date;type:timestamp(3);not null;"`
	UserId  int64       `gorm:"column:userId;type:integer;not null;primaryKey;"`
	StaffId int64       `gorm:"column:staffId;type:integer;not null;primaryKey;"`
	Type    LikeDislike `gorm:"column:type;type:\"likeDislike\";not null;"`
}

func (LikeDislikeStaff) TableName() string {
	return "LikeDislikeStaff"
}
