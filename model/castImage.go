package model

type CastImage struct {
	Url          string `gorm:"column:url;type:text;not null;uniqueIndex:CastImage_url_key;"`
	OriginalUrl  string `gorm:"column:originalUrl;type:text;not null;uniqueIndex:ProfileImage_url_key;"`
	OriginalSize int64  `gorm:"column:originalSize;type:integer;not null;"`
	Size         int64  `gorm:"column:size;type:integer;not null;"`
	Thumbnail    string `gorm:"column:thumbnail;type:text;not null;"`
	VpnStatus    string `gorm:"column:vpnStatus;type:text;not null;"`
	StaffId      int64  `gorm:"column:staffId;type:integer;not null;uniqueIndex:CastImage_staffId_key;"`
	CharacterId  int64  `gorm:"column:characterId;type:integer;not null;uniqueIndex:CastImage_characterId_key;"`
}

func (CastImage) TableName() string {
	return "CastImage"
}
