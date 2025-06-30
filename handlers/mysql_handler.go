package handlers

import (
	"github.com/bonheur15/go-db-manager/database"
	"github.com/gin-gonic/gin"
)

func CreateMySQLHandler(host, user, password, port string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.MysqlCreateDatabase(c, host, user, password, port)
	}
}

func MySQLResetCredentialsHandler(host, user, password, port string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.MysqlResetCredentials(c, host, user, password, port)
	}
}

func MySQLRenameDatabaseHandler(host, user, password, port string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.MysqlRenameDatabase(c, host, user, password, port)
	}
}

func MySQLDeleteDatabaseHandler(host, user, password, port string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.MysqlDeleteDatabase(c, host, user, password, port)
	}
}

func MySQLViewDatabaseStatsHandler(host, user, password, port string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.MysqlViewDatabaseStats(c, host, user, password, port)
	}
}
