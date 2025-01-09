package banner

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

// GetBanners 获取轮播图列表
func (s *Service) GinListBanner(ctx *gin.Context) {
	var (
		req = types.NewListRequest(ctx)
		opt = &model.OptionGetList{
			Page:         int(req.Page),
			Size:         int(req.Size),
			WithCount:    true,
			SelectFields: []string{"id", "title", "path", "url"}, // 对于非权限用户，可查询的字段
			QueryIn:      make(map[string][]interface{}),
		}
	)

	if len(req.Type) > 0 {
		opt.QueryIn["type"] = cvt.ToArray(req.Type)
	}

	_, err := s.base.CheckPermission(ctx)
	if err != nil {
		opt.QueryIn["enable"] = []interface{}{true} // 非权限用户，只能查询正常状态的轮播图
	} else {
		opt.SelectFields = req.Field // 权限用户，可查询指定字段
		if len(req.Enable) > 0 {
			opt.QueryIn["enable"] = cvt.ToArray(req.Enable)
		}

		if req.Wd != "" {
			opt.QueryLike = map[string][]interface{}{"title": {req.Wd}, "description": {req.Wd}}
		}
	}

	banners, total, err := s.dbModel.GetBannerList(opt)
	if err != nil {
		ctx.JSON(http.StatusServiceUnavailable,
			types.GinResponse{Code: http.StatusServiceUnavailable, Message: err.Error(), Error: err.Error()})
		return
	}

	var pbBanner []*Banner
	util.CopyStruct(banners, &pbBanner)
	ctx.JSON(http.StatusOK, &ListReply{Total: total, Banner: pbBanner})
}

type ListReply struct {
	Banner []*Banner `json:"banner,omitempty"` // 轮播图列表
	Total  int64     `json:"total,omitempty"`
}

// banner，轮播图
type Banner struct {
	Id          int64      `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Title       string     `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Path        string     `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	Sort        int32      `protobuf:"varint,4,opt,name=sort,proto3" json:"sort,omitempty"`
	Enable      bool       `protobuf:"varint,5,opt,name=enable,proto3" json:"enable,omitempty"`
	Type        int32      `protobuf:"varint,6,opt,name=type,proto3" json:"type,omitempty"`
	Url         string     `protobuf:"bytes,7,opt,name=url,proto3" json:"url,omitempty"`
	Description string     `protobuf:"bytes,8,opt,name=description,proto3" json:"description,omitempty"`
	CreatedAt   *time.Time `protobuf:"bytes,9,opt,name=created_at,json=createdAt,proto3,stdtime" json:"created_at,omitempty"`
	UpdatedAt   *time.Time `protobuf:"bytes,10,opt,name=updated_at,json=updatedAt,proto3,stdtime" json:"updated_at,omitempty"`
}
