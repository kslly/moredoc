package app

import (
	"moredoc/biz/advertisement"
	"moredoc/biz/article"
	"moredoc/biz/attachment"
	"moredoc/biz/banner"
	"moredoc/biz/category"
	"moredoc/biz/config"
	"moredoc/biz/document"
	"moredoc/biz/friendlink"
	"moredoc/biz/navigation"
	"moredoc/biz/user"

	"github.com/gin-gonic/gin"
)

// RegisterGinRouter 注册gin路由
func (s *Service) RegisterGinRouter() {
	attachment := attachment.NewService(s.db, s.Logger, s.au)

	s.engine.GET("/favicon.ico", attachment.Favicon)
	s.engine.GET("/static/images/logo.png", attachment.Logo)
	s.engine.GET("/sitemap.xml", func(ctx *gin.Context) {
		ctx.File("./sitemap/sitemap.xml")
	})

	s.engine.GET("/view/page/:hash/:page", attachment.ViewDocumentPages)
	s.engine.GET("/view/cover/:hash", attachment.ViewDocumentCover)
	s.engine.GET("/download/:jwt", attachment.DownloadDocument)

	checkPermissionGroup := s.engine.Group("/api/v1/upload")
	checkPermissionGroup.Use(s.au.AuthGin())
	{
		checkPermissionGroup.POST("avatar", attachment.UploadAvatar)
		checkPermissionGroup.POST("config", attachment.UploadConfig)
		checkPermissionGroup.POST("banner", attachment.UploadBanner)
		checkPermissionGroup.POST("document", attachment.UploadDocument)
		checkPermissionGroup.POST("category", attachment.UploadCategory)
		checkPermissionGroup.POST("article", attachment.UploadArticle)
	}

	conf := config.NewService(s.db, s.Logger, s.au)
	s.engine.GET("/api/v1/settings", conf.GinGetSettings)
	s.engine.GET("/api/v1/stats", conf.GinGetStats)

	nav := navigation.NewService(s.db, s.Logger)
	s.engine.GET("/api/v1/navigation/list", nav.GinListNavigation)

	doc := document.NewService(s.db, s.Logger, s.au)
	s.engine.GET("/api/v1/document/home", doc.GinListDocumentForHome)
	s.engine.GET("/api/v1/document/list", doc.GinListDocument)
	s.engine.POST("/api/v1/document", doc.GinCreateDocument)

	adv := advertisement.NewService(s.db, s.Logger)
	s.engine.GET("/api/v1/advertisement/position", adv.GinGetAdvertisementByPosition)

	article := article.NewService(s.db, s.Logger, s.au)
	s.engine.GET("/api/v1/article/list", article.GinListArticle)

	fl := friendlink.NewService(s.db, s.Logger, s.au)
	s.engine.GET("/api/v1/friendlink/list", fl.GinListFriendlink)

	banner := banner.NewService(s.db, s.Logger, s.au)
	s.engine.GET("/api/v1/banner/list", banner.GinListBanner)

	cate := category.NewService(s.db, s.Logger, s.au)
	s.engine.GET("/api/v1/category/list", cate.GinListCategory)

	user := user.NewService(s.db, s.Logger, s.au)
	s.engine.GET("/api/v1/user/captcha", user.GinGetUserCaptcha)
	s.engine.POST("/api/v1/user/login", user.GinLogin)
	s.engine.GET("/api/v1/user", user.GinGetUser)
	s.engine.GET("/api/v1/user/permission", user.GinGetUserPermissions)
	s.engine.GET("/api/v1/user/sign", user.GinGetSignedToday)
	s.engine.DELETE("/api/v1/user/logout", user.GinLogout)
	s.engine.GET("/api/v1/user/caniuploaddocument", user.GinCanIUploadDocument)
}
