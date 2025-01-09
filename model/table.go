package model

import "go.uber.org/zap"

func (m *DBModel) CreaterTable() error {
	tableModels := []interface{}{
		&Attachment{},
		&AttachmentContent{},
		&Banner{},
		&Category{},
		&Config{},
		&Document{},
		&DocumentCategory{},
		&DocumentError{},
		&DocumentScore{},
		&DocumentRelate{},
		&Download{},
		&Friendlink{},
		&User{},
		&Group{},
		&UserGroup{},
		&Permission{},
		&GroupPermission{},
		&Logout{},
		&Article{},
		&ArticleCategory{},
		&Favorite{},
		&Comment{},
		&Dynamic{},
		&Sign{},
		&Report{},
		&Navigation{},
		&Punishment{},
		&EmailCode{},
		&Advertisement{},
		&SearchRecord{},
		&Language{},
		&ArticleRelate{},
	}

	err := m.db.AutoMigrate(tableModels...)
	if err != nil {
		return err
	}

	// 获取所有数据库表，并把数据库表字段加入到全局map，以便根据指定字段查询数据
	tables, err := m.ShowTables()
	if err != nil {
		m.logger.Errorf("ShowTables", zap.Error(err))
		return err
	}

	for _, table := range tables {
		columns, err := m.showTableColumn(table)
		if err != nil {
			m.logger.Errorf("showTableColumn", zap.Error(err))
			return err
		}

		var fields []string
		for _, col := range columns {
			fields = append(fields, col.Field)
		}

		m.tableFields[table] = fields
		filedsMap := make(map[string]struct{})
		for _, field := range fields {
			filedsMap[field] = struct{}{}
		}
		m.tableFieldsMap[table] = filedsMap
	}
	return nil
}
