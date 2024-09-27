package model

import (
	"database/sql/driver"
)

type TitleRelation string
type LikeDislike string
type MbtiType string

const (
	PREQUEL             TitleRelation = "prequel"
	SEQUEL              TitleRelation = "sequel"
	SPIN_OFF            TitleRelation = "spin_off"
	SIDE_STORY          TitleRelation = "side_story"
	FULL_STORY          TitleRelation = "full_story"
	SUMMARY             TitleRelation = "summary"
	PARENT_STORY        TitleRelation = "parent_story"
	OTHER               TitleRelation = "other"
	ALTERNATIVE_SETTING TitleRelation = "alternative_setting"
	ALTERNATIVE_VERSION TitleRelation = "alternative_version"
	LIKE                LikeDislike   = "like"
	DISLIKE             LikeDislike   = "dislike"
)

var MbtiTypes = []MbtiType{ISTJ, ISFJ, INFJ, INTJ, ISTP, ISFP, INFP, INTP, ESTP, ESFP, ENFP, ENTP, ESTJ, ESFJ, ENFJ}

const (
	ISTJ MbtiType = "ISTJ"
	ISFJ MbtiType = "ISFJ"
	INFJ MbtiType = "INFJ"
	INTJ MbtiType = "INTJ"
	ISTP MbtiType = "ISTP"
	ISFP MbtiType = "ISFP"
	INFP MbtiType = "INFP"
	INTP MbtiType = "INTP"
	ESTP MbtiType = "ESTP"
	ESFP MbtiType = "ESFP"
	ENFP MbtiType = "ENFP"
	ENTP MbtiType = "ENTP"
	ESTJ MbtiType = "ESTJ"
	ESFJ MbtiType = "ESFJ"
	ENFJ MbtiType = "ENFJ"
	ENTJ MbtiType = "ENTJ"
)

func (tl *TitleRelation) Scan(value interface{}) error {
	*tl = TitleRelation(value.(string))
	return nil
}

func (tl TitleRelation) Value() (driver.Value, error) {
	return string(tl), nil
}

func (l *LikeDislike) Scan(value interface{}) error {
	*l = LikeDislike(value.(string))
	return nil
}

func (l LikeDislike) Value() (driver.Value, error) {
	return string(l), nil
}

func (m *MbtiType) Scan(value interface{}) error {
	*m = MbtiType(value.(string))
	return nil
}

func (m MbtiType) Value() (driver.Value, error) {
	return string(m), nil
}
