package model

import (
	"time"

	"go.uber.org/zap"
)

type Navigation struct {
	Id          int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	Title       string     `form:"title" json:"title,omitempty" gorm:"column:title;type:varchar(255);size:255;comment:链接名称;"`
	Href        string     `form:"href" json:"href,omitempty" gorm:"column:href;type:varchar(255);size:255;comment:跳转链接;"`
	Target      string     `form:"target" json:"target,omitempty" gorm:"column:target;type:varchar(16);size:16;comment:打开方式;"`
	Color       string     `form:"color" json:"color,omitempty" gorm:"column:color;type:varchar(32);size:32;comment:链接颜色;"`
	Sort        int        `form:"sort" json:"sort,omitempty" gorm:"column:sort;type:int(11);size:11;default:0;comment:排序，值越大越靠前;"`
	Enable      *bool      `form:"enable" json:"enable,omitempty" gorm:"column:enable;type:int(11);size:11;default:0;comment:是否启用;"`
	ParentId    int64      `form:"parent_id" json:"parent_id,omitempty" gorm:"column:parent_id;type:int(11);size:11;default:0;comment:上级id;"`
	Description string     `form:"description" json:"description,omitempty" gorm:"column:description;type:varchar(1024);size:1024;comment:描述;"`
	CreatedAt   *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt   *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
	Fixed       bool       `form:"fixed" json:"fixed,omitempty" gorm:"column:fixed;type:tinyint(1);size:1;default:0;comment:是否固定;"`
}

// CreateNavigation 创建Navigation
func (m *DBModel) CreateNavigation(navigation *Navigation) error {
	navigation.Fixed = false
	return m.Create(navigation)
}

// UpdateNavigation 更新Navigation
func (m *DBModel) UpdateNavigation(navigation *Navigation) (err error) {
	db := m.db.Model(navigation)
	db = db.Select(m.GetTableFields(TableNavigation))

	if navigation.Id == navigation.ParentId {
		navigation.ParentId = 0
	}

	return db.Omit("fixed").Where("id = ?", navigation.Id).Updates(navigation).Error
}

// DeleteNavigation 删除数据
// 连同子数据一起删除
func (m *DBModel) DeleteNavigation(ids []int64) error {
	err := m.db.Where("id in (?) and fixed = ?", ids, false).Delete(&Navigation{}).Error
	if err != nil {
		return err
	}

	var children []Navigation
	m.db.Select("id").Where("parent_id in (?)", ids).Find(&children)
	if len(children) == 0 {
		return nil
	}

	var childrenIds []int64
	for _, child := range children {
		childrenIds = append(childrenIds, child.Id)
	}
	return m.DeleteNavigation(childrenIds)
}

func (m *DBModel) initNavigation() {
	enable := true
	navs := []Navigation{
		{Title: "首页", Href: "/", Target: "_self", Color: "", Sort: 102400, Enable: &enable, Fixed: true},
		{Title: "文库资料", Href: "/category", Target: "_self", Sort: 102300, Enable: &enable, Fixed: true},
		{Title: "文章资讯", Href: "/article", Target: "_self", Sort: 102200, Enable: &enable, Fixed: true},
	}
	for _, nav := range navs {
		exist := &Navigation{}
		m.db.Model(&Navigation{}).Where("href = ?", nav.Href).First(exist)
		if exist.Id == 0 {
			err := m.Create(&nav)
			if err != nil {
				m.logger.Errorf("initNavigation", zap.Error(err))
			}
		}
	}
}
