package attachment

import (
	"fmt"
	"moredoc/biz/base"
	"moredoc/middleware/auth"
	"moredoc/model"
	"moredoc/pkg/logger"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var errorHash = map[string]interface{}{
	"code":    http.StatusBadRequest,
	"message": "hash值必须32位",
}

type Service struct {
	dbModel *model.DBModel
	logger  logger.Logger
	base    *base.Base
	au      *auth.Auth
}

func NewService(dbModel *model.DBModel, logger logger.Logger, au *auth.Auth) *Service {
	return &Service{
		dbModel: dbModel,
		logger:  logger,
		base:    base.NewBase(dbModel, au),
		au:      au}
}

func (s *Service) Favicon(ctx *gin.Context) {
	favicon := strings.TrimLeft(s.dbModel.GetConfigOfSystem("favicon").Favicon, "./")
	faviconIco := "favicon.ico"
	if favicon != "" {
		_, err := os.Stat(favicon)
		if err != nil {
			favicon = faviconIco
		}
	} else {
		favicon = faviconIco
	}
	ctx.File(favicon)
}

func (s *Service) Logo(ctx *gin.Context) {
	// 用户设置的logo
	logo := strings.TrimLeft(s.dbModel.GetConfigOfSystem("logo").Logo, "./")
	// 默认logo
	defaultLogo := "dist/static/images/logo.png"

	if logo != "" {
		_, err := os.Stat(logo)
		if err != nil {
			logo = defaultLogo
		}
	} else {
		logo = defaultLogo
	}
	ctx.File(logo)
}

// ViewDocumentPages 浏览文档页面
func (s *Service) ViewDocumentPages(ctx *gin.Context) {
	hash := ctx.Param("hash")
	if len(hash) != 32 {
		ctx.JSON(http.StatusNotFound, errorHash)
		return
	}
	page := strings.TrimLeft(ctx.Param("page"), "./")
	if strings.HasSuffix(page, ".svg") {
		if strings.HasSuffix(page, ".gzip.svg") {
			ctx.Header("Content-Encoding", "gzip")
		}
		ctx.Header("Content-Type", "image/svg+xml")
	}

	file := fmt.Sprintf("documents/%s/%s/%s", strings.Join(strings.Split(hash, "")[:5], "/"), hash, page)
	ctx.File(file)
}

func (s *Service) ViewDocumentCover(ctx *gin.Context) {
	hash := ctx.Param("hash")
	if len(hash) != 32 {
		ctx.JSON(http.StatusNotFound, errorHash)
		return
	}

	file := fmt.Sprintf("documents/%s/%s/cover.png", strings.Join(strings.Split(hash, "")[:5], "/"), hash)
	if len(hash) != 32 {
		ctx.JSON(http.StatusNotFound, map[string]interface{}{"code": http.StatusNotFound, "message": "文件不存在"})
		return
	}
	ctx.File(file)
}

// DownloadDocument 下载文档
func (s *Service) DownloadDocument(ctx *gin.Context) {
	claims := &jwt.StandardClaims{}
	token := ctx.Param("jwt")
	cfg := s.dbModel.GetConfigOfDownload(model.ConfigDownloadSecretKey)
	// 验证JWT是否合法
	jwtToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(cfg.SecretKey), nil
	})
	if err != nil || !jwtToken.Valid || len(claims.Id) != 32 {
		ctx.String(http.StatusBadRequest, "下载链接已失效")
		return
	}

	filename := ctx.Query("filename")
	file := fmt.Sprintf("documents/%s/%s%s", strings.Join(strings.Split(claims.Id, "")[:5], "/"), claims.Id, filepath.Ext(filename))
	ctx.FileAttachment(file, filename)
}
