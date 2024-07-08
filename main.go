package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gofor-little/env"
)

func loadEnv() {
	if err := env.Load(".env"); err != nil {
		panic(err)
	}
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func main() {
	fmt.Println("Started Program")
	loadEnv()

	mysqlHost := getEnv("mysql_host", "localhost")
	mysqlUser := getEnv("mysql_user", "root")
	mysqlPassword := getEnv("mysql_password", "")
	mysqlPort := getEnv("mysql_port", "3306")
	fmt.Printf("MySQL Host: %s\n", mysqlHost)

	routes := gin.Default()
	routes.GET("/server-info", getServerInfoHandler)

	routes.POST("/mysql/create-database", createMySQLHandler(mysqlCreateDatabase, mysqlHost, mysqlUser, mysqlPassword, mysqlPort))
	routes.POST("/mysql/reset-credentials", createMySQLHandler(mysqlResetCredentials, mysqlHost, mysqlUser, mysqlPassword, mysqlPort))
	routes.POST("/mysql/rename-database", createMySQLHandler(mysqlRenameDatabase, mysqlHost, mysqlUser, mysqlPassword, mysqlPort))
	routes.POST("/mysql/delete-database", createMySQLHandler(mysqlDeleteDatabase, mysqlHost, mysqlUser, mysqlPassword, mysqlPort))

	routes.Run(":8080")
}

func createMySQLHandler(handlerFunc func(*gin.Context, string, string, string, string), host, user, password, port string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handlerFunc(ctx, host, user, password, port)
	}
}
