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
// @description OpenLDAP management platform built with Go and Vue
// @termsOfService https://github.com/eryajf/go-ldap-admin

// @contact.name Authors: eryajf, nangongchengfeng
// @contact.url https://github.com/eryajf/go-ldap-admin
// @contact.email https://github.com/eryajf/go-ldap-admin

// @host 127.0.0.1:8888
// @BasePath /api
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {

	// Load configuration file into global config struct
	config.InitConfig()

	// Initialize logger
	common.InitLogger()

	// Initialize database (MySQL)
	common.InitDB()

	// Initialize LDAP connection
	common.InitLDAP()

	// Initialize Casbin policy enforcer
	common.InitCasbinEnforcer()

	// Initialize data validator
	common.InitValidate()

	// Initialize MySQL seed data
	common.InitData()

	// The operation-log middleware sends logs to a channel instead of RabbitMQ/Kafka.
	// Start 3 goroutines to drain the channel and persist logs to the database.
	for i := 0; i < 3; i++ {
		go isql.OperationLog.SaveOperationLogChannel(middleware.OperationLogChan)
	}

	// Register all routes
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
			// Ensure the socket file directory exists
			socketDir := filepath.Dir(socketPath)
			if err := os.MkdirAll(socketDir, 0755); err != nil {
				common.Log.Fatalf("Failed to create socket directory %s: %s", socketDir, err)
			}
			// Remove any leftover old socket file
			if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
				common.Log.Fatalf("Failed to remove old socket file %s: %s", socketPath, err)
			}
			listener, err := net.Listen("unix", socketPath)
			if err != nil {
				common.Log.Fatalf("Failed to listen on unix socket %s: %s", socketPath, err)
			}
			// Set socket file permissions so nginx and other processes can access it
			if err := os.Chmod(socketPath, 0666); err != nil {
				common.Log.Fatalf("Failed to chmod socket file %s: %s", socketPath, err)
			}
			common.Log.Info(fmt.Sprintf("Server is running at unix://%s", socketPath))
			if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
				common.Log.Fatalf("listen: %s\n", err)
			}
		default:
			// Default to TCP listener
			srv.Addr = fmt.Sprintf("%s:%d", host, port)
			common.Log.Info(fmt.Sprintf("Server is running at http://%s:%d", host, port))
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				common.Log.Fatalf("listen: %s\n", err)
			}
		}
	}()

	// Start scheduled tasks
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
