package attachment

import (
	"fmt"
	"mime/multipart"
	"moredoc/model"
	"moredoc/pkg/cvt"
	"moredoc/pkg/types"
	"moredoc/util"
	"moredoc/util/filetil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gofrs/uuid"
)

//	UploadArticle 上传文章相关图片和视频。这里不验证文件格式。
//
// 注意：当前适配了wangeditor的接口规范，如果需要适配其他编辑器，需要修改此接口或者增加其他接口
func (s *Service) UploadArticle(ctx *gin.Context) {
	typ := ctx.Query("type")
	if typ != "image" && typ != "video" {
		ctx.JSON(http.StatusOK, map[string]interface{}{"errno": 1, "msg": "类型参数错误"})
		return
	}

	user, err := s.base.CheckPermission(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, map[string]interface{}{"errno": 1, "msg": err.Error()})
		return
	}

	name := "file"
	fileHeader, err := ctx.FormFile(name)
	if err != nil {
		s.logger.Errorf("MultipartForm", zap.Error(err))
		ctx.JSON(http.StatusOK, map[string]interface{}{"errno": 1, "msg": err.Error()})
		return
	}

	attachment, err := s.saveFile(ctx, fileHeader)
	if err != nil {
		s.logger.Errorf("saveFile", zap.Error(err))
		ctx.JSON(http.StatusOK, map[string]interface{}{"errno": 1, "msg": err.Error()})
		return
	}
	attachment.UserId = user.UserId
	attachment.Type = model.AttachmentTypeArticle

	err = s.dbModel.Create(attachment)
	if err != nil {
		s.logger.Errorf("CreateAttachments", zap.Error(err))
		ctx.JSON(http.StatusOK, map[string]interface{}{"errno": 1, "msg": err.Error()})
		return
	}

	if typ == "image" {
		ctx.JSON(http.StatusOK, map[string]interface{}{"errno": 0, "data": map[string]interface{}{
			"url": attachment.Path,
			"alt": attachment.Name,
			// "href": "",
		}})
	} else {
		ctx.JSON(http.StatusOK, map[string]interface{}{"errno": 0, "data": map[string]interface{}{
			"url": attachment.Path,
			// "poster": "",
		}})
	}
}

// UploadBanner 上传轮播图，创建轮播图的时候，要根据附件id，更新附件的type_id字段
func (s *Service) UploadBanner(ctx *gin.Context) {
	s.uploadImage(ctx, model.AttachmentTypeBanner)
}

// 上传文档分类封面
func (s *Service) UploadCategory(ctx *gin.Context) {
	s.uploadImage(ctx, model.AttachmentTypeCategoryCover)
}

func (s *Service) uploadImage(ctx *gin.Context, attachmentType int) {
	name := "file"
	user, err := s.base.CheckPermission(ctx)
	if user == nil { // 需要登录才能上传
		ctx.JSON(http.StatusForbidden, types.GinResponse{Code: http.StatusForbidden, Message: err.Error(), Error: err.Error()})
		return
	}

	// 如果不是上传头像，需要验证用户是否有权限上传
	if attachmentType != model.AttachmentTypeAvatar && err != nil {
		ctx.JSON(http.StatusForbidden, types.GinResponse{Code: http.StatusForbidden, Message: err.Error(), Error: err.Error()})
		return
	}

	// 验证文件是否是图片
	fileHeader, err := ctx.FormFile(name)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: err.Error(), Error: err.Error()})
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !filetil.IsImage(ext) {
		message := "请上传图片格式文件，支持.jpg、.jpeg、.png、.gif、.webp、.bmp和.ico格式图片"
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: message, Error: message})
		return
	}

	attachment, err := s.saveFile(ctx, fileHeader)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: err.Error(), Error: err.Error()})
		return
	}
	attachment.Type = attachmentType
	attachment.UserId = user.UserId

	if attachmentType == model.AttachmentTypeAvatar {
		attachment.TypeId = user.UserId
		// 更新用户头像信息
		err = s.dbModel.UpdateByFields(&model.User{Id: user.UserId, Avatar: attachment.Path},
			model.TableUser, user.UserId, "avatar")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, types.GinResponse{Code: http.StatusInternalServerError,
				Message: err.Error(), Error: err.Error()})
		}
		// 标记删除旧头像附件记录
		s.dbModel.GetDB().Where("type = ? AND type_id = ?", model.AttachmentTypeAvatar, user.UserId).Delete(&model.Attachment{})
	}

	// 保存附件信息
	err = s.dbModel.Create(attachment)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.GinResponse{Code: http.StatusInternalServerError,
			Message: err.Error(), Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, types.GinResponse{Code: http.StatusOK, Message: "上传成功", Data: attachment})
}

// UploadDocument 上传文档
func (s *Service) UploadDocument(ctx *gin.Context) {
	// 检查用户是否已登录
	userClaims, err := s.base.CheckPermission(ctx)
	if err != nil {
		ctx.JSON(http.StatusForbidden, types.GinResponse{Code: http.StatusForbidden, Message: err.Error(), Error: err.Error()})
		return
	}

	// 检查用户是否有权限上传文档
	if !s.dbModel.CanIAccessUploadDocument(userClaims.UserId) {
		ctx.JSON(http.StatusForbidden, types.GinResponse{Code: http.StatusForbidden,
			Message: "没有权限上传文档", Error: "没有权限上传文档"})
		return
	}

	name := "file"
	fileheader, err := ctx.FormFile(name)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: err.Error(), Error: err.Error()})
		return
	}

	unsuportedExt := "不支持的文档类型"
	ext := strings.ToLower(filepath.Ext(fileheader.Filename))
	if !filetil.IsDocument(ext) {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: unsuportedExt, Error: unsuportedExt})
		return
	}

	allowedExt := s.dbModel.GetConfigOfSecurity(model.ConfigSecurityDocumentAllowedExt).DocumentAllowedExt
	if len(allowedExt) > 0 && !cvt.ExistStringElem(allowedExt, ext) {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: unsuportedExt, Error: unsuportedExt})
		return
	}

	attachment, err := s.saveFile(ctx, fileheader, true)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.GinResponse{Code: http.StatusInternalServerError,
			Message: err.Error(), Error: err.Error()})
		return
	}
	attachment.UserId = userClaims.UserId
	attachment.Type = model.AttachmentTypeDocument

	err = s.dbModel.Create(attachment)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.GinResponse{Code: http.StatusInternalServerError,
			Message: err.Error(), Error: err.Error()})
		return
	}

	// ctx.JSON(http.StatusOK, types.GinResponse{Code: http.StatusOK, Message: "ok", Data: attachment})
	ctx.JSON(http.StatusOK, types.GinResponse{Code: http.StatusOK, Message: "ok", Data: map[string]interface{}{"id": attachment.Id}})
}

// UploadAvatar 上传头像
func (s *Service) UploadAvatar(ctx *gin.Context) {
	s.uploadImage(ctx, model.AttachmentTypeAvatar)
}

// UploadConfig 上传配置项中的相关图片
func (s *Service) UploadConfig(ctx *gin.Context) {
	s.uploadImage(ctx, model.AttachmentTypeConfig)
}

// saveFile 保存文件。文件以md5值命名以及存储
// 同时，返回附件信息
func (s *Service) saveFile(ctx *gin.Context, fileHeader *multipart.FileHeader, isDocument ...bool) (*model.Attachment, error) {
	cacheDir := fmt.Sprintf("cache/uploads/%s", time.Now().Format("2006/01/02"))
	os.MkdirAll(cacheDir, os.ModePerm)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	cachePath := fmt.Sprintf("%s/%s%s", cacheDir, uuid.Must(uuid.NewV1()).String(), ext)
	defer func() {
		os.Remove(cachePath)
	}()

	// 保存到临时文件
	err := ctx.SaveUploadedFile(fileHeader, cachePath)
	if err != nil {
		s.logger.Errorf("SaveUploadedFile", zap.Error(err), zap.String("filename", fileHeader.Filename),
			zap.String("cachePath", cachePath))
		return nil, err
	}

	// 获取文件md5值
	md5hash, err := filetil.GetFileMD5(cachePath)
	if err != nil {
		return nil, err
	}

	savePathFormat := "uploads/%s/%s%s"
	if len(isDocument) > 0 && isDocument[0] {
		savePathFormat = "documents/%s/%s%s"
	}
	savePath := fmt.Sprintf(savePathFormat, strings.Join(strings.Split(md5hash, "")[0:5], "/"), md5hash, ext)
	os.MkdirAll(filepath.Dir(savePath), os.ModePerm)
	err = util.CopyFile(cachePath, savePath)
	if err != nil {
		s.logger.Errorf("Rename", zap.Error(err), zap.String("cachePath", cachePath), zap.String("savePath", savePath))
		return nil, err
	}

	attachment := &model.Attachment{
		Size:   fileHeader.Size,
		Name:   fileHeader.Filename,
		Ip:     ctx.ClientIP(),
		Ext:    ext,
		Enable: true, // 默认都是合法的
		Hash:   md5hash,
		Path:   "/" + savePath,
	}

	// 对于图片，直接获取图片的宽高
	if filetil.IsImage(ext) {
		attachment.Width, attachment.Height, _ = filetil.GetImageSize(cachePath)
	}

	return attachment, nil
}
