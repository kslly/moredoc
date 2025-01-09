package document

import (
	"moredoc/model"
	"moredoc/pkg/cvt"
	"moredoc/util"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *Service) listDocument(opt *model.OptionGetList, ctx *gin.Context) (*ListReply, error) {
	docs, total, err := s.dbModel.GetDocumentList(opt)
	if err != nil {
		return nil, err
	}

	var pbDocs []*Document
	err = util.CopyStruct(&docs, &pbDocs)
	if err != nil {
		s.logger.Errorf("CopyStruct failed", zap.Error(err))
	}

	var (
		docCates            []model.DocumentCategory
		docUsers            []model.User
		docIndexMap         = make(map[int64]int)
		userIndexesMap      = make(map[int64][]int)
		deletedUserIndexMap = make(map[int64][]int)
		docIds              []int64
		userIds             []int64
		display             model.ConfigDisplay
		userId              int64
	)
	user, _ := s.auth.CheckAuthorization(ctx.Request.Header.Get("authorization"))
	if user != nil {
		userId = user.UserId
	}
	isAdmin := s.dbModel.IsAdmin(userId)
	display = s.dbModel.GetConfigOfDisplay(
		model.ConfigDisplayShowDocumentDownloadCount,
		model.ConfigDisplayShowDocumentViewCount,
		model.ConfigDisplayShowDocumentFavoriteCount,
	)
	for i, doc := range pbDocs {
		docIndexMap[doc.Id] = i
		userIndexesMap[doc.UserId] = append(userIndexesMap[doc.UserId], i)
		docIds = append(docIds, doc.Id)
		if doc.UserId > 0 {
			userIds = append(userIds, doc.UserId)
		}
		if doc.DeletedUserId > 0 {
			userIds = append(userIds, doc.DeletedUserId)
			deletedUserIndexMap[doc.DeletedUserId] = append(deletedUserIndexMap[doc.DeletedUserId], i)
		}

		if !display.ShowDocumentDownloadCount && !(isAdmin || doc.UserId == userId) {
			doc.DownloadCount = 0
		}
		if !display.ShowDocumentViewCount && !(isAdmin || doc.UserId == userId) {
			doc.ViewCount = 0
		}
		if !display.ShowDocumentFavoriteCount && !(isAdmin || doc.UserId == userId) {
			doc.FavoriteCount = 0
		}
		pbDocs[i] = doc
	}

	if len(pbDocs) > 0 {
		docCates, _, _ = s.dbModel.GetDocumentCategoryList(&model.OptionGetList{
			WithCount:    false,
			SelectFields: []string{"document_id", "category_id"},
			QueryIn:      map[string][]interface{}{"document_id": cvt.ToArray(docIds)},
		})
		for _, docCate := range docCates {
			pbDocs[docIndexMap[docCate.DocumentId]].CategoryId = append(pbDocs[docIndexMap[docCate.DocumentId]].CategoryId, docCate.CategoryId)
		}

		docUsers, _, _ = s.dbModel.GetUserList(&model.OptionGetList{
			WithCount:    false,
			SelectFields: []string{"id", "username"},
			QueryIn:      map[string][]interface{}{"id": cvt.ToArray(userIds)},
		})

		// 查找文档相关联的附件。对于列表，只返回hash和id，不返回其他字段
		attachments, _, _ := s.dbModel.GetAttachmentList(&model.OptionGetList{
			WithCount:    false,
			SelectFields: []string{"hash", "id", "type_id"},
			QueryIn: map[string][]interface{}{
				"type_id": cvt.ToArray(docIds),
				"type":    {model.AttachmentTypeDocument},
			},
		})

		for _, attachment := range attachments {
			index := docIndexMap[attachment.TypeId]
			pbDocs[index].Attachment = &Attachment{
				Hash: attachment.Hash,
			}
		}

		for docId, errStr := range s.dbModel.GetConvertError(docIds...) {
			index := docIndexMap[docId]
			pbDocs[index].ConvertError = errStr
		}

		for _, docUser := range docUsers {
			indexes := userIndexesMap[docUser.Id]
			for _, index := range indexes {
				pbDocs[index].Username = docUser.Username
			}

			indexes = deletedUserIndexMap[docUser.Id]
			for _, index := range indexes {
				pbDocs[index].DeletedUsername = docUser.Username
			}
		}
	}

	return &ListReply{
		Total:    total,
		Document: pbDocs,
	}, nil
}
