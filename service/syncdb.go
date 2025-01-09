package service

import (
	"database/sql"
	"fmt"
	"moredoc/conf"
	"moredoc/model"
	"moredoc/pkg/logger"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

func SyncDB(cfg *conf.Config, lg logger.Logger) {
	err := checkAndCreateDatabase(cfg.Database.DSN, lg)
	if err != nil {
		lg.Fatalf("checkAndCreateDatabase", zap.Error(err))
		return
	}

	lg.Infof("start syncdb")
	dbModel, err := model.NewDBModel(&cfg.Database, lg)
	if err != nil {
		lg.Fatalf("NewDBModel", zap.Error(err))
		return
	}
	defer dbModel.CloseDB()

	err = dbModel.SyncDB()
	if err != nil {
		lg.Fatalf("SyncDB", zap.Error(err))
		return
	}
	lg.Infof("syncdb success")
}

func checkAndCreateDatabase(dsn string, loggger logger.Logger) (err error) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		loggger.Errorf("ParseDSN", zap.Error(err))
		return
	}

	dbName := cfg.DBName
	if dbName == "" {
		loggger.Errorf("ParseDSN", zap.String("database", "数据库名称不能为空"))
		return
	}

	conn := fmt.Sprintf("%s:%s@tcp(%s)/", cfg.User, cfg.Passwd, cfg.Addr)
	db, err := sql.Open("mysql", conn)
	if err != nil {
		loggger.Errorf("sql.Open", zap.Error(err))
		return
	}

	createDB := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_unicode_ci", cfg.DBName)
	_, err = db.Exec(createDB)
	if err != nil {
		loggger.Errorf("db.Exec", zap.Error(err))
	}
	return
}
