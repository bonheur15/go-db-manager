package handlers

import (
	"time"

	"github.com/bonheur15/go-db-manager/utils"
	"github.com/gin-gonic/gin"
)

func GetServerInfoHandler(c *gin.Context) {
	startTime := time.Now().UnixMilli()
	serverInfo, err := utils.GetServerInfo()
	if err != nil {
		c.JSON(500, gin.H{
			"error":           true,
			"message":         err.Error(),
			"action":          "server-info",
			"timestamp":       time.Now(),
			"action_duration": time.Now().UnixMilli() - startTime,
			"data":            nil,
		})
		return
	}

	c.JSON(200, gin.H{
		"data":            serverInfo,
		"error":           false,
		"action":          "server-info",
		"message":         "Server Info",
		"timestamp":       time.Now(),
		"action_duration": time.Now().UnixMilli() - startTime,
	})
}
