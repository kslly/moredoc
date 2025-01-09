package user

import (
	"errors"
	"moredoc/biz/base"
	"moredoc/middleware/auth"
	"moredoc/model"
	"moredoc/pkg/cvt"
	"moredoc/pkg/logger"
	"moredoc/pkg/types"
	"moredoc/util"
	"moredoc/util/captcha"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Service struct {
	dbModel *model.DBModel
	logger  logger.Logger
	auth    *auth.Auth
	base    *base.Base
}

func NewService(dbModel *model.DBModel, logger logger.Logger, auth *auth.Auth) *Service {
	return &Service{
		dbModel: dbModel,
		logger:  logger,
		auth:    auth,
		base:    base.NewBase(dbModel, auth)}
}

// GetUserCaptcha 获取用户验证码
func (s *Service) GinGetUserCaptcha(ctx *gin.Context) {
	var (
		cfgCaptcha  = s.dbModel.GetConfigOfCaptcha()
		cfgSecurity = s.dbModel.GetConfigOfSecurity()
		res         = &CaptchaReply{
			Enable: false,
			Type:   cfgCaptcha.Type,
		}
		query = ctx.Request.URL.Query()
		typ   = cvt.ToString(query.Get("type"))
		err   error
	)

	defer func() {
		if err != nil {
			s.logger.Errorf("err:%s", err.Error())
			ctx.JSON(http.StatusServiceUnavailable,
				types.GinResponse{Code: http.StatusServiceUnavailable, Message: err.Error(), Error: err.Error()})
		}
	}()
	switch typ {
	case "register":
		res.Enable = cfgSecurity.EnableCaptchaRegister
	case "login":
		res.Enable = cfgSecurity.EnableCaptchaLogin
	case "find_password":
		res.Enable = cfgSecurity.EnableCaptchaFindPassword
	case "comment":
		res.Enable = cfgSecurity.EnableCaptchaComment
	default:
		err = errors.New(base.ErrorMessageUnsupportedCaptchaType)
		return
	}

	if res.Enable {
		res.Id, res.Captcha, err = captcha.GenerateCaptcha(cfgCaptcha.Type, cfgCaptcha.Length, cfgCaptcha.Width, cfgCaptcha.Height)
		if err != nil {
			return
		}
	}
	ctx.JSON(http.StatusOK, res)
}

// GetUser 根据ID获取用户信息
// 对于非管理员，只能获取公开字段
// 如果传递了Id参数，则表示查询用户的公开信息，否则查询当前用户的私有信息
func (s *Service) GinGetUser(ctx *gin.Context) {
	var (
		query           = ctx.Request.URL.Query()
		userId          = cvt.ToInt64(query.Get("id"))
		fields          = s.dbModel.GetUserPublicFields()
		pbUser          = &types.User{}
		userClaims, err = s.base.CheckPermission(ctx)
	)

	defer func() {
		ctx.JSON(http.StatusOK, pbUser)
	}()

	if err == nil || (userClaims != nil && (userClaims.UserId == userId || userId == 0)) {
		fields = []string{} // 有权限或者查的是用户自己的资料
	}

	if userId <= 0 {
		if userClaims == nil {
			s.logger.Errorf("没有获取到用户信息:%d", userId)
			return
		}
		userId = userClaims.UserId
	}

	user, err := s.dbModel.GetUser(userId, fields...)
	if err != nil {
		s.logger.Errorf("获取用户信息错误:%,err:%s", userId, err.Error())
		return
	}

	util.CopyStruct(&user, pbUser)
	pbUser.Remark = ""
	ctx.JSON(http.StatusOK, pbUser)
}

// GetUserPermissions 获取用户权限
func (s *Service) GinGetUserPermissions(ctx *gin.Context) {
	userClaims, err := s.auth.CheckAuthorization(ctx.Request.Header.Get("authorization"))
	if err != nil {
		s.logger.Errorf(base.ErrorMessageInvalidToken)
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest,
			Message: base.ErrorMessageInvalidToken, Error: base.ErrorMessageInvalidToken})
		return
	}

	permissions, err := s.dbModel.GetUserPermissinsByUserId(userClaims.UserId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.GinResponse{Code: http.StatusInternalServerError,
			Message: err.Error(), Error: err.Error()})
		return
	}

	var pbPermissions []*Permission
	util.CopyStruct(&permissions, &pbPermissions)
	ctx.JSON(http.StatusOK, &UserPermissionsReply{Permission: pbPermissions})
}
