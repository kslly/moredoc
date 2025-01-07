package model

import (
	"errors"
	"time"
)

type Language struct {
	Id        int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	Language  string     `form:"language" json:"language,omitempty" gorm:"column:language;type:varchar(64);size:64;comment:语言;"`
	Enable    bool       `form:"enable" json:"enable,omitempty" gorm:"column:enable;type:tinyint(1);size:1;default:0;index:idx_enable;comment:是否启用;"`
	Code      string     `form:"code" json:"code,omitempty" gorm:"column:code;type:varchar(16);size:16;comment:语言代码;index:idx_code,unique;"`
	Total     int        `form:"total" json:"total,omitempty" gorm:"column:total;type:int(11);size:11;default:0;comment:文档数;"`
	Sort      int        `form:"sort" json:"sort,omitempty" gorm:"column:sort;type:int(11);size:11;default:0;comment:排序，值越大越靠前;"`
	CreatedAt *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
}

func (m *DBModel) UpdateLanguageStatus(languageId []int64, enable bool) (err error) {
	return m.DB().Model(&Language{}).Where("id in (?)", languageId).Update("enable", enable).Error
}

// UpdateLanguage 更新Language
func (m *DBModel) UpdateLanguage(language *Language, field ...string) (err error) {
	db := m.DB().Model(language)
	if len(field) > 0 {
		db = db.Select(field)
	}
	return db.Updates(language).Error
}

func (m *DBModel) CreateLanguage(language *Language) error {
	var lang Language
	err := m.DB().Where("code = ?", language.Code).Find(&lang).Error
	if err != nil {
		return err
	}

	if lang.Id > 0 {
		return errors.New("语言代码已存在")
	}

	return m.DB().Create(language).Error
}

func (m *DBModel) DeleteLanguage(id []int64) (err error) {
	return m.DB().Where("id in (?)", id).Delete(&Language{}).Error
}
