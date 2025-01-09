package base

import (
	"fmt"
	"moredoc/middleware/auth"
	"moredoc/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

const (
	ErrorMessageUsernameOrPasswordError = "用户名或密码不正确"
	ErrorMessageInvalidToken            = "您未登录或您的登录已过期，请重新登录或刷新页面重试"
	ErrorMessageUserNotExists           = "用户不存在"
	ErrorMessageInvalidOldPassword      = "原密码不正确"
	ErrorMessageUnsupportedCaptchaType  = "不支持的验证码类型"
)

var (
	errorMessagePermissionDeniedFormat = "您没有权限访问【%s】"
	// userLRU 是存储 user 的 LRU 缓存。它用于减少数据库查询的次数。key 是 user.UserId，value是user
	userLRU = expirable.NewLRU[int64, *auth.UserClaims](256, nil, time.Second*5)
)

type Base struct {
	au      *auth.Auth
	dbModel *model.DBModel
}

func NewBase(dbModel *model.DBModel, au *auth.Auth) *Base {
	return &Base{au: au, dbModel: dbModel}
}

func (b *Base) CheckPermission(ctx *gin.Context) (*auth.UserClaims, error) {
	user, err := b.au.CheckAuthorization(ctx.Request.Header.Get("authorization"))
	if err != nil {
		return nil, err
	}
	if v, ok := userLRU.Get(user.UserId); ok {
		return v, nil
	}

	permission, yes := b.dbModel.CheckPermissionByUserId(user.UserId, ctx.Request.URL.Path, ctx.Request.Method)
	if !yes {
		item := permission.Title
		if permission.Title == "" {
			item = permission.Path
		}
		return user, fmt.Errorf(errorMessagePermissionDeniedFormat, item)
	}
	user.HaveAccess = true
	userLRU.Add(user.UserId, user)
	return user, nil
}
