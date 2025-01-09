package friendlink

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
	return &Service{
		dbModel: dbModel,
		logger:  logger,
		base:    base.NewBase(dbModel, au)}
}

func (s *Service) GinListFriendlink(ctx *gin.Context) {
	var (
		req = types.NewListRequest(ctx)
		opt = &model.OptionGetList{
			WithCount:    true,
			Page:         int(req.Page),
			Size:         int(req.Size),
			SelectFields: req.Field,
		}
	)

	_, err := s.base.CheckPermission(ctx)
	if err == nil {
		// 管理员可使用like查询
		if req.Wd != "" {
			wd := "%" + req.Wd + "%"
			opt.QueryLike = map[string][]interface{}{
				"title":       {wd},
				"description": {wd},
			}
		}
		// 管理员可查询指定状态的友链
		if len(req.Enable) > 0 {
			opt.QueryIn = map[string][]interface{}{"enable": cvt.ToArray(req.Enable)}
		}
	} else {
		// 非管理员可查询的字段
		opt.SelectFields = []string{"id", "title", "link"}
		opt.QueryIn = map[string][]interface{}{"enable": {true}}
	}

	friendlink, total, err := s.dbModel.GetFriendlinkList(opt)
	if err != nil {
		s.logger.Errorf("err:%s", err.Error())
		ctx.JSON(http.StatusServiceUnavailable,
			types.GinResponse{Code: http.StatusServiceUnavailable, Message: err.Error(), Error: err.Error()})
		return
	}

	var pbFriendlink []*Friendlink
	util.CopyStruct(friendlink, &pbFriendlink)
	ctx.JSON(http.StatusOK, &ListReply{Friendlink: pbFriendlink, Total: total})
}

type ListReply struct {
	Friendlink []*Friendlink `json:"friendlink,omitempty"` // 友情链接列表
	Total      int64         `json:"total,omitempty"`
}

// 友情链接
type Friendlink struct {
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Title       string     `json:"title,omitempty"`
	Link        string     `json:"link,omitempty"`
	Description string     `json:"description,omitempty"`
	Id          int32      `json:"id,omitempty"`
	Sort        int32      `json:"sort,omitempty"`
	Enable      bool       `json:"enable,omitempty"`
}
