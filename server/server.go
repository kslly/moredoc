package server

import (
	"context"
	"fmt"
	"moredoc/pkg/logger"
)

// BaseModule 定义每个服务模块都需要的一些属性
type BaseModule struct {
	Logger logger.Logger

	Name  string
	Ready bool
	Debug bool
}

// GetName 获取服务模块的名称
func (bm *BaseModule) GetName() string {
	return bm.Name
}

// IsReady 返回服务模块是否ready
func (bm *BaseModule) IsReady() bool {
	return bm.Ready
}

func (bm *BaseModule) SetReady(f bool) {
	bm.Ready = f
}

func (bm *BaseModule) IsDebug() bool {
	return bm.Debug
}

func (bm *BaseModule) SetDebug(f bool) {
	bm.Debug = f
}

// Server 主运行实例
type Server struct {
	Logger  logger.Logger
	modules map[string]Module
	Name    string
}

// Module 模块服务
type Module interface {
	GetName() string
	Setup(context.Context, []string) error
	IsReady() bool
	SetReady(bool)
	Run(context.Context) error
	Stop(context.Context)
	IsDebug() bool
}

// NewApp 实例化主运行实例
func NewServer(appName string, lg logger.Logger) *Server {
	app := &Server{
		Name:    appName,
		modules: map[string]Module{},
		Logger:  lg,
	}
	return app
}

// Setup 对所有注册的模块初始化
func (s *Server) Setup(ctx context.Context, args []string) error {
	for _, item := range s.modules {
		if !item.IsReady() {
			if err := item.Setup(ctx, args); err != nil {
				s.Logger.Errorf("module setup error,service：%s,err:%s", item.GetName(), err.Error())
			}
		}
		item.SetReady(true)
		s.Logger.Infof("module ready,service:%s", item.GetName)
	}
	return nil
}

// SetupOne 对指定模块初始化
func (s *Server) SetupOne(ctx context.Context, name string, args []string) (Module, error) {
	m, ok := s.modules[name]
	if !ok {
		return nil, fmt.Errorf("no registered module '%s'", name)
	}
	if ok && !m.IsReady() {
		if err := m.Setup(ctx, args); err != nil {
			return nil, err
		}
		if m.IsDebug() {
			s.Logger.SetLevel(logger.DebugLevel)
		}
		m.SetReady(true)
		s.Logger.Infof("module ready,service:%s", m.GetName())
	}
	return m, nil
}

// AddModule 对指定模块进行初始化
func (s *Server) AddModule(ctx context.Context, m Module) {
	s.modules[m.GetName()] = m
	s.Logger.Infof("module registered,service:%s", m.GetName())
}

// Run 运行所有已注册的模块
func (s *Server) Run(ctx context.Context) {
	for _, item := range s.modules {
		if item.IsReady() {
			if err := item.Run(ctx); err != nil {
				s.Logger.Errorf("err:%s", err.Error())
				break
			}
			s.Logger.Infof("module running,service%s", item.GetName())
		}
	}
}

// Stop 停止所有已注册的模块
func (s *Server) Stop(ctx context.Context) {
	for _, item := range s.modules {
		item.Stop(ctx)
		s.Logger.Infof("module stoped,service%s", item.GetName())
	}
}

// StopOne 停止指定模块
func (s *Server) StopOne(ctx context.Context, name string) {
	m, ok := s.modules[name]
	if ok {
		m.Stop(ctx)
		s.Logger.Infof("module stoped,service%s", m.GetName())
	}
}

// GetModules 获取注册到主进程中的服务模块名
func (s *Server) GetModules() (modules []string) {
	for _, item := range s.modules {
		modules = append(modules, item.GetName())
	}
	return
}
