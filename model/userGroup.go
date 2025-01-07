package model

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserGroup struct {
	Id        int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	UserId    int64      `form:"user_id" json:"user_id,omitempty" gorm:"column:user_id;type:bigint(20);size:20;default:0;index:user_group,unique;index:user_id;comment:用户ID;"`
	GroupId   int64      `form:"group_id" json:"group_id,omitempty" gorm:"column:group_id;type:bigint(20);size:20;default:0;index:user_group,unique;comment:组ID;"`
	CreatedAt *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
}

// 这里是proto文件中的结构体，可以根据需要删除或者调整
//message UserGroup {
// int64 id = 1;
// int64 user_id = 2;
// int64 group_id = 3;
//   = 0;
//   = 0;
//}

// GetUserGroupList 获取UserGroup列表
func (m *DBModel) GetUserGroupList(opt *OptionGetList) (userGroupList []UserGroup, total int64, err error) {
	db, total, err := m.queryCond(TableUserGroup, &UserGroup{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	err = db.Find(&userGroupList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetUserGroupList", zap.Error(err))
	}
	return
}
