package model

import "database/sql/driver"

type TitleRelation string
type UserRole string
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
	TEST_USER           UserRole      = "test_user"
	USER                UserRole      = "user"
	DEV                 UserRole      = "dev"
	ADMIN               UserRole      = "admin"
	LIKE                LikeDislike   = "like"
	DISLIKE             LikeDislike   = "dislike"
)

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
	*tl = TitleRelation(value.([]byte))
	return nil
}

func (tl TitleRelation) Value() (driver.Value, error) {
	return string(tl), nil
}

func (ur *UserRole) Scan(value interface{}) error {
	*ur = UserRole(value.([]byte))
	return nil
}

func (ur UserRole) Value() (driver.Value, error) {
	return string(ur), nil
}

func (l *LikeDislike) Scan(value interface{}) error {
	*l = LikeDislike(value.([]byte))
	return nil
}

func (l LikeDislike) Value() (driver.Value, error) {
	return string(l), nil
}

func (m *MbtiType) Scan(value interface{}) error {
	*m = MbtiType(value.([]byte))
	return nil
}

func (m MbtiType) Value() (driver.Value, error) {
	return string(m), nil
}
