package category

import (
	"moredoc/biz/base"
	"moredoc/middleware/auth"
	"moredoc/model"
	"moredoc/pkg/cvt"
	"moredoc/pkg/logger"
	"moredoc/pkg/types"
	"moredoc/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Service struct {
	dbModel *model.DBModel
	logger  logger.Logger
	base    *base.Base
}

func NewService(dbModel *model.DBModel, logger logger.Logger, au *auth.Auth) *Service {
	return &Service{dbModel: dbModel,
		logger: logger,
		base:   base.NewBase(dbModel, au)}
}

func (s *Service) GinListCategory(ctx *gin.Context) {
	var (
		req = types.NewListRequest(ctx)
		opt = &model.OptionGetList{
			WithCount:    false,
			QueryIn:      make(map[string][]interface{}),
			SelectFields: req.Field,
			Page:         int(req.Page),
			Size:         int(req.Size),
		}
	)

	if len(req.ParentId) > 0 {
		opt.QueryIn["parent_id"] = cvt.ToArray(req.ParentId)
	}

	// 管理员，可以通过关键字搜索
	if _, err := s.base.CheckPermission(ctx); err == nil {
		if req.Wd != "" {
			opt.QueryLike = map[string][]interface{}{"title": {req.Wd}}
		}

		if len(req.Enable) > 0 {
			opt.QueryIn["enable"] = cvt.ToArray(req.Enable)
		}
	} else {
		// 非管理员，只能查询启用的
		opt.QueryIn["enable"] = []interface{}{true}
	}

	if len(req.Type) > 0 {
		opt.QueryIn["type"] = cvt.ToArray(req.Type)
	} else {
		opt.QueryIn["type"] = []interface{}{0}
	}

	cates, total, err := s.dbModel.GetCategoryList(opt)
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable,
			types.GinResponse{Code: http.StatusServiceUnavailable, Message: err.Error(), Error: err.Error()})
		return
	}

	var pbCates []*Category
	util.CopyStruct(&cates, &pbCates)
	ctx.JSON(http.StatusOK, &ListReply{Total: total, Category: pbCates})
}

type ListReply struct {
	Category []*Category `json:"category,omitempty"` // 分类列表响应
	Total    int64       `json:"total,omitempty"`
}

// 文档分类
type Category struct {
	Id              int32      `json:"id,omitempty"`
	ParentId        int32      `json:"parent_id,omitempty"`
	Title           string     `json:"title,omitempty" validate:"required"`
	DocCount        int32      `json:"doc_count,omitempty"`
	Sort            int32      `json:"sort,omitempty"`
	Enable          bool       `json:"enable,omitempty"`
	Cover           string     `json:"cover,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
	Icon            string     `json:"icon,omitempty"`
	Description     string     `json:"description,omitempty"`
	ShowDescription bool       `json:"show_description,omitempty"`
	Type            int32      `json:"type,omitempty"`
}
