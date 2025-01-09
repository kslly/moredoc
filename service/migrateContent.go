package service

import (
	"moredoc/conf"
	"moredoc/model"
	"moredoc/pkg/logger"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func MigrateContent(cfg *conf.Config, lg logger.Logger) {
	lg.Infof("start...")
	dbModel, err := model.NewDBModel(&cfg.Database, lg)
	if err != nil {
		lg.Fatalf("NewDBModel", zap.Error(err))
		return
	}

	page := 1
	size := 100
	for {
		var attachments []model.Attachment
		err := dbModel.DB().Where("type = ?", model.AttachmentTypeDocument).Select("hash", "path").Group("hash").Offset((page - 1) * size).Limit(size).Order("id asc").Find(&attachments).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			lg.Errorf("查询附件失败", zap.Error(err))
			break
		}
		if len(attachments) == 0 {
			break
		}

		for _, attachment := range attachments {
			lg.Infof("MigrateContent", zap.String("hash", attachment.Hash))
			textFile := strings.TrimLeft(strings.TrimSuffix(attachment.Path, filepath.Ext(attachment.Path)), "./") + "/content.txt"
			content, err := os.ReadFile(textFile)
			if err != nil {
				lg.Debugf("读取文本内容失败，跳过...", zap.String("hash", attachment.Hash), zap.Error(err), zap.String("textFile", textFile))
				continue
			}
			contentStr := string(content)
			replacer := strings.NewReplacer("\r", " ", "\n", " ", "\t", " ")
			contentStr = strings.TrimSpace(replacer.Replace(contentStr))
			err = dbModel.SetAttachmentContentByHash(attachment.Hash, []byte(contentStr))
			if err != nil {
				lg.Errorf("SetAttachmentContentByHash", zap.Error(err), zap.String("hash", attachment.Hash))
			} else {
				// 删除原 txt 文件
				os.Remove(textFile)
			}
		}
		page++
	}

	lg.Infof("migrate content done!")
}
