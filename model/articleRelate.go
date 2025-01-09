package model

import (
	"moredoc/pkg/cvt"
	"moredoc/util/segword/jieba"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ArticleRelate struct {
	Id               int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	ArticleId        int64      `form:"article_id" json:"article_id,omitempty" gorm:"column:article_id;type:bigint(20);size:20;default:0;comment:;index:idx_article_id,unique"`
	RelatedArticleId string     `form:"related_article_id" json:"related_article_id,omitempty" gorm:"column:related_article_id;type:text;comment:;"`
	CreatedAt        *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:创建时间;"`
	UpdatedAt        *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:更新时间;"`
}

func (m *DBModel) GetRelatedArticles(identifier string, fields ...string) (articles []Article, err error) {
	var (
		relate   ArticleRelate
		ids      []int64
		cfg      = m.GetConfigOfSecurity(ConfigSecurityDocumentRelatedDuration)
		keywords []interface{}
		opt      = &OptionGetArticleList{
			WithCount:    false,
			Page:         1,
			Size:         11,
			QueryIn:      make(map[string][]interface{}),
			QueryLike:    make(map[string][]interface{}),
			SelectFields: []string{"id", "title", "keywords", "identifier"},
		}
		isExpired bool
		article   Article
	)
	if len(fields) > 0 {
		opt.SelectFields = fields
	}

	if cfg.DocumentRelatedDuration <= 0 {
		return
	}

	// 文章不存在
	m.db.Select("id", "title", "keywords").Where("identifier = ?", identifier).First(&article)
	if article.Id == 0 {
		return
	}

	err = m.db.Where("article_id = ?", article.Id).First(&relate).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Errorf("GetRelatedArticles", zap.Error(err))
		return
	}

	// 未过期
	if relate.Id > 0 && relate.UpdatedAt.Add(time.Duration(cfg.DocumentRelatedDuration)*time.Hour*24).After(time.Now()) {
		json.Unmarshal([]byte(relate.RelatedArticleId), &ids)
	} else {
		isExpired = true
	}

	if len(ids) == 0 {
		for _, kw := range strings.Split(article.Keywords, ",") {
			kw = strings.TrimSpace(kw)
			if kw == "" {
				continue
			}
			keywords = append(keywords, strings.TrimSpace(kw))
		}
		if len(keywords) == 0 {
			// 从标题中提取关键词
			for _, kv := range jieba.SegWords(article.Title) {
				keywords = append(keywords, kv)
			}
		}
		opt.QueryLike["title"] = keywords
		opt.QueryLike["keywords"] = keywords
		opt.QueryLike["description"] = keywords
	} else {
		opt.QueryIn["id"] = cvt.ToArray(ids)
	}
	articles, _, _ = m.GetArticleList(opt)
	if isExpired && len(articles) > 0 {
		for _, art := range articles {
			if art.Id == article.Id {
				continue
			}
			ids = append(ids, art.Id)
			if len(ids) >= 10 {
				break
			}
		}
		bs, _ := json.Marshal(ids)
		relate.ArticleId = article.Id
		relate.RelatedArticleId = string(bs)
		if relate.Id > 0 {
			m.UpdateByFields(&relate, TableArticleRelate, relate.Id)
		} else {
			m.Create(&relate)
		}
	}
	return
}
