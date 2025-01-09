package main

import (
	"context"
	"log"
	"moredoc/internal/app"
	"moredoc/server"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"moredoc/pkg/env"
	"moredoc/pkg/logger"
)

func main() {
	lg, err := logger.NewLogger()
	if err != nil {
		log.Print("instantiation logger error: ", err)
		return
	}
	// 接收系统信号, 通过context关系退出服务
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	defer stop()
	ser := server.NewServer(filepath.Base(os.Args[0]), lg)

	// 注册子模块
	ser.AddModule(ctx, app.NewService(ser))

	commands := env.NewArgs(os.Args[1:])
	if !commands.Present() {
		ser.Logger.Errorf("Invalid Params,Missing startup service name")
		return
	}
	m, err := ser.SetupOne(ctx, commands.First(), commands.Tail())
	if err != nil {
		ser.Logger.Errorf("Invalid Params,Get instance error:%s", err.Error())
		return
	}
	if err := m.Run(ctx); err != nil {
		ser.Logger.Errorf("Invalid Params,Get running error:%s", err.Error())
		return
	}
	m.Stop(ctx)
}
