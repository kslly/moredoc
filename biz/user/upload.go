package user

import (
	"moredoc/biz/base"
	"moredoc/pkg/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 检查上传权限
func (s *Service) GinCanIUploadDocument(ctx *gin.Context) {
	user, err := s.auth.CheckAuthorization(ctx.Request.Header.Get("authorization"))
	if err != nil {
		s.logger.Errorf(base.ErrorMessageInvalidToken)
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest,
			Message: base.ErrorMessageInvalidToken, Error: base.ErrorMessageInvalidToken})
		return
	}

	if !s.dbModel.CanIAccessUploadDocument(user.UserId) {
		ctx.Status(http.StatusForbidden)
	}

	ctx.Status(http.StatusOK)
}
