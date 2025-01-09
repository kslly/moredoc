package service

import (
	"moredoc/conf"
	"moredoc/model"
	"moredoc/pkg/logger"

	"go.uber.org/zap"
)

func Reconvert(cfg *conf.Config, logger logger.Logger, ext string, documentId int64) {
	db, err := model.NewDBModel(&cfg.Database, logger)
	if err != nil {
		logger.Fatalf("NewDBModel", zap.Error(err))
		return
	}
	defer db.CloseDB()

	logger.Infof("Reconvert", zap.Int64("documentId", documentId), zap.String("ext", ext))
	db.ReconvertDocoument(documentId, ext)
	logger.Infof("Reconvert", zap.Int64("documentId", documentId), zap.String("ext", ext), zap.String("status", "done!"))
}
