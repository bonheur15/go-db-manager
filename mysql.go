package main

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

const (
	mysqlDbUser     = "root"
	mysqlDbPassword = ""
	mysqlDbHost     = "localhost"
	mysqlDbPort     = "3306"
)

func randomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

func mysqlCreateDatabase(c *gin.Context) {
	startTime := time.Now().UnixMilli()
	type RequesBody struct {
		DatabaseName string `json:"database_name"`
	}
	var requestBody RequesBody
	err := c.BindJSON(&requestBody)
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort))
	if err != nil {
		c.JSON(500, gin.H{
			"error":           true,
			"message":         err.Error(),
			"action":          "mysql-connection-open",
			"timestamp":       time.Now(),
			"action_duration": time.Now().UnixMilli() - startTime,
			"data":            nil,
		})
		return
	}
	defer db.Close()
	_, err = db.Exec("CREATE DATABASE " + requestBody.DatabaseName)
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-create-database")
		return
	}
	username, _ := randomString(12)
	password, _ := randomString(16)
	_, err = db.Exec(fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s'", username, password))
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-create-user")
		return
	}
	_, err = db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%'", requestBody.DatabaseName, username))
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-grant-privileges-user")
		return
	}

	SuccessResponse(c, map[string]string{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "mysql-create-database", "Database Created")

}

func mysqlResetCredentials(c *gin.Context) {
	startTime := time.Now().UnixMilli()
	type RequesBody struct {
		DatabaseName string `json:"database_name"`
	}
	var requestBody RequesBody
	err := c.BindJSON(&requestBody)
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort))
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-connection-open")
		return
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT User FROM mysql.db WHERE Db = '%s'", requestBody.DatabaseName))
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-get-existing-users")

		return
	}
	defer rows.Close()
	for rows.Next() {
		var username string
		rows.Scan(&username)
		db.Exec(fmt.Sprintf("DROP USER '%s'@'%%'", username))
	}
	username, _ := randomString(12)
	password, _ := randomString(16)
	_, err = db.Exec(fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s'", username, password))
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-create-user")
		return
	}
	_, err = db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%'", requestBody.DatabaseName, username))

	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-grant-privileges-user")
		return
	}

	SuccessResponse(c, map[string]string{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "mysql-reset-credentials", "Database Credentials Has been reset")

}
func mysqlRenameDatabase(c *gin.Context) {
	startTime := time.Now().UnixMilli()
	type RequesBody struct {
		OldDatabaseName string `json:"old_database_name"`
		NewDatabaseName string `json:"new_database_name"`
	}
	var requestBody RequesBody
	err := c.BindJSON(&requestBody)
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort))
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-connection-open")
		return
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("RENAME TABLE %s TO %s", requestBody.OldDatabaseName, requestBody.NewDatabaseName))
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-rename-table")
		return
	}

	SuccessResponse(c, map[string]string{
		"old_database_name": requestBody.OldDatabaseName,
		"new_database_name": requestBody.NewDatabaseName,
	}, startTime, "mysql-rename-database", "Database Renamed")
}

func SuccessResponse(c *gin.Context, data map[string]string, startTime int64, action string, message string) {
	c.JSON(200, gin.H{
		"data":            data,
		"error":           false,
		"action":          "mysql-rename-database",
		"message":         "Database Renamed",
		"timestamp":       time.Now(),
		"action_duration": time.Now().UnixMilli() - startTime,
	})
}

func ErrorResponse(c *gin.Context, error error, startTime int64, action string) {
	c.JSON(400, gin.H{
		"error":           true,
		"message":         error.Error(),
		"action":          action,
		"timestamp":       time.Now(),
		"action_duration": time.Now().UnixMilli() - startTime,
		"data":            nil,
	})
}
