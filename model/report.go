package model

import (
	"time"
)

type Report struct {
	Id            int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	DocumentId    int64      `form:"document_id" json:"document_id,omitempty" gorm:"column:document_id;type:bigint(20);size:20;default:0;comment:文档ID;index:idx_document_id;"`
	DocumentTitle string     `form:"document_title" json:"document_title,omitempty" gorm:"column:document_title;type:varchar(255);size:255;default:'';comment:文档标题;"`
	UserId        int64      `form:"user_id" json:"user_id,omitempty" gorm:"column:user_id;type:bigint(20);size:20;default:0;comment:用户ID;index:idx_user_id;"`
	Username      string     `form:"username" json:"username,omitempty" gorm:"column:username;type:varchar(64);size:64;default:'';comment:用户名;"`
	Reason        int        `form:"reason" json:"reason,omitempty" gorm:"column:reason;type:int(11);size:11;default:0;comment:举报原因;"`
	Status        bool       `form:"status" json:"status,omitempty" gorm:"column:status;type:tinyint(4);size:4;default:0;comment:是否已处理;index:idx_status;"`
	Remark        string     `form:"remark" json:"remark,omitempty" gorm:"column:remark;type:varchar(255);size:255;default:'';comment:备注;"`
	CreatedAt     *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt     *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
}

// CreateReport 创建Report
func (m *DBModel) CreateReport(report *Report) error {
	doc, _ := m.GetDocument(report.DocumentId, "id", "title")
	report.DocumentTitle = doc.Title
	user, _ := m.GetUser(report.UserId, "id", "username")
	report.Username = user.Username
	return m.Create(report)
}

func (m *DBModel) GetReportByDocUser(docId, userId int64) (report Report, err error) {
	err = m.db.Where("document_id = ? and user_id = ?", docId, userId).First(&report).Error
	return
}
