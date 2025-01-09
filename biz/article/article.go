package article

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
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Service struct {
	dbModel *model.DBModel
	logger  logger.Logger
	base    *base.Base
}

func NewService(dbModel *model.DBModel, logger logger.Logger, auth *auth.Auth) *Service {
	return &Service{
		dbModel: dbModel,
		logger:  logger,
		base:    base.NewBase(dbModel, auth),
	}
}

// ListArticles 获取文章列表。
// 如果是有权限，则可以根据关键字来查询，否则只能简单查询
func (s *Service) GinListArticle(ctx *gin.Context) {
	var (
		req = types.NewListRequest(ctx)
		opt = &model.OptionGetArticleList{
			Page:        cvt.ToInt(req.Page),
			Size:        cvt.ToInt(req.Size),
			WithCount:   true,
			QueryLike:   make(map[string][]interface{}),
			QueryIn:     make(map[string][]interface{}),
			Sort:        []string{req.Order},
			IsRecommend: req.IsRecommend,
		}
	)

	user, _ := s.base.CheckPermission(ctx)
	if user == nil {
		opt.QueryLike["status"] = []interface{}{1}
	} else if user.HaveAccess || (len(req.UserId) > 0 && req.UserId[0] == user.UserId) {
		// 管理员或者是作者，可以查询相关状态的文档
		if req.Wd != "" {
			opt.QueryLike["title"] = []interface{}{req.Wd}
			opt.QueryLike["keywords"] = []interface{}{req.Wd}
			opt.QueryLike["description"] = []interface{}{req.Wd}
		}
		if len(req.Status) > 0 {
			opt.QueryIn["status"] = cvt.ToArray(req.Status)
		}
	}

	if len(req.CategoryId) > 0 {
		opt.QueryIn["category_id"] = cvt.ToArray(req.CategoryId)
	}

	if len(req.UserId) > 0 {
		opt.QueryIn["user_id"] = cvt.ToArray(req.UserId)
	}

	articles, total, err := s.dbModel.GetArticleList(opt)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.logger.Errorf("ListArticle", zap.Error(err))
		ctx.JSON(http.StatusServiceUnavailable,
			types.GinResponse{Code: http.StatusServiceUnavailable, Message: "获取文章列表失败", Error: err.Error()})
		return
	}

	var pbArticle []*Article
	err = util.CopyStruct(articles, &pbArticle)
	if err != nil {
		s.logger.Errorf("ListArticle", zap.Error(err))
		ctx.JSON(http.StatusServiceUnavailable,
			types.GinResponse{Code: http.StatusServiceUnavailable, Message: "获取文章列表失败", Error: err.Error()})
		return
	}

	var (
		userIds         []int64
		userMapArticles = make(map[int64][]int)
	)
	for idx, article := range articles {
		userIds = append(userIds, article.UserId)
		userMapArticles[article.UserId] = append(userMapArticles[article.UserId], idx)
	}

	if len(userIds) > 0 {
		users, _, _ := s.dbModel.GetUserList(&model.OptionGetList{
			QueryIn:      map[string][]interface{}{"id": cvt.ToArray(userIds)},
			SelectFields: []string{"id", "username", "avatar"},
		})

		for _, user := range users {
			for _, idx := range userMapArticles[user.Id] {
				pbArticle[idx].User = &types.User{
					Id:       user.Id,
					Username: user.Username,
					Avatar:   user.Avatar,
				}
			}
		}
	}

	ctx.JSON(http.StatusOK, &ListReply{
		Total:   total,
		Article: pbArticle,
	})
}

type ListReply struct {
	Article []*Article `json:"article,omitempty"` // 文章列表响应
	Total   int64      `json:"total,omitempty"`
}

// 文章
type Article struct {
	CreatedAt     *time.Time  `json:"created_at,omitempty"`
	UpdatedAt     *time.Time  `json:"updated_at,omitempty"`
	DeletedAt     *time.Time  `json:"deleted_at,omitempty"`
	RecommendAt   *time.Time  `json:"recommend_at,omitempty"`
	User          *types.User `json:"user,omitempty"`
	CategoryId    []int64     `json:"category_id,omitempty"`
	Identifier    string      `json:"identifier,omitempty"`
	Author        string      `json:"author,omitempty"`
	Title         string      `json:"title,omitempty"`
	Keywords      string      `json:"keywords,omitempty"`
	Description   string      `json:"description,omitempty"`
	Content       string      `json:"content,omitempty"`
	RejectReason  string      `json:"reject_reason,omitempty"`
	Id            int64       `json:"id,omitempty"`
	ViewCount     int64       `json:"view_count,omitempty"`
	FavoriteCount int64       `json:"favorite_count,omitempty"`
	CommentCount  int64       `json:"comment_count,omitempty"`
	UserId        int64       `json:"user_id,omitempty"`
	Status        int32       `json:"status,omitempty"`
	IsRecommend   bool        `json:"is_recommend,omitempty"`
}
