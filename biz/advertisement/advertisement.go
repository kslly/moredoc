package advertisement

import (
	"moredoc/model"
	"moredoc/pkg/cvt"
	"moredoc/pkg/logger"
	"moredoc/pkg/types"
	"moredoc/util"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	dbModel *model.DBModel
	logger  logger.Logger
}

func NewService(dbModel *model.DBModel, logger logger.Logger) *Service {
	return &Service{dbModel: dbModel, logger: logger}
}

func (s *Service) GinGetAdvertisementByPosition(ctx *gin.Context) {
	// 根据广告位获取广告
	var (
		query    = ctx.Request.URL.Query()
		position = cvt.ToStringArray(query["position"])
		now      = time.Now()
		opt      = &model.OptionGetList{
			WithCount: false,
			Page:      1,
			Size:      100000,
			QueryIn: map[string][]interface{}{
				"enable": {true},
			},
			QueryRange: map[string][2]interface{}{
				"start_time": {nil, now},
				"end_time":   {now, nil},
			},
		}
	)

	if len(position) > 0 {
		opt.QueryIn["position"] = cvt.ToArray(position)
	}

	advs, _, err := s.dbModel.GetAdvertisementList(opt)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Errorf("GetAdvertisement", zap.Error(err))
		ctx.JSON(http.StatusServiceUnavailable,
			types.GinResponse{Code: http.StatusServiceUnavailable, Message: "获取广告失败", Error: err.Error()})
		return
	}

	res := &ListReply{}
	err = util.CopyStruct(advs, &res.Advertisement)
	if err != nil {
		s.logger.Errorf("GetAdvertisement", zap.Error(err))
		ctx.JSON(http.StatusServiceUnavailable,
			types.GinResponse{Code: http.StatusServiceUnavailable, Message: err.Error(), Error: err.Error()})
		return
	}

	for idx := range res.Advertisement {
		res.Advertisement[idx].Remark = "" // 去除备注
	}

	ctx.JSON(http.StatusOK, res)
}

type ListReply struct {
	Advertisement []*Advertisement `json:"advertisement,omitempty"`
	// 总数
	Total int64 `json:"total,omitempty"`
}

type Advertisement struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	Id        int64      `json:"id,omitempty"`
	UserId    int64      `json:"user_id,omitempty"`
	Position  string     `json:"position,omitempty"`
	Content   string     `json:"content,omitempty"`
	Enable    bool       `json:"enable,omitempty"`
	Remark    string     `json:"remark,omitempty"`
	Title     string     `json:"title,omitempty"`
}
