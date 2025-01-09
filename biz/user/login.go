package user

import (
	"encoding/json"
	"moredoc/biz/base"
	"moredoc/model"
	"moredoc/pkg/types"
	"moredoc/util"
	"moredoc/util/captcha"
	"moredoc/util/validate"
	"net/http"
	"time"

	"github.com/alexandrevicenzi/unchained"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) getValidFieldMap() map[string]string {
	return map[string]string{"Username": "用户名", "Password": "密码"}
}

// Login 用户登录
func (s *Service) GinLogin(ctx *gin.Context) {
	var req = &RegisterAndLoginRequest{}
	err := json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil {
		s.logger.Errorf("数据转换错误:%v", err.Error())
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: err.Error(), Error: err.Error()})
		return
	}
	err = validate.ValidateStruct(req, s.getValidFieldMap())
	if err != nil {
		s.logger.Errorf("err:%s", err.Error())
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: err.Error(), Error: err.Error()})
		return
	}

	// 如果启用了验证码，则需要进行验证码验证
	cfg := s.dbModel.GetConfigOfSecurity(model.ConfigSecurityEnableCaptchaLogin)
	if cfg.EnableCaptchaLogin {
		if req.CaptchaId == "" || req.Captcha == "" {
			ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: "请输入验证码", Error: "请输入验证码"})
			return
		}
		if !captcha.VerifyCaptcha(req.CaptchaId, req.Captcha, true) {
			ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: "验证码错误", Error: "验证码错误"})
			return
		}
	}

	user, err := s.dbModel.GetUserByUsername(req.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest,
			Message: err.Error(), Error: err.Error()})
		return
	}

	if user.Id <= 0 {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest,
			Message: "用户名或密码错误", Error: "用户名或密码错误"})
		return
	}

	ok, err := unchained.CheckPassword(req.Password, user.Password)
	if !ok || err != nil {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest,
			Message: "用户名或密码错误", Error: "用户名或密码错误"})
		return
	}

	token, err := s.auth.CreateJWTToken(user.Id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest,
			Message: err.Error(), Error: err.Error()})
		return
	}

	pbUser := &types.User{}
	util.CopyStruct(&user, pbUser)

	ip := util.GetGRPCRemoteIP(ctx)
	loginAt := time.Now()
	err = s.dbModel.UpdateByFields(&model.User{Id: user.Id, LoginAt: &loginAt, LastLoginIp: ip},
		model.TableUser, user.Id, "login_at", "last_login_ip")
	if err != nil {
		s.logger.Errorf("UpdateUser", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, types.GinResponse{Code: http.StatusInternalServerError,
			Message: err.Error(), Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &LoginReply{Token: token, User: pbUser})
}

func (s *Service) GinLogout(ctx *gin.Context) {
	user, err := s.auth.CheckAuthorization(ctx.Request.Header.Get("authorization"))
	if err != nil {
		s.logger.Errorf(base.ErrorMessageInvalidToken)
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest,
			Message: base.ErrorMessageInvalidToken, Error: base.ErrorMessageInvalidToken})
		return
	}

	// 标记退出的用户token
	s.dbModel.Logout(user.UserId, user.UUID, user.ExpiresAt)
	ctx.Status(http.StatusOK)
}
