package biz

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	pb "moredoc/api/v1"
	"moredoc/middleware/auth"
	"moredoc/model"
	"moredoc/pkg/cvt"
	"moredoc/pkg/logger"
	"moredoc/util"
	"moredoc/util/device"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ConfigAPIService struct {
	pb.UnimplementedConfigAPIServer
	dbModel *model.DBModel
	logger  logger.Logger
}

func NewConfigAPIService(dbModel *model.DBModel, logger logger.Logger) (service *ConfigAPIService) {
	return &ConfigAPIService{dbModel: dbModel, logger: logger}
}

func (s *ConfigAPIService) checkPermission(ctx context.Context) (userClaims *auth.UserClaims, err error) {
	return checkGRPCPermission(s.dbModel, ctx)
}

// UpdateConfig 更新配置
func (s *ConfigAPIService) UpdateConfig(ctx context.Context, req *pb.Configs) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	var cfgs []*model.Config
	err = util.CopyStruct(req.Config, &cfgs)
	if err != nil {
		s.logger.Errorf("util.CopyStruct", zap.Any("req", req), zap.Any("cfgs", cfgs), zap.Error(err))
		fmt.Println(err.Error())
	}

	doesUpdateSEO := false
	isEmail := false
	for idx, cfg := range cfgs {
		if cfg.Category == model.ConfigCategoryRelease {
			return nil, status.Error(codes.PermissionDenied, "不允许修改此类配置！")
		}

		if cfg.Value == "******" {
			// 6个星号，不修改原值
			exist, _ := s.dbModel.GetConfigByNameCategory(cfg.Name, cfg.Category)
			cfgs[idx].Value = exist.Value
		}
		isEmail = isEmail || cfg.Category == model.ConfigCategoryEmail
		if cfg.Category == model.ConfigCategorySystem {
			doesUpdateSEO = true
		}
	}

	err = s.dbModel.UpdateConfigs(cfgs, "value")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if isEmail {
		cfgEmail := s.dbModel.GetConfigOfEmail(model.ConfigEmailEnable, model.ConfigEmailTestEmail)
		if cfgEmail.Enable && cfgEmail.TestEmail != "" {
			err = s.dbModel.SendMail("测试邮件", cfgEmail.TestEmail, "这是一封测试邮件")
			if err != nil {
				return nil, status.Error(codes.Internal, "邮件发送失败:"+err.Error())
			}
		}
	}

	if doesUpdateSEO {
		s.dbModel.InitSEO()
	}

	return &emptypb.Empty{}, nil
}

// ListConfig 查询配置
func (s *ConfigAPIService) ListConfig(ctx context.Context, req *pb.ListConfigRequest) (*pb.Configs, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	opt := &model.OptionGetList{
		QueryIn: map[string][]interface{}{
			"category": cvt.ToArray(req.Category),
		},
	}

	configs, err := s.dbModel.GetConfigList(opt)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for idx, cfg := range configs {
		if cfg.IsSecret && cfg.Value != "" {
			configs[idx].Value = "******"
		}
	}

	var pbConfigs []*pb.Config
	util.CopyStruct(&configs, &pbConfigs)

	return &pb.Configs{Config: pbConfigs}, nil
}

// UpdateSitemap 更新站点地图
func (s *ConfigAPIService) UpdateSitemap(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	_, err := s.checkPermission(ctx)
	if err != nil {
		return nil, err
	}

	err = s.dbModel.UpdateSitemap()
	if err != nil {
		s.logger.Errorf("UpdateSitemap", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *ConfigAPIService) GetEnvs(ctx context.Context, req *emptypb.Empty) (res *pb.Envs, err error) {
	res = &pb.Envs{}
	_, errPermission := s.checkPermission(ctx)
	if errPermission != nil {
		return
	}
	envs := []*pb.EnvDependent{
		{
			Name:        "LibreOffice",
			Description: "LibreOffice是由文档基金会开发的自由及开放源代码的办公套件。魔豆文库用于将office等文档转为pdf。",
			Cmd:         "soffice",
			IsRequired:  true,
		},
		{
			Name:        "Calibre",
			Description: "calibre是一个自由开源的电子书软件套装。魔豆文库用于将epub、mobi等电子书转为pdf。",
			Cmd:         "ebook-convert",
			IsRequired:  true,
		},
		{
			// mupdf
			Name:        "MuPDF",
			Description: "MuPDF是一款以C语言编写的自由及开放源代码软件库，是PDF和XPS解析和渲染引擎。魔豆文库用于将PDF转为svg、png等图片。",
			Cmd:         "mutool",
			IsRequired:  true,
		},
		{
			// mupdf
			Name:        "ImageMagick",
			Description: "ImageMagick是一个用于创建、编辑、合成和转换位图图像的自由软件套件。魔豆文库用于将PDF转为svg、png、jpg等图片。",
			Cmd:         "convert",
			IsRequired:  true,
		},
		{
			// inkscape
			Name:        "Inkscape",
			Description: "Inkscape是一个自由开源的矢量图形编辑器。在mupdf处理PDF出现兼容问题失败时，自动切换inkscape来处理。",
			Cmd:         "inkscape",
			IsRequired:  true,
		},
		{
			Name:        "SVGO",
			Description: "SVGO 是一个基于 Node.js 的工具，用于优化 SVG 矢量图形文件。魔豆文库用于压缩svg图片大小。",
			Cmd:         "svgo",
			IsRequired:  false,
		},
		{
			Name:        "PM2",
			Description: "PM2是JavaScript运行时Node.js的进程管理器。用于做魔豆文库的系统守护进程。Windows下建议使用PM2。",
			Cmd:         "pm2",
			IsRequired:  false,
		}, {
			Name:        "Supervisor",
			Description: "Supervisor是一个客户端/服务器系统，用于监视进程状态，当进程不再运行时自动重启它们。用于做魔豆文库的系统守护进程。Linux下建议使用Supervisor。",
			Cmd:         "supervisorctl",
			IsRequired:  false,
		},
	}
	for i := 0; i < len(envs); i++ {
		now := time.Now()
		err := util.CheckCommandExists(envs[i].Cmd)
		envs[i].IsInstalled = err == nil
		envs[i].CheckedAt = &now
		envs[i].Version = util.GetCommandVersion(envs[i].Cmd)
		if err != nil {
			envs[i].Error = err.Error()
		}
	}

	// MySQL 的group by 查询
	yes, sqlMode := s.dbModel.IsSupportGroupBy()
	if !yes {
		now := time.Now()
		envs = append(envs, &pb.EnvDependent{
			Name:        "GroupBy",
			IsRequired:  true,
			IsInstalled: false, // 表示
			Description: "魔豆文库的部分查询需要MySQL的group by功能，如果不支持，将导致部分功能无法正常使用和出错。",
			Error: fmt.Sprintf(
				`您当前MySQL数据库不支持 group by 查询，请修改数据库的 sql_mode 配置，去掉 ONLY_FULL_GROUP_BY。<br> 如将当前的 <br/>sql_mode=%s <br/>修改为<br/>sql_mode=%s`,
				sqlMode, strings.ReplaceAll(sqlMode, "ONLY_FULL_GROUP_BY,", ""),
			),
			CheckedAt: &now,
		})
	}

	res.Envs = envs
	return
}

func (s *ConfigAPIService) GetDeviceInfo(ctx context.Context, req *emptypb.Empty) (res *pb.DeviceInfo, err error) {
	_, err = s.checkPermission(ctx)
	if err != nil {
		return
	}

	res = &pb.DeviceInfo{
		Cpu:    &pb.CPUInfo{},
		Memory: &pb.MemoryInfo{},
	}
	cpu := device.GetCPU()

	err = util.CopyStruct(&cpu, res.Cpu)
	if err != nil {
		s.logger.Errorf("util.CopyStruct", zap.Any("cpu", cpu), zap.Any("res.Cpu", res.Cpu), zap.Error(err))
		return
	}
	res.Cpu.Cores = int32(runtime.NumCPU())

	mem := device.GetMemory()
	err = util.CopyStruct(&mem, res.Memory)
	if err != nil {
		s.logger.Errorf("util.CopyStruct", zap.Any("mem", mem), zap.Any("res.Memory", res.Memory), zap.Error(err))
		return
	}

	res.Memory.Free = res.Memory.Total - res.Memory.Used

	disks := device.GetDisk()
	if len(disks) > 0 {
		for _, disk := range disks {
			pbDisk := &pb.DiskInfo{}
			err = util.CopyStruct(&disk, pbDisk)
			if err != nil {
				s.logger.Errorf("util.CopyStruct", zap.Any("disk", disk), zap.Any("res.Disk", res.Disk), zap.Error(err))
				return
			}
			res.Disk = append(res.Disk, pbDisk)
		}
	}
	return res, nil
}

// 设置sql_mode
func (s *ConfigAPIService) SetSQLMode(ctx context.Context, req *emptypb.Empty) (res *emptypb.Empty, err error) {
	var userClaims *auth.UserClaims
	userClaims, err = checkGRPCLogin(s.dbModel, ctx)
	if err != nil {
		return
	}

	if userClaims.UserId != 1 {
		return nil, status.Error(codes.PermissionDenied, "只有用户ID为1的用户才有权限执行此操作！")
	}

	err = s.dbModel.SetSQLMode()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// 获取最近发布版本
func (s *ConfigAPIService) GetLatestRelease(ctx context.Context, req *emptypb.Empty) (res *pb.Release, err error) {
	var userClaims *auth.UserClaims
	userClaims, err = checkGRPCLogin(s.dbModel, ctx)
	if err != nil {
		return
	}
	if !s.dbModel.IsAdmin(userClaims.UserId) {
		return nil, status.Error(codes.PermissionDenied, "只有管理员才有权限执行此操作！")
	}

	latestRelease := s.dbModel.GetConfigOfRelease()
	res = &pb.Release{
		TagName:   latestRelease.TagName,
		Name:      latestRelease.Name,
		Body:      latestRelease.Body,
		Source:    latestRelease.Source,
		Ignore:    latestRelease.Ignore,
		ReleaseAt: latestRelease.ReleaseAt,
		Current:   util.Version,
	}
	return
}

// 更新最新发布版本
func (s *ConfigAPIService) RefreshLatestRelease(ctx context.Context, req *emptypb.Empty) (res *pb.Release, err error) {
	var userClaims *auth.UserClaims
	userClaims, err = checkGRPCLogin(s.dbModel, ctx)
	if err != nil {
		return
	}
	if !s.dbModel.IsAdmin(userClaims.UserId) {
		return nil, status.Error(codes.PermissionDenied, "只有管理员才有权限执行此操作！")
	}

	err = s.dbModel.RefreshLatestRelease()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	latestRelease := s.dbModel.GetConfigOfRelease()
	res = &pb.Release{
		TagName:   latestRelease.TagName,
		Name:      latestRelease.Name,
		Body:      latestRelease.Body,
		Source:    latestRelease.Source,
		Ignore:    latestRelease.Ignore,
		ReleaseAt: latestRelease.ReleaseAt,
		Current:   util.Version,
	}
	return
}

// 忽略版本
func (s *ConfigAPIService) IgnoreRelease(ctx context.Context, req *pb.Release) (res *emptypb.Empty, err error) {
	var userClaims *auth.UserClaims
	userClaims, err = checkGRPCLogin(s.dbModel, ctx)
	if err != nil {
		return
	}
	if !s.dbModel.IsAdmin(userClaims.UserId) {
		return nil, status.Error(codes.PermissionDenied, "只有管理员才有权限执行此操作！")
	}

	err = s.dbModel.IgnoreRelease(req.TagName)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// 设置版本获取源
func (s *ConfigAPIService) SetReleaseSource(ctx context.Context, req *pb.Release) (res *emptypb.Empty, err error) {
	var userClaims *auth.UserClaims
	userClaims, err = checkGRPCLogin(s.dbModel, ctx)
	if err != nil {
		return
	}
	if !s.dbModel.IsAdmin(userClaims.UserId) {
		return nil, status.Error(codes.PermissionDenied, "只有管理员才有权限执行此操作！")
	}

	err = s.dbModel.SetReleaseSource(req.Source)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}
