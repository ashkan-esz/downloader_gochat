package model

type ComputedFavoriteGenres struct {
	UserId  int64   `gorm:"column:userId;type:integer;not null;uniqueIndex:ComputedFavoriteGenres_userId_genre_key"`
	Genre   string  `gorm:"column:genre;type:text;not null;uniqueIndex:ComputedFavoriteGenres_userId_genre_key"`
	Count   string  `gorm:"column:count;type:integer;not null;"`
	Percent float32 `gorm:"column:percent;type:double precision;not null;"`
}

func (ComputedFavoriteGenres) TableName() string {
	return "ComputedFavoriteGenres"
}
