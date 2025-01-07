package model

import (
	"fmt"
	v1 "moredoc/api/v1"
	"moredoc/util"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	TableAdvertisement    = "mnt_advertisement"
	TableArticleRelate    = "mnt_article_relate"
	TableAttachment       = "mnt_attachment"
	TableArticle          = "mnt_article"
	TableArticleCategory  = "mnt_article_category"
	TableBanner           = "mnt_banner"
	TableCategory         = "mnt_category"
	TableComment          = "mnt_comment"
	TableConfig           = "mnt_config"
	TableDocument         = "mnt_document"
	TableDocumentCategory = "mnt_document_category"
	TableDocumentError    = "mnt_document_error"
	TableDocumentRelate   = "mnt_document_relate"
	TableDocumentScore    = "mnt_document_score"
	TableDownload         = "mnt_download"
	TableDynamic          = "mnt_dynamic"
	TableEmailCode        = "mnt_email_code"
	TableFavorite         = "mnt_favorite"
	TableFriendlink       = "mnt_friendlink"
	TableGroup            = "mnt_group"
	TableGroupPermission  = "mnt_group_permission"
	TableLanguage         = "mnt_language"
	TableLogout           = "mnt_logout"
	TableNavigation       = "mnt_navigation"
	TablePermission       = "mnt_permission"
	TablePunishment       = "mnt_punishment"
	TableReport           = "mnt_report"
	TableSearchRecord     = "mnt_search_record"
	TableSign             = "mnt_sign"
	TableUser             = "mnt_user"
	TableUserGroup        = "mnt_user_group"
)

func (m *DBModel) DeleteByIds(ids []int64, st interface{}) error {
	return m.db.Where("id in (?)", ids).Delete(st).Error
}

func (m *DBModel) Create(data interface{}) error {
	return m.db.Create(data).Error
}

func (m *DBModel) GetByID(id int64, value interface{}) error {
	return m.db.Where("id = ?", id).First(value).Error
}

// UpdateUser 更新User，如果需要更新指定字段，则请指定updateFields参数
func (m *DBModel) UpdateByFields(value interface{}, table string, id int64, fields ...string) error {
	db := m.db.Model(value)

	fields = m.FilterValidFields(table, fields...)
	if len(fields) > 0 { // 更新指定字段
		db = db.Select(fields)
	} else {
		db = db.Select(m.GetTableFields(table))
	}

	return db.Where("id = ?", id).Updates(value).Error
}

func (m *DBModel) Count(modelIndex interface{}) int64 {
	var (
		db    = m.db.Model(modelIndex)
		count int64
		err   = db.Count(&count).Error
	)
	if err != nil {
		m.logger.Error("CountFriendlink", zap.Error(err))
	}
	return count
}

type OptionGetList struct {
	Page         int
	Size         int
	WithCount    bool                      // 是否返回总数
	Ids          []int64                   // id列表
	SelectFields []string                  // 查询字段
	QueryRange   map[string][2]interface{} // map[field][]{min,max}
	QueryIn      map[string][]interface{}  // map[field][]{value1,value2,...}
	QueryLike    map[string][]interface{}  // map[field][]{value1,value2,...}
	Sort         []string
	IsRecycle    bool // 是否是回收站模式查询
	IsRecommend  []bool
	FeeType      string // 费用类型：free免费，charge收费
}

type OptionGetArticleList struct {
	Page         int
	Size         int
	WithCount    bool                      // 是否返回总数
	Ids          []interface{}             // id列表
	SelectFields []string                  // 查询字段
	QueryRange   map[string][2]interface{} // map[field][]{min,max}
	QueryIn      map[string][]interface{}  // map[field][]{value1,value2,...}
	QueryLike    map[string][]interface{}  // map[field][]{value1,value2,...}
	Sort         []string
	IsRecycle    bool   // 是否是回收站模式查询
	IsRecommend  []bool // 是否是推荐模式查询
}

// GetDocumentList 获取Document列表
func (m *DBModel) GetDocumentList(opt *OptionGetList) (documentList []Document, total int64, err error) {
	tableDocument := TableDocument + " d"
	db := m.db.Unscoped().Table(tableDocument)
	if opt.IsRecycle {
		// 回收站模式，只根据删除的倒序排序
		opt.Sort = []string{"d.deleted_at desc"}
		db = db.Where("d.deleted_at IS NOT NULL")
	} else {
		db = db.Where("d.deleted_at IS NULL")
	}

	m.logger.Debug("GetDocumentList", zap.Any("opt", opt))

	db = m.generateQueryIn(db, tableDocument, opt.QueryIn)
	db = m.generateQueryLike(db, tableDocument, opt.QueryLike)
	db = m.generateQueryRange(db, tableDocument, opt.QueryRange)
	if len(opt.Ids) > 0 {
		db = db.Where("d.id in (?)", opt.Ids)
	}

	if categoryIds, ok := opt.QueryIn["category_id"]; ok && len(categoryIds) > 0 {
		db = db.Joins("left join "+TableDocumentCategory+" dc on dc.document_id = d.id").Where("dc.category_id in (?)", categoryIds)
	}

	if l := len(opt.IsRecommend); l == 1 {
		if opt.IsRecommend[0] {
			db = db.Where("d.`recommend_at` IS NOT NULL")
		} else {
			db = db.Where("d.`recommend_at` IS NULL")
		}
	}

	if opt.FeeType != "" {
		switch opt.FeeType {
		case "free":
			db = db.Where("d.`price` = ?", 0)
		case "charge":
			db = db.Where("d.`price` > ?", 0)
		}
	}

	if opt.WithCount {
		err = db.Group("d.id").Count(&total).Error
		if err != nil {
			m.logger.Error("GetDocumentList", zap.Error(err))
			return
		}
	}

	opt.SelectFields = m.FilterValidFields(tableDocument, opt.SelectFields...)
	if len(opt.SelectFields) > 0 {
		db = db.Select(opt.SelectFields)
	} else {
		db = db.Select(m.GetTableFields(tableDocument))
	}

	if len(opt.Sort) > 0 {
		db = m.generateQuerySort(db, tableDocument, opt.Sort)
	} else {
		db = db.Order("d.id desc")
	}

	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)
	err = db.Group("d.id").Find(&documentList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetDocumentList", zap.Error(err))
	}
	return
}

// GetAdvertisementList 获取Advertisement列表
func (m *DBModel) GetAdvertisementList(opt *OptionGetList) (advertisementList []Advertisement, total int64, err error) {
	db, total, err := m.queryCond(TableAdvertisement, &Advertisement{}, opt, true)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	db = m.generateQuerySort(db, TableAdvertisement, opt.Sort)
	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Find(&advertisementList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetAdvertisementList", zap.Error(err))
		return nil, total, err
	}
	return advertisementList, total, err
}

func (m *DBModel) queryCond(table string, value interface{}, opt *OptionGetList, withNULL ...bool) (*gorm.DB, int64, error) {
	db := m.db.Model(value)
	db = m.generateQueryRange(db, table, opt.QueryRange, withNULL...)
	db = m.generateQueryIn(db, table, opt.QueryIn)
	db = m.generateQueryLike(db, table, opt.QueryLike)

	if len(opt.Ids) > 0 {
		db = db.Where("id in (?)", opt.Ids)
	}

	var total int64
	if opt.WithCount {
		err := db.Count(&total).Error
		if err != nil {
			m.logger.Error("GetAttachmentList", zap.Error(err))
			return nil, total, err
		}
	}

	opt.SelectFields = m.FilterValidFields(TableAttachment, opt.SelectFields...)
	if len(opt.SelectFields) > 0 {
		db = db.Select(opt.SelectFields)
	}
	return db, total, nil
}

// GetAttachmentList 获取Attachment列表
func (m *DBModel) GetAttachmentList(opt *OptionGetList) (attachmentList []Attachment, total int64, err error) {
	db, total, err := m.queryCond(TableAttachment, &Attachment{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	// TODO: 没有排序参数的话，可以自行指定排序字段
	if len(opt.Sort) > 0 {
		db = m.generateQuerySort(db, TableAttachment, opt.Sort)
	} else {
		db = db.Order("id desc")
	}

	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Find(&attachmentList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetAttachmentList", zap.Error(err))
	}
	return
}

// GetBannerList 获取Banner列表
func (m *DBModel) GetBannerList(opt *OptionGetList) (bannerList []Banner, total int64, err error) {
	db, total, err := m.queryCond(TableBanner, &Banner{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Order("enable desc, sort desc").Find(&bannerList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetBannerList", zap.Error(err))
	}
	return
}

// GetCommentList 获取Comment列表
func (m *DBModel) GetCommentList(opt *OptionGetList) (commentList []Comment, total int64, err error) {
	db, total, err := m.queryCond(TableComment, &Comment{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	db = m.generateQuerySort(db, TableComment, opt.Sort)

	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Find(&commentList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetCommentList", zap.Error(err))
	}
	return
}

// GetDocumentCategoryList 获取DocumentCategory列表
func (m *DBModel) GetDocumentCategoryList(opt *OptionGetList) (documentCategoryList []DocumentCategory, total int64, err error) {
	db, total, err := m.queryCond(TableDocumentCategory, &DocumentCategory{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Find(&documentCategoryList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetDocumentCategoryList", zap.Error(err))
	}
	return
}

// GetDynamicList 获取Dynamic列表
func (m *DBModel) GetDynamicList(opt *OptionGetList) (dynamicList []Dynamic, total int64, err error) {
	db, total, err := m.queryCond(TableDynamic, &Dynamic{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	db = m.generateQuerySort(db, TableDynamic, opt.Sort)
	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)
	err = db.Find(&dynamicList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetDynamicList", zap.Error(err))
	}
	return
}

// GetLanguageList 获取Language列表
func (m *DBModel) GetLanguageList(opt *OptionGetList) (languageList []Language, total int64, err error) {
	db, total, err := m.queryCond(TableLanguage, &Language{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	if len(opt.Sort) == 0 {
		opt.Sort = []string{"enable desc", "sort desc", "id asc"}
	}
	db = m.generateQuerySort(db, TableLanguage, opt.Sort)
	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Find(&languageList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetLanguageList", zap.Error(err))
	}
	return
}

// GetNavigationList 获取Navigation列表
func (m *DBModel) GetNavigationList(opt *OptionGetList) (navigationList []Navigation, total int64, err error) {
	db, total, err := m.queryCond(TableNavigation, &Navigation{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	if len(opt.Sort) == 0 {
		opt.Sort = []string{"sort desc"}
	}
	db = m.generateQuerySort(db, TableNavigation, opt.Sort)

	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Find(&navigationList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetNavigationList", zap.Error(err))
	}
	return
}

// GetPunishmentList 获取Punishment列表
func (m *DBModel) GetPunishmentList(opt *OptionGetList) (punishmentList []Punishment, total int64, err error) {
	db, total, err := m.queryCond(TablePunishment, &Punishment{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	db = m.generateQuerySort(db, TablePunishment, opt.Sort)

	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Find(&punishmentList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetPunishmentList", zap.Error(err))
	}
	return
}

// GetReportList 获取Report列表
func (m *DBModel) GetReportList(opt *OptionGetList) (reportList []*v1.Report, total int64, err error) {
	db, total, err := m.queryCond(TableReport, &Report{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	db = m.generateQuerySort(db, TableReport, opt.Sort)

	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Find(&reportList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetReportList", zap.Error(err))
	}
	return
}

// GetSearchRecordList 获取SearchRecord列表
func (m *DBModel) GetSearchRecordList(opt *OptionGetList) (searchRecordList []SearchRecord, total int64, err error) {
	db, total, err := m.queryCond(TableSearchRecord, &SearchRecord{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}

	db = m.generateQuerySort(db, TableSearchRecord, opt.Sort)

	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Find(&searchRecordList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetSearchRecordList", zap.Error(err))
	}
	return
}

// GetUserList 获取User列表
func (m *DBModel) GetUserList(opt *OptionGetList) (userList []User, total int64, err error) {
	db := m.db.Model(&User{})

	for field, rangeValue := range opt.QueryRange {
		fields := m.FilterValidFields(TableUser, field)
		if len(fields) == 0 {
			continue
		}
		if rangeValue[0] != nil {
			db = db.Where(fmt.Sprintf("%s >= ?", field), rangeValue[0])
		}
		if rangeValue[1] != nil {
			db = db.Where(fmt.Sprintf("%s <= ?", field), rangeValue[1])
		}
	}

	for field, values := range opt.QueryIn {
		if field == "group_id" {
			db = db.Joins(fmt.Sprintf("left JOIN %s ug ON ug.user_id = %s.id", TableUserGroup,
				TableUser)).Where("ug.group_id in (?)", values)
			continue
		}
		fields := m.FilterValidFields(TableUser, field)
		if len(fields) == 0 {
			continue
		}
		db = db.Where(fmt.Sprintf("%s in (?)", field), values)
	}

	db = m.generateQueryLike(db, TableUser, opt.QueryLike)

	if len(opt.Ids) > 0 {
		db = db.Where("id in (?)", opt.Ids)
	}

	if opt.WithCount {
		err = db.Count(&total).Error
		if err != nil {
			m.logger.Error("GetUserList", zap.Error(err))
			return
		}
	}

	opt.SelectFields = m.FilterValidFields(TableUser, opt.SelectFields...)
	if len(opt.SelectFields) > 0 {
		db = db.Select(opt.SelectFields)
	}

	var sorts []string
	if len(opt.Sort) > 0 {
		db = m.generateQuerySort(db, TableUser, opt.Sort)
	} else {
		sorts = append(sorts, "id desc")
	}

	if len(sorts) > 0 {
		db = db.Order(strings.Join(sorts, ","))
	}

	opt.Page = util.LimitMin(opt.Page, 1)
	opt.Size = util.LimitRange(opt.Size, 10, 1000)

	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Find(&userList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetUserList", zap.Error(err))
	}
	return
}

// GetFriendlinkList 获取Friendlink列表
func (m *DBModel) GetFriendlinkList(opt *OptionGetList) (friendlinkList []Friendlink, total int64, err error) {
	db, total, err := m.queryCond(TableFriendlink, &Friendlink{}, opt)
	if err != nil {
		m.logger.Error("err:", zap.Error(err))
		return nil, total, err
	}
	db = db.Offset((opt.Page - 1) * opt.Size).Limit(opt.Size)

	err = db.Order("enable desc,sort desc").Find(&friendlinkList).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		m.logger.Error("GetFriendlinkList", zap.Error(err))
	}
	return
}
