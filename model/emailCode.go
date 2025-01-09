package model

import (
	"time"

	"go.uber.org/zap"
)

type EmailCode struct {
	Id        int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	Email     string     `form:"email" json:"email,omitempty" gorm:"column:email;type:varchar(64);size:64;index:idx_email;comment:邮箱;"`
	Ip        string     `form:"ip" json:"ip,omitempty" gorm:"column:ip;type:varchar(64);size:64;comment:IP地址;"`
	Code      string     `form:"code" json:"code,omitempty" gorm:"column:code;type:varchar(16);size:16;comment:邮箱验证码;"`
	CodeType  int        `form:"code_type" json:"code_type,omitempty" gorm:"column:code_type;type:smallint(6);size:6;default:0;comment:验证码类型,0注册,1登录;"`
	Success   bool       `form:"success" json:"success,omitempty" gorm:"column:success;type:tinyint(1);size:1;default:0;comment:是否发送成功;"`
	Error     string     `form:"error" json:"error,omitempty" gorm:"column:error;type:text;comment:错误信息;"`
	CreatedAt *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:;"`
	UpdatedAt *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:;"`
	IsUsed    bool       `form:"is_used" json:"is_used,omitempty" gorm:"column:is_used;type:tinyint(1);size:1;default:0;comment:是否已使用;"`
}

// 获取最新一条邮箱验证码
func (m *DBModel) GetLatestEmailCode(email string, codeType int32) (code EmailCode) {
	err := m.db.Where("email = ? and code_type = ?", email, codeType).Order("id desc").Find(&code).Error
	if err != nil {
		m.logger.Errorf("GetLatestEmailCode", zap.Error(err))
	}
	return
}
