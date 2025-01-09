package user

import (
	v1 "moredoc/api/v1"
	"moredoc/biz/base"
	"moredoc/pkg/types"
	"moredoc/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Service) GinGetSignedToday(ctx *gin.Context) {
	user, err := s.auth.CheckAuthorization(ctx.Request.Header.Get("authorization"))
	if err != nil {
		s.logger.Errorf(base.ErrorMessageInvalidToken)
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest,
			Message: base.ErrorMessageInvalidToken, Error: base.ErrorMessageInvalidToken})
		return
	}

	sign := s.dbModel.GetSignedToday(user.UserId)
	if sign.Id == 0 {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest,
			Message: "您今天还没有签到", Error: "您今天还没有签到"})
		return
	}
	pbSign := &v1.Sign{}
	util.CopyStruct(&sign, pbSign)
	ctx.JSON(http.StatusOK, pbSign)
}
