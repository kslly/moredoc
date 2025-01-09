package model

import (
	"time"

	"go.uber.org/zap"
)

type SearchRecord struct {
	Id        int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	UserId    int64      `form:"user_id" json:"user_id,omitempty" gorm:"column:user_id;type:bigint(20);size:20;default:0;comment:搜索用户;"`
	Ip        string     `form:"ip" json:"ip,omitempty" gorm:"column:ip;type:varchar(64);size:64;comment:IP地址;"`
	Total     int        `form:"total" json:"total,omitempty" gorm:"column:total;type:int(11);size:11;default:0;comment:搜索结果;"`
	Page      int        `form:"page" json:"page,omitempty" gorm:"column:page;type:int(11);size:11;default:0;comment:搜索页码;"`
	UserAgent string     `form:"user_agent" json:"user_agent,omitempty" gorm:"column:user_agent;type:varchar(512);size:512;comment:请求客户端;"`
	Keywords  string     `form:"keywords" json:"keywords,omitempty" gorm:"column:keywords;type:varchar(64);size:64;comment:搜索关键字;"`
	SpendTime float64    `form:"spend_time" json:"spend_time,omitempty" gorm:"column:spend_time;type:decimal(10,2);size:10;default:0.00;comment:搜索耗时;"`
	CreatedAt *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;index:idx_created_at;comment:创建时间，搜索时间;"`
	UpdatedAt *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
	Type      int        `form:"type" json:"type,omitempty" gorm:"column:type;type:int(11);size:11;default:0;comment:搜索类型,0文档，1文章;"`
}

var searchRecordQueue = make(chan *SearchRecord, 1024)

// CreateSearchRecord 创建SearchRecord
func (m *DBModel) CreateSearchRecord(searchRecord *SearchRecord) (err error) {
	// 添加到队列
	now := time.Now()
	searchRecord.CreatedAt = &now
	searchRecord.UpdatedAt = &now
	searchRecordQueue <- searchRecord
	return
}

// DeleteSearchRecord 删除数据
func (m *DBModel) DeleteSearchRecord(ids []int64) (err error) {
	return m.DB().Where("id in (?)", ids).Delete(&SearchRecord{}).Error
}

// 通过队列创建搜索记录
func (m *DBModel) createSearchRecordFromQueue() {
	// 读取队列，批量添加到数据库中
	var (
		searchRecordList []*SearchRecord
		ticker           = time.NewTicker(time.Second * 10)
	)
	for {
		select {
		case searchRecord := <-searchRecordQueue:
			searchRecordList = append(searchRecordList, searchRecord)
		case <-ticker.C:
			if len(searchRecordList) > 0 {
				err := m.DB().CreateInBatches(&searchRecordList, 10).Error
				if err != nil {
					m.logger.Errorf("createSearchRecordByQueue", zap.Error(err))
				}
				searchRecordList = nil

				// 清理过期数据
				retentionDays := m.GetConfigOfSecurity(ConfigSecuritySearchRecordRetentionDays).SearchRecordRetentionDays
				err = m.DB().Where("created_at < ?", time.Now().AddDate(0, 0, -int(retentionDays)).Format("2006-01-02 00:00:00")).Delete(&SearchRecord{}).Error
				if err != nil {
					m.logger.Errorf("createSearchRecordByQueue", zap.Error(err))
				}
			}
		}
	}
}
