package config

import (
	"moredoc/biz/base"
	"moredoc/middleware/auth"
	"moredoc/model"
	"moredoc/pkg/logger"
	"moredoc/pkg/types"
	"moredoc/util"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

type Service struct {
	dbModel *model.DBModel
	logger  logger.Logger
	base    *base.Base
}

func NewService(dbModel *model.DBModel, logger logger.Logger, au *auth.Auth) *Service {
	return &Service{dbModel: dbModel,
		logger: logger,
		base:   base.NewBase(dbModel, au)}
}

// 获取系统配置（针对所有用户，只读）
func (s *Service) GinGetSettings(ctx *gin.Context) {
	var (
		res = &Settings{
			System:   &ConfigSystem{},
			Footer:   &ConfigFooter{},
			Security: &ConfigSecurity{},
			Display:  &ConfigDisplay{},
		}

		system = s.dbModel.GetConfigOfSystem()
		err    = util.CopyStruct(&system, res.System)
	)
	defer func() {
		if err != nil {
			s.logger.Errorf("err:%s", err.Error())
			ctx.JSON(http.StatusServiceUnavailable,
				types.GinResponse{Code: http.StatusServiceUnavailable,
					Message: err.Error(),
					Error:   err.Error(),
				})
		}
	}()
	if err != nil {
		return
	}
	system.Analytics = strings.TrimSpace(system.Analytics)
	if system.Analytics != "" {
		gq, errGQ := goquery.NewDocumentFromReader(strings.NewReader(system.Analytics))
		if errGQ == nil {
			var texts []string
			gq.Find("script").Each(func(i int, selection *goquery.Selection) {
				if text := strings.TrimSpace(selection.Text()); text != "" {
					texts = append(texts, text)
				}
			})
			if len(texts) > 0 {
				res.System.Analytics = strings.Join(texts, "\n")
			}
		}
	}
	res.System.Version = util.Version
	res.System.CreditName = s.dbModel.GetConfigOfScore(model.ConfigScoreCreditName).CreditName
	footer := s.dbModel.GetConfigOfFooter()
	err = util.CopyStruct(&footer, res.Footer)
	if err != nil {
		return
	}

	security := s.dbModel.GetConfigOfSecurity()
	err = util.CopyStruct(&security, res.Security)
	if err != nil {
		return
	}

	display := s.dbModel.GetConfigOfDisplay()
	err = util.CopyStruct(&display, res.Display)
	if err != nil {
		return
	}

	langs, _, errMsg := s.dbModel.GetLanguageList(&model.OptionGetList{
		WithCount:    false,
		SelectFields: []string{"id", "language", "code"},
		QueryIn: map[string][]interface{}{
			"enable": {true},
		},
	})
	if errMsg != nil {
		err = errMsg
		return
	}
	util.CopyStruct(&langs, &res.Language)
	ctx.JSON(http.StatusOK, res)
}

// 获取系统配置
func (s *Service) GinGetStats(ctx *gin.Context) {
	res := &Stats{
		UserCount:       s.dbModel.Count(&model.User{}),
		DocumentCount:   s.dbModel.Count(&model.Document{}),
		CategoryCount:   0,
		ArticleCount:    s.dbModel.Count(&model.Article{}),
		CommentCount:    0,
		BannerCount:     0,
		FriendlinkCount: 0,
		Os:              getOSRelease(),
		Version:         util.Version,
		Hash:            util.Hash,
		BuildAt:         util.BuildAt,
	}

	res.UserCount += s.dbModel.GetConfigOfDisplay(model.ConfigDisplayVirtualRegisterCount).VirtualRegisterCount
	_, err := s.base.CheckPermission(ctx)
	if err == nil {
		res.CategoryCount = s.dbModel.Count(&model.Category{})
		res.CommentCount = s.dbModel.Count(&model.Comment{})
		res.BannerCount = s.dbModel.Count(&model.Banner{})
		res.FriendlinkCount = s.dbModel.Count(&model.Friendlink{})
		res.ReportCount = s.dbModel.Count(&model.Report{})
	}
	ctx.JSON(http.StatusOK, res)
}

// 获取系统发行版本信息
func getOSRelease() string {
	var (
		name    string
		version string
	)
	switch runtime.GOOS {
	case "linux":
		content, err := os.ReadFile("/etc/os-release")
		if err != nil {
			return runtime.GOOS
		}
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "NAME=") {
				name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
			}
			if strings.HasPrefix(line, "VERSION_ID=") {
				version = strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
			}
		}
		if name != "" {
			return name + " " + version
		}
	}
	return runtime.GOOS
}
