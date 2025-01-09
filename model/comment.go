package model

import (
	// "fmt"
	// "strings"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	CommentStatusPending  = iota // 待审核
	CommentStatusApproved        // 已审核
	CommentStatusRejected        // 已拒绝
)

type Comment struct {
	Id           int64      `form:"id" json:"id,omitempty" gorm:"primaryKey;autoIncrement;column:id;comment:;"`
	UserId       int64      `form:"user_id" json:"user_id,omitempty" gorm:"column:user_id;type:bigint(20);size:20;index:idx_user_id;comment:发布评论的用户;"`
	ParentId     int64      `form:"parent_id" json:"parent_id,omitempty" gorm:"column:parent_id;type:bigint(20);size:20;default:0;comment:上级ID;index:idx_parent_id;"`
	Content      string     `form:"content" json:"content,omitempty" gorm:"column:content;type:text;comment:评论内容;"`
	DocumentId   int64      `form:"document_id" json:"document_id,omitempty" gorm:"column:document_id;type:bigint(20);size:20;default:0;comment:兼容字段，文档ID或文章ID;index:idx_document_id;"`
	Status       int8       `form:"status" json:"status,omitempty" gorm:"column:status;type:tinyint(4);size:4;default:0;comment:0 待审，1过审，2拒绝;"`
	CommentCount int        `form:"comment_count" json:"comment_count,omitempty" gorm:"column:comment_count;type:int(11);size:11;default:0;comment:评论数量;"`
	IP           string     `form:"ip" json:"ip,omitempty" gorm:"column:ip;type:varchar(64);size:64;default:'';comment:IP地址;"`
	CreatedAt    *time.Time `form:"created_at" json:"created_at,omitempty" gorm:"column:created_at;type:datetime;comment:评论时间;index:idx_created_at;"`
	UpdatedAt    *time.Time `form:"updated_at" json:"updated_at,omitempty" gorm:"column:updated_at;type:datetime;comment:评论更新时间;"`
	Type         int32      `form:"type" json:"type,omitempty" gorm:"column:type;type:int(11);size:11;default:0;comment:评论类型，0表示文档评论，1表示文章评论;index:idx_type"` // 枚举见CategoryType
}

// CreateComment 创建Comment
func (m *DBModel) CreateDocumentComment(comment *Comment) (err error) {
	doc := &Document{}
	m.db.Where("id = ?", comment.DocumentId).Select("id", "title", "user_id", "uuid").Find(doc)

	tx := m.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = tx.Create(comment).Error
	if err != nil {
		m.logger.Errorf("CreateComment", zap.Error(err))
		return
	}

	// 文档评论数+1
	err = tx.Model(&Document{}).Where("id = ?", comment.DocumentId).Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	if err != nil {
		m.logger.Errorf("CreateComment", zap.Error(err))
		return
	}

	// 用户评论数+1
	err = tx.Model(&User{}).Where("id = ?", comment.UserId).Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	if err != nil {
		m.logger.Errorf("CreateComment", zap.Error(err))
		return
	}

	dynamic := &Dynamic{
		UserId: comment.UserId,
		Type:   DynamicTypeComment,
	}
	// 更新上级评论的评论数
	if comment.ParentId > 0 {
		err = tx.Model(&Comment{}).Where("id = ?", comment.ParentId).Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error
		if err != nil {
			m.logger.Errorf("CreateComment", zap.Error(err))
			return
		}
		dynamic.Content = fmt.Sprintf(`在文档《<a href="/document/%s">%s</a>》中回复了评论`, doc.UUID, html.EscapeString(doc.Title))
	} else {
		dynamic.Content = fmt.Sprintf(`评论了文档《<a href="/document/%s">%s</a>》`, doc.UUID, html.EscapeString(doc.Title))
	}

	// 增加评论动态
	err = tx.Create(dynamic).Error
	if err != nil {
		m.logger.Errorf("CreateComment", zap.Error(err))
		return
	}

	// 是否可以获得积分奖励
	canRewarded := false
	cfgScore := m.GetConfigOfScore(ConfigScoreDocumentCommented, ConfigScoreDocumentCommentedLimit)
	// 用户文档自评，积分不增加
	if comment.UserId != doc.UserId && cfgScore.DocumentCommented > 0 && cfgScore.DocumentCommentedLimit > 0 {
		var count int64
		errCount := tx.Model(&Comment{}).Where("document_id = ? and created_at > ?", comment.DocumentId, time.Now().Format("2006-01-02")).Count(&count).Error
		if errCount != nil && errCount != gorm.ErrRecordNotFound {
			m.logger.Errorf("CreateComment", zap.Error(errCount))
			return
		}

		// 用户积分增加
		if int32(count) < cfgScore.DocumentCommentedLimit {
			err = tx.Model(&User{}).Where("id = ?", doc.UserId).Update("credit_count", gorm.Expr("credit_count + ?", cfgScore.DocumentCommented)).Error
			if err != nil {
				m.logger.Errorf("CreateComment", zap.Error(err))
				return
			}
			canRewarded = true
		}
	}

	// 被评论的文档作者增加动态和积分
	newDynamic := &Dynamic{
		UserId:  doc.UserId,
		Type:    DynamicTypeComment,
		Content: fmt.Sprintf(`您上传的文档《<a href="/document/%s">%s</a>》被评论了`, doc.UUID, html.EscapeString(doc.Title)),
	}
	if canRewarded {
		newDynamic.Content = fmt.Sprintf(`您上传的文档《<a href="/document/%s">%s</a>》被评论了，获得 %d %s奖励`, doc.UUID, html.EscapeString(doc.Title), cfgScore.DocumentCommented, cfgScore.CreditName)
	}

	err = tx.Create(newDynamic).Error
	if err != nil {
		m.logger.Errorf("CreateComment", zap.Error(err))
		return
	}
	return
}

// 创建文章评论
func (m *DBModel) CreateArticleComment(comment *Comment) (err error) {
	article := &Article{}
	m.db.Where("id = ?", comment.DocumentId).Select("id", "title", "user_id", "identifier").Find(article)
	if article.Id == 0 {
		return errors.New("文章不存在")
	}

	tx := m.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = tx.Create(comment).Error
	if err != nil {
		m.logger.Errorf("CreateComment", zap.Error(err))
		return
	}

	// 文档评论数+1
	err = tx.Model(article).Where("id = ?", comment.DocumentId).Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	if err != nil {
		m.logger.Errorf("CreateComment", zap.Error(err))
		return
	}

	// 用户评论数+1
	err = tx.Model(&User{}).Where("id = ?", comment.UserId).Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	if err != nil {
		m.logger.Errorf("CreateComment", zap.Error(err))
		return
	}

	dynamic := &Dynamic{
		UserId: comment.UserId,
		Type:   DynamicTypeComment,
	}
	// 更新上级评论的评论数
	if comment.ParentId > 0 {
		err = tx.Model(&Comment{}).Where("id = ?", comment.ParentId).Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error
		if err != nil {
			m.logger.Errorf("CreateComment", zap.Error(err))
			return
		}
		dynamic.Content = fmt.Sprintf(`在文章《<a href="/article/%s">%s</a>》中回复了评论`, article.Identifier, html.EscapeString(article.Title))
	} else {
		dynamic.Content = fmt.Sprintf(`评论了文档《<a href="/article/%s">%s</a>》`, article.Identifier, html.EscapeString(article.Title))
	}

	// 增加评论动态
	err = tx.Create(dynamic).Error
	if err != nil {
		m.logger.Errorf("CreateComment", zap.Error(err))
		return
	}
	return
}

// DeleteComment 删除数据
// 删除评论之后，对应文档的评论数量也要减少，对应的父级文档评论数量也要减少，用户评论数量也要减少
func (m *DBModel) DeleteComment(ids []int64, limitUserId ...int64) (err error) {
	tx := m.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var (
		comments []Comment
		user     = &User{}
		document = &Document{}
	)
	cond := []string{"id in (?)"}
	args := []interface{}{ids}
	if len(limitUserId) > 0 {
		cond = append(cond, "user_id in (?)")
		args = append(args, limitUserId)
	}
	condStr := strings.Join(cond, " and ")
	tx.Where(condStr, args...).Select("id", "parent_id", "document_id", "user_id").Find(&comments)
	if len(comments) == 0 {
		err = errors.New("评论不存在或没有权限删除")
		return err
	}

	err = tx.Where(condStr, args...).Delete(&Comment{}).Error
	if err != nil {
		m.logger.Errorf("DeleteComment", zap.Error(err))
		return
	}

	for _, comment := range comments {
		// 更新文档评论数
		err = tx.Model(document).Where("id = ?", comment.DocumentId).UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1)).Error
		if err != nil {
			m.logger.Errorf("DeleteComment", zap.Error(err))
			return
		}

		// 更新父级评论数
		if comment.ParentId > 0 {
			err = tx.Model(&comment).Where("id = ?", comment.ParentId).UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1)).Error
			if err != nil {
				m.logger.Errorf("DeleteComment", zap.Error(err))
				return
			}
		}

		// 更新用户评论数
		err = tx.Model(user).Where("id = ?", comment.UserId).UpdateColumn("comment_count", gorm.Expr("comment_count - ?", 1)).Error
		if err != nil {
			m.logger.Errorf("DeleteComment", zap.Error(err))
			return
		}
	}

	return
}

func (m *DBModel) UpdateCommentStatus(ids []int64, status int32) error {
	return m.db.Model(&Comment{}).Where("id in (?) and status != ?", ids, status).Update("status", status).Error
}

func (m *DBModel) GetDefaultCommentStatus(userId int64) int {
	// 默认待审核
	status := CommentStatusPending

	var group Group
	// 查询用户所在用户组，是否评论不需要审核
	err := m.db.Select("g.id").Where("ug.user_id = ? and g.enable_comment_approval = ?", userId, false).Table(TableGroup + " g").Joins(
		"left join " + TableUserGroup + " ug on g.id=ug.group_id",
	).Find(&group).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Errorf("GetDefaultCommentStatus", zap.Error(err))
		return status
	}

	m.logger.Debugf("GetDefaultCommentStatus", zap.Any("group", group))

	if group.Id > 0 {
		status = CommentStatusApproved
	}
	return status
}
