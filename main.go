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

	// MySQL Configuration
	mysqlHost := getEnv("mysql_host", "localhost")
	mysqlUser := getEnv("mysql_user", "root")
	mysqlPassword := getEnv("mysql_password", "")
	mysqlPort := getEnv("mysql_port", "3306")
	fmt.Printf("MySQL Host: %s\n", mysqlHost)

	// MongoDB Configuration
	mongoURI := getEnv("mongo_uri", "mongodb://admin:password@localhost:27017")
	fmt.Printf("MongoDB URI: %s\n", mongoURI)

	// PostgreSQL Configuration
	postgresHost := getEnv("postgres_host", "localhost")
	postgresUser := getEnv("postgres_user", "postgres")
	postgresPassword := getEnv("postgres_password", "password")
	postgresPort := getEnv("postgres_port", "5432")
	fmt.Printf("PostgreSQL Host: %s\n", postgresHost)

	routes := gin.Default()
	routes.GET("/server-info", getServerInfoHandler)

	// MySQL Endpoints
	routes.POST("/mysql/create-database", createMySQLHandler(mysqlCreateDatabase, mysqlHost, mysqlUser, mysqlPassword, mysqlPort))
	routes.POST("/mysql/reset-credentials", createMySQLHandler(mysqlResetCredentials, mysqlHost, mysqlUser, mysqlPassword, mysqlPort))
	routes.POST("/mysql/rename-database", createMySQLHandler(mysqlRenameDatabase, mysqlHost, mysqlUser, mysqlPassword, mysqlPort))
	routes.POST("/mysql/delete-database", createMySQLHandler(mysqlDeleteDatabase, mysqlHost, mysqlUser, mysqlPassword, mysqlPort))
	routes.POST("/mysql/view-database-stats", createMySQLHandler(mysqlViewDatabaseStats, mysqlHost, mysqlUser, mysqlPassword, mysqlPort))

	// MongoDB Endpoints
	routes.POST("/mongo/create-database", createMongoHandler(mongoCreateDatabase, mongoURI))
	routes.POST("/mongo/reset-credentials", createMongoHandler(mongoResetCredentials, mongoURI))
	routes.POST("/mongo/rename-database", createMongoHandler(mongoRenameDatabase, mongoURI))
	routes.POST("/mongo/delete-database", createMongoHandler(mongoDeleteDatabase, mongoURI))
	routes.POST("/mongo/view-database-stats", createMongoHandler(mongoViewDatabaseStats, mongoURI))

	// PostgreSQL Endpoints
	routes.POST("/postgres/create-database", createPostgresHandler(postgresCreateDatabase, postgresHost, postgresUser, postgresPassword, postgresPort))
	routes.POST("/postgres/reset-credentials", createPostgresHandler(postgresResetCredentials, postgresHost, postgresUser, postgresPassword, postgresPort))
	routes.POST("/postgres/rename-database", createPostgresHandler(postgresRenameDatabase, postgresHost, postgresUser, postgresPassword, postgresPort))
	routes.POST("/postgres/delete-database", createPostgresHandler(postgresDeleteDatabase, postgresHost, postgresUser, postgresPassword, postgresPort))
	routes.POST("/postgres/view-database-stats", createPostgresHandler(postgresViewDatabaseStats, postgresHost, postgresUser, postgresPassword, postgresPort))

	routes.Run(":8080")
}

func createMySQLHandler(handlerFunc func(*gin.Context, string, string, string, string), host, user, password, port string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handlerFunc(ctx, host, user, password, port)
	}
}

func createMongoHandler(handlerFunc func(*gin.Context, string), uri string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handlerFunc(ctx, uri)
	}
}

func createPostgresHandler(handlerFunc func(*gin.Context, string, string, string, string), host, user, password, port string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handlerFunc(ctx, host, user, password, port)
	}
}
