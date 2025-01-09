package types

import (
	"moredoc/pkg/cvt"
	"time"

	"github.com/gin-gonic/gin"
)

type GinResponse struct {
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
	Code    int         `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// 请求列表
type ListRequest struct {
	Field       []string `json:"field,omitempty"`
	CreatedAt   []string `json:"created_at,omitempty"`
	Language    []string `json:"language,omitempty"`
	ParentId    []int64  `json:"parent_id,omitempty"`
	UserId      []int64  `json:"user_id,omitempty"`
	CategoryId  []int64  `json:"category_id,omitempty"`
	Status      []int32  `json:"status,omitempty"`
	Type        []int32  `json:"type,omitempty"`
	IsRecommend []bool   `json:"is_recommend,omitempty"`
	Enable      []bool   `json:"enable,omitempty"`
	Sort        string   `json:"sort,omitempty"`
	Order       string   `json:"order,omitempty"`
	Wd          string   `json:"wd,omitempty"`
	Ext         string   `json:"ext,omitempty"`
	FeeType     string   `json:"fee_type,omitempty"`
	Page        int64    `json:"page,omitempty"`
	Size        int64    `json:"size,omitempty"`
	Limit       int64    `json:"limit,omitempty"`
}

func NewListRequest(ctx *gin.Context) *ListRequest {
	query := ctx.Request.URL.Query()
	return &ListRequest{
		Page:        cvt.ToInt64(query.Get("page")),
		Size:        cvt.ToInt64(query.Get("size")),
		Wd:          cvt.ToString(query.Get("wd")),
		Field:       cvt.ToStringArray(query["field"]),
		Order:       cvt.ToString(query.Get("order")),
		CategoryId:  cvt.ToInt64Array(query["category_id"]),
		UserId:      cvt.ToInt64Array(query["user_id"]),
		Status:      cvt.ToInt32Array(query["status"]),
		IsRecommend: cvt.ToIntBoolArray(query["is_recommend"]),
		CreatedAt:   cvt.ToStringArray(query["created_at"]),
		Sort:        cvt.ToString(query.Get("sort")),
		Enable:      cvt.ToIntBoolArray(query["enable"]),
		Type:        cvt.ToInt32Array(query["type"]),
		ParentId:    cvt.ToInt64Array(query["parent_id"]),
		Limit:       cvt.ToInt64(query.Get("limit")),
		Ext:         cvt.ToString(query.Get("ext")),
		FeeType:     cvt.ToString(query.Get("fee_type")),
		Language:    cvt.ToStringArray(query["language"]),
	}
}

// 用户信息
type User struct {
	LoginAt       *time.Time `json:"login_at,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
	GroupId       []int64    `json:"group_id,omitempty"`
	Username      string     `json:"username,omitempty"`
	Mobile        string     `json:"mobile,omitempty"`
	Email         string     `json:"email,omitempty"`
	Address       string     `json:"address,omitempty"`
	Signature     string     `json:"signature,omitempty"`
	LastLoginIp   string     `json:"last_login_ip,omitempty"`
	RegisterIp    string     `json:"register_ip,omitempty"`
	Avatar        string     `json:"avatar,omitempty"`
	Identity      string     `json:"identity,omitempty"`
	Realname      string     `json:"realname,omitempty"`
	Remark        string     `json:"remark,omitempty"`
	Id            int64      `json:"id,omitempty"`
	DocCount      int32      `json:"doc_count,omitempty"`
	FollowCount   int32      `json:"follow_count,omitempty"`
	FansCount     int32      `json:"fans_count,omitempty"`
	FavoriteCount int32      `json:"favorite_count,omitempty"`
	CommentCount  int32      `json:"comment_count,omitempty"`
	Status        int32      `json:"status,omitempty"`
	CreditCount   int32      `json:"credit_count,omitempty"`
	ArticleCount  int32      `json:"article_count,omitempty"`
}
