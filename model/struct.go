package model

import (
	"time"
)

type Advertisement struct {
	Id        int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	UserId    int64      `form:"user_id" json:"user_id,omitempty" gorm:"column:user_id;type:bigint(20);size:20;default:0;comment:用户ID;"`
	Position  string     `form:"position" json:"position,omitempty" gorm:"column:position;type:varchar(64);size:64;index:idx_position;comment:广告位;"`
	StartTime *time.Time `form:"start_time" json:"start_time,omitempty" gorm:"column:start_time;type:datetime;comment:开始时间;"`
	EndTime   *time.Time `form:"end_time" json:"end_time,omitempty" gorm:"column:end_time;type:datetime;comment:截止时间;"`
	CreatedAt *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
	Title     string     `form:"title" json:"title,omitempty" gorm:"column:title;type:varchar(255);size:255;comment:广告标题;"`
	Content   string     `form:"content" json:"content,omitempty" gorm:"column:content;type:longtext;comment:广告内容;"`
	Enable    bool       `form:"enable" json:"enable,omitempty" gorm:"column:enable;type:tinyint(1);size:1;default:1;comment:是否启用;"`
	Remark    string     `form:"remark" json:"remark,omitempty" gorm:"column:remark;type:text;comment:备注;"`
}

type DocumentCategory struct {
	Id         int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	DocumentId int64      `form:"document_id" json:"document_id,omitempty" gorm:"column:document_id;type:bigint(20);size:20;default:0;index:document_id;index:idx_doc_cate,unique;comment:文档ID;"`
	CategoryId int64      `form:"category_id" json:"category_id,omitempty" gorm:"column:category_id;type:bigint(20);size:20;default:0;index:category_id;index:idx_doc_cate,unique;comment:分类ID;"`
	CreatedAt  *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt  *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
}

type Dynamic struct {
	Id        int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	UserId    int64      `form:"user_id" json:"user_id,omitempty" gorm:"column:user_id;type:bigint(20);size:20;default:0;index:idx_user_id;comment:;"`
	Content   string     `form:"content" json:"content,omitempty" gorm:"column:content;type:text;comment:内容;"`
	Type      int        `form:"type" json:"type,omitempty" gorm:"column:type;type:smallint(6);size:6;default:0;comment:类型;index:idx_type;"`
	CreatedAt *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
}

type Friendlink struct {
	Id          int        `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	Title       string     `form:"title" json:"title,omitempty" gorm:"column:title;type:varchar(64);size:64;comment:链接名称;"`
	Link        string     `form:"link" json:"link,omitempty" gorm:"column:link;type:varchar(255);size:255;comment:链接地址;"`
	Description string     `form:"description" json:"description,omitempty" gorm:"column:description;type:text;comment:描述，备注;"`
	Sort        int        `form:"sort" json:"sort,omitempty" gorm:"column:sort;type:int(11);size:11;default:0;comment:排序，值越大越靠前;"`
	Enable      bool       `form:"enable" json:"enable,omitempty" gorm:"column:enable;type:tinyint(4);size:4;default:0;"`
	CreatedAt   *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt   *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
}

const (
	DynamicTypeComment        = 1  // 发表评论
	DynamicTypeFavorite       = 2  // 收藏文档
	DynamicTypeUpload         = 3  // 上传文档
	DynamicTypeDownload       = 4  // 下载文档
	DynamicTypeLogin          = 5  // 登录
	DynamicTypeRegister       = 6  // 注册
	DynamicTypeAvatar         = 7  // 更新了头像
	DynamicTypePassword       = 8  // 修改密码
	DynamicTypeInfo           = 9  // 修改个人信息
	DynamicTypeVerify         = 10 // 实名认证
	DynamicTypeSign           = 11 // 签到
	DynamicTypeShare          = 12 // 分享文档
	DynamicTypeFollow         = 13 // 关注用户
	DynamicTypeDeleteDocument = 14 // 删除文档
)
