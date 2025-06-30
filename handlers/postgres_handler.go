package handlers

import (
	"github.com/bonheur15/go-db-manager/database"
	"github.com/gin-gonic/gin"
)

func CreatePostgresHandler(host, user, password, port, sslMode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.PostgresCreateDatabase(c, host, user, password, port, sslMode)
	}
}

func PostgresResetCredentialsHandler(host, user, password, port, sslMode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.PostgresResetCredentials(c, host, user, password, port, sslMode)
	}
}

func PostgresRenameDatabaseHandler(host, user, password, port, sslMode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.PostgresRenameDatabase(c, host, user, password, port, sslMode)
	}
}

func PostgresDeleteDatabaseHandler(host, user, password, port, sslMode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.PostgresDeleteDatabase(c, host, user, password, port, sslMode)
	}
}

func PostgresViewDatabaseStatsHandler(host, user, password, port, sslMode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.PostgresViewDatabaseStats(c, host, user, password, port, sslMode)
	}
}

func PostgresGetTotalQueriesHandler(host, user, password, port, sslMode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.PostgresGetTotalQueries(c, host, user, password, port, sslMode)
	}
}
