package model

import "time"

type RelatedMovie struct {
	Date      time.Time     `gorm:"column:date;type:timestamp(3);not null;default:CURRENT_TIMESTAMP;"`
	MovieId   string        `gorm:"column:movieId;type:text;not null;uniqueIndex:RelatedMovie_movieId_relatedId_key;"`
	RelatedId string        `gorm:"column:relatedId;type:text;not null;uniqueIndex:RelatedMovie_movieId_relatedId_key;"`
	Relation  TitleRelation `gorm:"column:relation;type:\"titleRelation\";not null"`
}

func (RelatedMovie) TableName() string {
	return "RelatedMovie"
}
