package document

import (
	"moredoc/pkg/types"
	"time"
)

type ListReply struct {
	Document []*Document `json:"document,omitempty"` // 文档列表
	Total    int64       `json:"total,omitempty"`
}

// 首页文档查询返回项
type ListDocumentForHomeItem struct {
	Document      []*Document `json:"document,omitempty"`
	CategoryCover string      `json:"category_cover,omitempty"`
	CategoryName  string      `json:"category_name,omitempty"`
	CategoryId    int64       `json:"category_id,omitempty"`
}

// 查询文档（针对首页的查询）
type ListDocumentForHomeResponse struct {
	Document []*ListDocumentForHomeItem `json:"document,omitempty"`
}

// 创建文档
type CreateDocumentItem struct {
	Title        string `json:"title,omitempty"`
	Language     string `json:"language,omitempty"`
	AttachmentId int64  `json:"attachment_id,omitempty"`
	Price        int32  `json:"price,omitempty"`
}

// 创建文档
type CreateDocumentRequest struct {
	Document   []*CreateDocumentItem `json:"document,omitempty"`
	CategoryId []int64               `json:"category_id,omitempty"`
	Overwrite  bool                  `json:"overwrite,omitempty"`
}

// 文档
type Document struct {
	CreatedAt       *time.Time  `json:"created_at,omitempty"`
	UpdatedAt       *time.Time  `json:"updated_at,omitempty"`
	DeletedAt       *time.Time  `json:"deleted_at,omitempty"`
	RecommendAt     *time.Time  `json:"recommend_at,omitempty"`
	Attachment      *Attachment `json:"attachment,omitempty"`
	User            *types.User `json:"user,omitempty"`
	CategoryId      []int64     `json:"category_id,omitempty"`
	Cover           string      `json:"cover,omitempty"`
	Uuid            string      `json:"uuid,omitempty"`
	DeletedUsername string      `json:"deleted_username,omitempty"`
	Ext             string      `json:"ext,omitempty"`
	Username        string      `json:"username,omitempty"`
	ConvertError    string      `json:"convert_error,omitempty"`
	PreviewExt      string      `json:"preview_ext,omitempty"`
	Language        string      `json:"language,omitempty"`
	Content         string      `json:"content,omitempty"`
	Title           string      `json:"title,omitempty"`
	Keywords        string      `json:"keywords,omitempty"`
	Description     string      `json:"description,omitempty"`
	Size            int64       `json:"size,omitempty"`
	UserId          int64       `json:"user_id,omitempty"`
	DeletedUserId   int64       `json:"deleted_user_id,omitempty"`
	Id              int64       `json:"id,omitempty"`
	Status          int32       `json:"status,omitempty"`
	Width           int32       `json:"width,omitempty"`
	Height          int32       `json:"height,omitempty"`
	Preview         int32       `json:"preview,omitempty"`
	Pages           int32       `json:"pages,omitempty"`
	DownloadCount   int32       `json:"download_count,omitempty"`
	ViewCount       int32       `json:"view_count,omitempty"`
	FavoriteCount   int32       `json:"favorite_count,omitempty"`
	CommentCount    int32       `json:"comment_count,omitempty"`
	Score           int32       `json:"score,omitempty"`
	ScoreCount      int32       `json:"score_count,omitempty"`
	Price           int32       `json:"price,omitempty"`
	EnableGzip      bool        `json:"enable_gzip,omitempty"`
}

// 附件
type Attachment struct {
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Hash        string     `json:"hash,omitempty"`
	Path        string     `json:"path,omitempty"`
	Name        string     `json:"name,omitempty"`
	Ext         string     `json:"ext,omitempty"`
	Ip          string     `json:"ip,omitempty"`
	Username    string     `json:"username,omitempty"`
	TypeName    string     `json:"type_name,omitempty"`
	Description string     `json:"description,omitempty"`
	Id          int64      `json:"id,omitempty"`
	UserId      int64      `json:"user_id,omitempty"`
	TypeId      int64      `json:"type_id,omitempty"`
	Size        int64      `json:"size,omitempty"`
	Width       int64      `json:"width,omitempty"`
	Height      int64      `json:"height,omitempty"`
	Type        int32      `json:"type,omitempty"`
	Enable      bool       `json:"enable,omitempty"`
}
