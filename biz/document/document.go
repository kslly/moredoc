package document

import (
	"encoding/json"
	"fmt"
	"moredoc/biz/base"
	"moredoc/middleware/auth"
	"moredoc/model"
	"moredoc/pkg/cvt"
	"moredoc/pkg/logger"
	"moredoc/pkg/types"
	"moredoc/util"
	"moredoc/util/filetil"
	"moredoc/util/segword/jieba"
	"net/http"
	"strings"
	"time"

	"github.com/araddon/dateparse"
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
		base:    base.NewBase(dbModel, auth),
	}
}

// 针对首页的文档查询
func (s *Service) GinListDocumentForHome(ctx *gin.Context) {
	// 1. 查询启用了的分类
	categories, _, _ := s.dbModel.GetCategoryList(&model.OptionGetList{
		WithCount: false,
		QueryIn: map[string][]interface{}{
			"enable":    {true},
			"parent_id": {0},
			"type":      {model.CategoryTypeDocument}, // 仅限文档分类
		},
	})

	if len(categories) == 0 {
		ctx.JSON(http.StatusOK, &ListDocumentForHomeResponse{})
	}

	var (
		query = ctx.Request.URL.Query()
		field = cvt.ToStringArray(query["field"])
		limit = cvt.ToInt(query.Get("limit"))
	)

	if limit == 0 || limit > 100 {
		limit = 5
	}

	defaultFields := []string{"id", "title", "ext", "uuid", "pages"}
	if len(field) > 0 {
		defaultFields = append(defaultFields, field...)
	}

	var docIds []int64
	resp := &ListDocumentForHomeResponse{}
	for _, category := range categories {
		docs, _, _ := s.dbModel.GetDocumentList(&model.OptionGetList{
			WithCount: false,
			QueryIn: map[string][]interface{}{
				"category_id": {category.Id},
				"status":      {model.DocumentStatusConverted},
			},
			Page:         1,
			Size:         limit,
			Sort:         []string{"id desc"},
			SelectFields: defaultFields,
		})

		var pbDocs []*Document
		util.CopyStruct(&docs, &pbDocs)
		resp.Document = append(resp.Document, &ListDocumentForHomeItem{
			CategoryId:    category.Id,
			CategoryName:  category.Title,
			CategoryCover: category.Cover,
			Document:      pbDocs,
		})

		for _, doc := range docs {
			docIds = append(docIds, doc.Id)
		}
	}

	// 查找文档相关联的附件。对于列表，只返回hash和id，不返回其他字段
	attachments, _, _ := s.dbModel.GetAttachmentList(&model.OptionGetList{
		WithCount:    false,
		SelectFields: []string{"hash", "id", "type_id"},
		QueryIn: map[string][]interface{}{
			"type_id": cvt.ToArray(docIds),
			"type":    {model.AttachmentTypeDocument},
		},
	})

	docIdMapAttachmentHash := make(map[int64]string)
	for _, attachment := range attachments {
		docIdMapAttachmentHash[attachment.TypeId] = attachment.Hash
	}

	for _, item := range resp.Document {
		for _, doc := range item.Document {
			if hash, ok := docIdMapAttachmentHash[doc.Id]; ok {
				doc.Cover = fmt.Sprintf("/view/cover/%s", hash)
			}
		}
	}

	ctx.JSON(http.StatusOK, resp)
}

// ListDocument 查询文档列表
// 1. 对于普通用户，只能查询未禁用的文档，且最多只能查询100页
// 2. 对于管理员，可以查询所有文档，可以根据关键字进行查询
func (s *Service) GinListDocument(ctx *gin.Context) {
	var (
		req = types.NewListRequest(ctx)
		opt = &model.OptionGetList{
			WithCount:    req.Limit <= 0,
			Page:         cvt.ToInt(req.Page),
			Size:         cvt.ToInt(req.Size),
			SelectFields: req.Field,
			QueryIn:      make(map[string][]interface{}),
			QueryLike:    make(map[string][]interface{}),
			QueryRange:   make(map[string][2]interface{}),
			IsRecommend:  req.IsRecommend,
			FeeType:      req.FeeType,
		}
	)

	if len(req.Order) > 0 {
		opt.Sort = []string{req.Order}
	}

	if len(req.CategoryId) > 0 {
		opt.QueryIn["category_id"] = cvt.ToArray(req.CategoryId)
	}

	if len(req.UserId) > 0 {
		opt.QueryIn["user_id"] = []interface{}{req.UserId[0]}
	}

	if len(req.Language) > 0 {
		var languages []interface{}
		for _, lang := range req.Language {
			if lang == "" {
				continue
			}
			languages = append(languages, lang)
		}
		if len(languages) > 0 {
			opt.QueryIn["language"] = languages
		}
	}

	if exts := filetil.GetExts(req.Ext); len(exts) > 0 {
		opt.QueryIn["ext"] = cvt.ToArray(exts)
	}

	if l := len(req.CreatedAt); l > 0 {
		end := time.Now()
		start, _ := dateparse.ParseLocal(req.CreatedAt[0])
		if l > 1 {
			end, _ = dateparse.ParseLocal(req.CreatedAt[1])
		}
		opt.QueryRange["created_at"] = [2]interface{}{start, end}
	}

	_, err := s.base.CheckPermission(ctx)
	if err == nil { // 有权限，则不限页数
		if req.Wd != "" {
			opt.QueryLike["title"] = []interface{}{req.Wd}
			opt.QueryLike["keywords"] = []interface{}{req.Wd}
			opt.QueryLike["description"] = []interface{}{req.Wd}
		}

		if len(req.Status) > 0 {
			opt.QueryIn["status"] = cvt.ToArray(req.Status)
		}
	} else {
		opt.Size = util.LimitRange(opt.Size, 1, 24)
		opt.Page = util.LimitRange(opt.Page, 1, 100)
		if len(req.Status) == 1 && req.Status[0] == model.DocumentStatusConverted {
			opt.QueryIn["status"] = []interface{}{
				model.DocumentStatusConverted,
			}
		} else {
			opt.QueryIn["status"] = []interface{}{
				model.DocumentStatusPending, model.DocumentStatusConverting,
				model.DocumentStatusConverted, model.DocumentStatusFailed,
			}
		}
	}

	if req.Limit > 0 {
		opt.Size = int(req.Limit)
		opt.Page = 1
	}
	resp, err := s.listDocument(opt, ctx)
	if err != nil {
		s.logger.Errorf("err:%s", err.Error())
		ctx.JSON(http.StatusServiceUnavailable,
			types.GinResponse{Code: http.StatusServiceUnavailable, Message: err.Error(), Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CreateDocument 创建文档
// 0. 判断是否有权限
// 1. 同名覆盖：找到该作者上传的相同title和ext的文档，然后用新文件覆盖，同时文档状态改为待转换
// 2. 相同hash的文档如果已经被转换了，则该文档的状态直接改为已转换
// 3. 判断附件ID是否与用户ID匹配，不匹配则跳过该文档
func (s *Service) GinCreateDocument(ctx *gin.Context) {
	user, err := s.auth.CheckAuthorization(ctx.Request.Header.Get("authorization"))
	if err != nil {
		s.logger.Errorf(base.ErrorMessageInvalidToken)
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest,
			Message: base.ErrorMessageInvalidToken, Error: base.ErrorMessageInvalidToken})
		return
	}

	if !s.dbModel.CanIAccessUploadDocument(user.UserId) {
		s.logger.Errorf("没有权限上传文档")
		ctx.JSON(http.StatusForbidden, types.GinResponse{Code: http.StatusForbidden,
			Message: "没有权限上传文档", Error: "没有权限上传文档"})
		return
	}
	var (
		attachmentIds []int64
		attachmentMap = make(map[int64]model.Attachment)
		req           = &CreateDocumentRequest{}
	)
	err = json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil {
		s.logger.Errorf("数据转换错误:%v", err.Error())
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: err.Error(), Error: err.Error()})
		return
	}

	for _, item := range req.Document {
		attachmentIds = append(attachmentIds, item.AttachmentId)
	}

	attachments, _, _ := s.dbModel.GetAttachmentList(&model.OptionGetList{
		Ids:     attachmentIds,
		QueryIn: map[string][]interface{}{"user_id": {user.UserId}},
	})
	if len(attachments) == 0 {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest,
			Message: "文档文件参数attachment_id不正确", Error: "文档文件参数attachment_id不正确"})
		return
	}

	for _, attachment := range attachments {
		attachmentMap[attachment.Id] = attachment
	}

	var (
		documents        []model.Document
		docMapAttachment = make(map[int]int64)
	)

	documentStatus := s.dbModel.GetDefaultDocumentStatus(user.UserId)
	for idx, doc := range req.Document {
		attachment, ok := attachmentMap[doc.AttachmentId]
		if !ok {
			continue
		}

		doc := model.Document{
			Title:    doc.Title,
			Keywords: strings.Join(jieba.SegWords(doc.Title), ","),
			UserId:   user.UserId,
			UUID:     util.GenDocumentMD5UUID(),
			Score:    300,
			Price:    int(doc.Price),
			Size:     attachment.Size,
			Ext:      attachment.Ext,
			Status:   documentStatus,
			Language: doc.Language,
		}
		docMapAttachment[idx] = attachment.Id
		documents = append(documents, doc)
	}

	docs, err := s.dbModel.CreateDocuments(documents, req.CategoryId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, types.GinResponse{Code: http.StatusBadRequest, Message: err.Error(), Error: err.Error()})
		return
	}

	attachIdTypeIdMap := make(map[int64]int64)
	for idx, doc := range docs {
		if attachmentId, ok := docMapAttachment[idx]; ok {
			attachIdTypeIdMap[attachmentId] = doc.Id
		}
	}

	s.dbModel.SetAttachmentTypeId(attachIdTypeIdMap)

	ctx.Status(http.StatusOK)
}
