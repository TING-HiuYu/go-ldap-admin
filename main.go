package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/eryajf/go-ldap-admin/logic"

	"github.com/eryajf/go-ldap-admin/config"
	"github.com/eryajf/go-ldap-admin/middleware"
	"github.com/eryajf/go-ldap-admin/public/common"
	"github.com/eryajf/go-ldap-admin/routes"
	"github.com/eryajf/go-ldap-admin/service/isql"
)

// @title Go Ldap Admin
// @version 1.0
// @description 基于Go+Vue实现的openLDAP后台管理项目
// @termsOfService https://github.com/eryajf/go-ldap-admin

// @contact.name 项目作者：二丫讲梵 、 swagger作者：南宫乘风
// @contact.url https://github.com/eryajf/go-ldap-admin
// @contact.email https://github.com/eryajf/go-ldap-admin

// @host 127.0.0.1:8888
// @BasePath /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {

	// 加载配置文件到全局配置结构体
	config.InitConfig()

	// 初始化日志
	common.InitLogger()

	// 初始化数据库(mysql)
	common.InitDB()

	// 初始化ldap连接
	common.InitLDAP()

	// 初始化casbin策略管理器
	common.InitCasbinEnforcer()

	// 初始化Validator数据校验
	common.InitValidate()

	// 初始化mysql数据
	common.InitData()

	// 操作日志中间件处理日志时没有将日志发送到rabbitmq或者kafka中, 而是发送到了channel中
	// 这里开启3个goroutine处理channel将日志记录到数据库
	for i := 0; i < 3; i++ {
		go isql.OperationLog.SaveOperationLogChannel(middleware.OperationLogChan)
	}

	// 注册所有路由
	r := routes.InitRoutes()

	host := "0.0.0.0"
	port := config.Conf.System.Port

	srv := &http.Server{
		Handler: r,
	}

	listenType := config.Conf.System.ListenType

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		switch listenType {
		case "socket":
			socketPath := config.Conf.System.Socket
			if socketPath == "" {
				common.Log.Fatal("listen-type is socket but socket path is empty")
			}
			// 确保 socket 文件所在目录存在
			socketDir := filepath.Dir(socketPath)
			if err := os.MkdirAll(socketDir, 0755); err != nil {
				common.Log.Fatalf("Failed to create socket directory %s: %s", socketDir, err)
			}
			// 删除可能残留的旧 socket 文件
			if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
				common.Log.Fatalf("Failed to remove old socket file %s: %s", socketPath, err)
			}
			listener, err := net.Listen("unix", socketPath)
			if err != nil {
				common.Log.Fatalf("Failed to listen on unix socket %s: %s", socketPath, err)
			}
			// 设置 socket 文件权限，方便 nginx 等进程访问
			if err := os.Chmod(socketPath, 0666); err != nil {
				common.Log.Fatalf("Failed to chmod socket file %s: %s", socketPath, err)
			}
			common.Log.Info(fmt.Sprintf("Server is running at unix://%s", socketPath))
			if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
				common.Log.Fatalf("listen: %s\n", err)
			}
		default:
			// 默认使用 TCP 监听
			srv.Addr = fmt.Sprintf("%s:%d", host, port)
			common.Log.Info(fmt.Sprintf("Server is running at http://%s:%d", host, port))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				common.Log.Fatalf("listen: %s\n", err)
			}
		}
	}()

	// 启动定时任务
	logic.InitCron()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(quit, os.Interrupt)
	<-quit
	common.Log.Info("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		common.Log.Fatal("Server forced to shutdown:", err)
	}

	common.Log.Info("Server exiting!")

}
