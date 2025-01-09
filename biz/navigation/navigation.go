package navigation

import (
	"moredoc/model"
	"moredoc/pkg/cvt"
	"moredoc/pkg/logger"
	"moredoc/pkg/types"
	"moredoc/util"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Service struct {
	dbModel *model.DBModel
	logger  logger.Logger
}

func NewService(dbModel *model.DBModel, logger logger.Logger) *Service {
	return &Service{dbModel: dbModel, logger: logger}
}

func (s *Service) GinListNavigation(ctx *gin.Context) {
	var (
		req = types.NewListRequest(ctx)
		opt = &model.OptionGetList{
			Page:         cvt.ToInt(req.Page),
			Size:         cvt.ToInt(req.Size),
			WithCount:    true,
			SelectFields: req.Field,
		}

		err error
	)

	if req.Order != "" {
		opt.Sort = strings.Split(req.Order, ",")
	}

	if req.Wd != "" {
		opt.QueryLike = map[string][]interface{}{
			"title":       {req.Wd},
			"description": {req.Wd},
			"href":        {req.Wd},
		}
	}

	defer func() {
		if err != nil {
			s.logger.Errorf("err:%s", err.Error())
			ctx.JSON(http.StatusServiceUnavailable,
				types.GinResponse{Code: http.StatusServiceUnavailable,
					Message: err.Error(),
					Error:   err.Error(),
				})
		}
	}()

	navs, total, errMsg := s.dbModel.GetNavigationList(opt)
	if errMsg != nil {
		err = errMsg
		return
	}

	res := &ListReply{
		Total: total,
	}

	err = util.CopyStruct(navs, &res.Navigation)
	if err != nil {
		return
	}

	ctx.JSON(http.StatusOK, res)
}

type ListReply struct {
	Navigation []*Navigation `json:"navigation,omitempty"`
	Total      int64         `json:"total,omitempty"`
}

type Navigation struct {
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Title       string     `json:"title,omitempty"`
	Href        string     `json:"href,omitempty"`
	Target      string     `json:"target,omitempty"`
	Color       string     `json:"color,omitempty"`
	Description string     `json:"description,omitempty"`
	Id          int64      `json:"id,omitempty"`
	ParentId    int64      `json:"parent_id,omitempty"`
	Sort        int32      `json:"sort,omitempty"`
	Enable      bool       `json:"enable,omitempty"`
	Fixed       bool       `json:"fixed,omitempty"`
}
