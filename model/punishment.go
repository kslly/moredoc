package model

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	PunishmentTypeDisabled              = 1 // 禁用账户：禁止登录、禁止评论、禁止上传、禁止下载、禁止收藏
	PunishmentTypeCommentLimited        = 2 // 禁止评论
	PunishmentTypeUploadLimited         = 3 // 禁止上传
	PunishmentTypeDownloadLimited       = 4 // 禁止下载
	PunishmentTypeFavoriteLimited       = 5 // 禁止收藏
	PunishmentTypePublishArticleLimited = 6 // 禁止发布文章
)

type Punishment struct {
	Id        int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:自增主键;"`
	UserId    int64      `form:"user_id" json:"user_id,omitempty" gorm:"column:user_id;type:bigint(20);size:20;default:0;index:idx_user_id;comment:用户ID;"`
	Type      int        `form:"type" json:"type,omitempty" gorm:"column:type;type:int(11);size:11;default:0;comment:惩罚类型，对应user表的status;"`
	Enable    bool       `form:"enable" json:"enable,omitempty" gorm:"column:enable;type:tinyint(1);size:1;default:0;index:idx_enable;comment:0 关闭，1启用;"`
	Operators string     `form:"operators" json:"operators,omitempty" gorm:"column:operators;type:text;comment:操作信息;"`
	Reason    string     `form:"reason" json:"reason,omitempty" gorm:"column:reason;type:text;comment:惩罚原因;"`
	Remark    string     `form:"remark" json:"remark,omitempty" gorm:"column:remark;type:text;comment:惩罚备注;"`
	EndTime   *time.Time `form:"end_time" json:"end_time,omitempty" gorm:"column:end_time;type:datetime;comment:惩罚结束时间，没有结束时间，则表示永久;"`
	CreatedAt *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
}

type PunishmentOperator struct {
	UserId    int64 `json:"u"`
	Type      int32 `json:"t"`
	Timestamp int64 `json:"ts"`
}

func (m *DBModel) MakePunishmentOperators(userId int64, punishmentType int32, operaterStr ...string) string {
	var operators []PunishmentOperator

	if len(operaterStr) > 0 && operaterStr[0] != "" {
		err := jsoniter.Unmarshal([]byte(operaterStr[0]), &operators)
		if err != nil {
			m.logger.Errorf("FormatPunishmentOperators", zap.Error(err))
			return operaterStr[0]
		}
	}

	operators = append(operators, PunishmentOperator{
		UserId:    userId,
		Type:      punishmentType,
		Timestamp: time.Now().Unix(),
	})

	operatersByte, err := jsoniter.Marshal(operators)
	if err != nil {
		m.logger.Errorf("FormatPunishmentOperators", zap.Error(err))
		if len(operaterStr) > 0 {
			return operaterStr[0]
		}
		return ""
	}
	return string(operatersByte)
}

// GetPunishment 根据id获取Punishment
func (m *DBModel) GetPunishment(id interface{}, fields ...string) (punishment Punishment, err error) {
	db := m.db

	fields = m.FilterValidFields(TablePunishment, fields...)
	if len(fields) > 0 {
		db = db.Select(fields)
	}

	err = db.Where("id = ?", id).First(&punishment).Error
	return
}

func (m *DBModel) isInPunishing(userId int64, types []int) (yes bool, err error) {
	if userId <= 1 {
		return false, nil
	}

	punishment := &Punishment{}
	err = m.db.Model(punishment).Select("id").
		Where(
			"user_id = ? and enable = ? and type in ? and (end_time IS NULL or end_time > ?)",
			userId, true, types, time.Now(),
		).Find(&punishment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		m.logger.Errorf("isInPunishing", zap.Error(err))
		return
	}
	return punishment.Id > 0, nil
}
