package handlers

import (
	"github.com/bonheur15/go-db-manager/database"
	"github.com/gin-gonic/gin"
)

func CreateMongoHandler(mongoURI string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.MongoCreateDatabase(c, mongoURI)
	}
}

func MongoResetCredentialsHandler(mongoURI string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.MongoResetCredentials(c, mongoURI)
	}
}

func MongoRenameDatabaseHandler(mongoURI string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.MongoRenameDatabase(c, mongoURI)
	}
}

func MongoDeleteDatabaseHandler(mongoURI string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.MongoDeleteDatabase(c, mongoURI)
	}
}

func MongoViewDatabaseStatsHandler(mongoURI string) gin.HandlerFunc {
	return func(c *gin.Context) {
		database.MongoViewDatabaseStats(c, mongoURI)
	}
}
