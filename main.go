package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Started Program")
	routes := gin.Default()
	routes.GET("/server-info", getServerInfoHandler)
	routes.POST("/mysql/create-database", mysqlCreateDatabase)
	routes.POST("/mysql/reset-credentials", mysqlResetCredentials)
	routes.POST("/mysql/rename-database", mysqlRenameDatabase)
	routes.POST("/mysql/delete-database", mysqlDeleteDatabase)
	routes.Run(":8080")
}
