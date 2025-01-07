package model

import (
	"time"

	"go.uber.org/zap"
)

type Banner struct {
	Id          int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	Title       string     `form:"title" json:"title,omitempty" gorm:"column:title;type:varchar(255);size:255;comment:轮播图名称;"`
	Description string     `form:"description" json:"description,omitempty" gorm:"column:description;type:varchar(255);size:255;comment:轮播图描述、备注;"`
	Path        string     `form:"path" json:"path,omitempty" gorm:"column:path;type:varchar(255);size:255;comment:轮播图地址;"`
	Sort        int        `form:"sort" json:"sort,omitempty" gorm:"column:sort;type:int(11);size:11;default:0;comment:排序，值越大越靠前;"`
	Enable      bool       `form:"enable" json:"enable,omitempty" gorm:"column:enable;type:tinyint(4);size:4;default:1;"`
	Type        int8       `form:"type" json:"type,omitempty" gorm:"column:type;type:tinyint(4);size:4;default:0;comment:0 网站轮播图，1 小程序轮播图，2 APP轮播图;"`
	CreatedAt   *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt   *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
	Url         string     `form:"url" json:"url,omitempty" gorm:"column:url;type:varchar(255);size:255;comment:轮播图跳转地址;"`
}

// DeleteBanner 删除数据
func (m *DBModel) DeleteBanner(ids []int64) (err error) {
	sess := m.db.Begin()
	defer func() {
		if err != nil {
			sess.Rollback()
		} else {
			sess.Commit()
		}
	}()

	err = sess.Where("id in (?)", ids).Delete(&Banner{}).Error
	if err != nil {
		m.logger.Error("DeleteBanner", zap.Error(err))
		return
	}

	return sess.Where("type = ? and type_id in (?)", AttachmentTypeBanner, ids).Delete(&Attachment{}).Error
}
