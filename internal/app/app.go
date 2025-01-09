package app

import (
	"context"
	"fmt"
	"moredoc/conf"
	"moredoc/middleware/auth"
	"moredoc/model"
	"moredoc/pkg/env"
	"moredoc/server"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

var (
// dbPath = ".dataReflux.db"
)

type Config struct {
	env.BaseConfig
}

type Service struct {
	config *Config
	db     *model.DBModel
	au     *auth.Auth
	engine *gin.Engine
	jwt    *conf.JWT
	server.BaseModule
	baseStaticPath string // 静态文件存储目录
}

func ginCors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}

// 实例化Service
func NewService(ser *server.Server) (s *Service) {
	cf := new(Config)
	s = &Service{
		config: cf,
		BaseModule: server.BaseModule{
			Name:   "serve",
			Logger: ser.Logger,
		},
	}

	cf.BaseConfig.NewBaseConfig(s.Name, env.UseBasicStorage...)
	// conf.StringVar(&s.baseStaticPath, "baseStaticPath", "/opt/viid/app/", "静态文件默认相对路径")
	cf.StringVar(&s.baseStaticPath, "baseStaticPath", "/home/kongs/go/src/code/moredoc", "静态文件默认相对路径")
	s.jwt = new(conf.JWT)
	cf.StringVar(&s.jwt.Secret, "secret", "kdc", "生成token的密钥")
	cf.Int64Var(&s.jwt.ExpireDays, "expiredays", 1, "生成token的密钥")
	return
}

func (s *Service) Setup(ctx context.Context, args []string) error {
	err := s.config.BaseConfig.ModuleParaser(args)
	if err != nil {
		return err
	}

	s.Ready = true
	s.Debug = s.config.BaseConfig.IsDebug
	s.au = auth.NewAuth(s.jwt)

	err = s.config.Mysql.CheckDB()
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	s.db, err = model.NewDBModel(&conf.Database{
		DSN:     s.config.Mysql.Dns(),
		ShowSQL: s.config.Mysql.ShowSQL,
		MaxIdle: s.config.Mysql.MaxIdle,
		MaxOpen: s.config.Mysql.MaxOpen,
		Prefix:  s.config.Mysql.Prefix,
	}, s.Logger)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}
	// 自动迁移模式（如果表不存在则创建，如果存在则忽略）
	err = s.db.CreaterTable()
	if err != nil {
		return err
	}

	if s.config.IsDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	s.engine = gin.Default()
	// s.engine.Use(ginCors())
	s.engine.Use(gzip.Gzip(gzip.BestCompression, gzip.WithExcludedExtensions([]string{".svg", ".png", ".gif", ".jpeg", ".jpg", ".ico"})), // gzip
		gin.Recovery(), // recovery
		// cors.Default(), // allows all origins
		cors.New(cors.Config{
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
			AllowHeaders:     []string{"*"},
			AllowOrigins:     []string{"*"}, //  Referrer Policy: strict-origin-when-cross-origin
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		// 写一个中间件，如果是user-agent是百度的，就返回一个sitemap.xml
		s.db.SSRMidleware)

	return nil
}

func (s *Service) Run(ctx context.Context) error {
	if !s.Ready {
		return fmt.Errorf("%s: no ready", s.Name)
	}
	// 使用static包提供静态文件服务,将静态文件服务挂载到 /static 路由路径上
	s.engine.Use(static.Serve("/", static.LocalFile(s.baseStaticPath+"/dist/", true)))
	s.engine.Use(static.Serve("/uploads", static.LocalFile("./uploads", true)))
	s.engine.Use(static.Serve("/sitemap", static.LocalFile("./sitemap", true)))

	// 设置路由处理程序，用于处理GET请求
	s.engine.GET("/", func(c *gin.Context) {
		// 在这里可以返回HTML模板或其他内容
		c.HTML(200, "index.html", nil)
	})
	s.RegisterGinRouter()

	err := s.engine.Run(s.config.BindAddr)
	if err != nil {
		s.Logger.Errorf("cannot listen and server: %v", err)
	}
	return err
}

func (s *Service) Stop(ctx context.Context) {
	s.db.CloseDB()
	s.Logger.Infof("%s was stopped", s.Name)
}
