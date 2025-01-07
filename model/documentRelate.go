package model

import (
	"moredoc/util"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DocumentRelate struct {
	Id                int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	DocumentId        int64      `form:"document_id" json:"document_id,omitempty" gorm:"column:document_id;type:bigint(20);size:20;default:0;comment:;index:idx_document_id,unique"`
	RelatedDocumentId string     `form:"related_document_id" json:"related_document_id,omitempty" gorm:"column:related_document_id;type:text;comment:;"`
	CreatedAt         *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt         *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
}

func (m *DBModel) GetRelatedDocuments(documentId int64) (docs []Document, err error) {
	var (
		docRelate DocumentRelate
		docIds    []int64
		cfg       = m.GetConfigOfSecurity(ConfigSecurityDocumentRelatedDuration)
		keywords  []interface{}
		opt       = &OptionGetList{
			WithCount: false,
			Page:      1,
			Size:      11,
			QueryIn:   make(map[string][]interface{}),
			QueryLike: make(map[string][]interface{}),
		}
		isExpired bool
	)

	if cfg.DocumentRelatedDuration <= 0 {
		return
	}

	err = m.db.Where("document_id = ?", documentId).First(&docRelate).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetRelatedDocuments", zap.Error(err))
		return
	}

	// 未过期
	if docRelate.Id > 0 && docRelate.UpdatedAt.Add(time.Duration(cfg.DocumentRelatedDuration)*time.Hour*24).After(time.Now()) {
		json.Unmarshal([]byte(docRelate.RelatedDocumentId), &docIds)
	} else {
		isExpired = true
	}

	if len(docIds) == 0 {
		doc, _ := m.GetDocument(documentId, "id", "title", "keywords")
		if doc.Id > 0 {
			for _, kw := range strings.Split(doc.Keywords, ",") {
				keywords = append(keywords, strings.TrimSpace(kw))
			}
			opt.QueryLike["title"] = keywords
			opt.QueryLike["keywords"] = keywords
			opt.QueryLike["description"] = keywords
		}
	} else {
		opt.QueryIn["id"] = util.Slice2Interface(docIds)
	}
	docs, _, _ = m.GetDocumentList(opt)
	if !isExpired {
		return
	}
	for _, doc := range docs {
		if documentId == doc.Id {
			continue
		}
		docIds = append(docIds, doc.Id)
		if len(docIds) >= 10 {
			break
		}
	}
	bs, _ := json.Marshal(docIds)
	docRelate.DocumentId = documentId
	docRelate.RelatedDocumentId = string(bs)
	if docRelate.Id > 0 {
		m.UpdateByFields(&docRelate, TableDocumentRelate, docRelate.Id)
	} else {
		m.Create(&docRelate)
	}
	return
}
