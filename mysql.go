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

func mysqlCreateDatabase(c *gin.Context, mysqlDbHost string, mysqlDbUser string, mysqlDbPassword string, mysqlDbPort string) {
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
	_, err = db.Exec("FLUSH PRIVILEGES")
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-flush-privileges-user")
		return
	}

	SuccessResponse(c, map[string]string{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "mysql-create-database", "Database Created")

}

func mysqlResetCredentials(c *gin.Context, mysqlDbHost string, mysqlDbUser string, mysqlDbPassword string, mysqlDbPort string) {
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
func mysqlRenameDatabase(c *gin.Context, mysqlDbHost string, mysqlDbUser string, mysqlDbPassword string, mysqlDbPort string) {
	startTime := time.Now().UnixMilli()

	var requestBody struct {
		OldDatabaseName string `json:"old_database_name"`
		NewDatabaseName string `json:"new_database_name"`
	}
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
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", requestBody.NewDatabaseName))
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-create-new-database")
		return
	}
	rows, err := db.Query(fmt.Sprintf("SHOW TABLES FROM `%s`", requestBody.OldDatabaseName))
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-show-tables")
		return
	}
	defer rows.Close()

	var tableName string
	for rows.Next() {
		err = rows.Scan(&tableName)
		if err != nil {
			ErrorResponse(c, err, startTime, "mysql-scan-table-name")
			return
		}
		_, err = db.Exec(fmt.Sprintf("CREATE TABLE `%s`.`%s` LIKE `%s`.`%s`",
			requestBody.NewDatabaseName, tableName, requestBody.OldDatabaseName, tableName))
		if err != nil {
			ErrorResponse(c, err, startTime, "mysql-create-table")
			return
		}
		_, err = db.Exec(fmt.Sprintf("INSERT INTO `%s`.`%s` SELECT * FROM `%s`.`%s`",
			requestBody.NewDatabaseName, tableName, requestBody.OldDatabaseName, tableName))
		if err != nil {
			ErrorResponse(c, err, startTime, "mysql-copy-data")
			return
		}
	}
	privilegesQuery := fmt.Sprintf(`
		SELECT CONCAT('GRANT ', privilege_type, ' ON %s.* TO ', grantee, ';')
		FROM information_schema.schema_privileges
		WHERE table_schema = '%s';
	`, requestBody.NewDatabaseName, requestBody.OldDatabaseName)

	privilegesRows, err := db.Query(privilegesQuery)
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-fetch-privileges")
		return
	}
	defer privilegesRows.Close()

	var grantStatement string
	for privilegesRows.Next() {
		err = privilegesRows.Scan(&grantStatement)
		if err != nil {
			ErrorResponse(c, err, startTime, "mysql-scan-privilege")
			return
		}
		_, err = db.Exec(grantStatement)
		if err != nil {
			ErrorResponse(c, err, startTime, "mysql-apply-privilege")
			return
		}
	}

	revokePrivilegesQuery := fmt.Sprintf(`
		SELECT CONCAT('REVOKE ', privilege_type, ' ON %s.* FROM ', grantee, ';')
		FROM information_schema.schema_privileges
		WHERE table_schema = '%s';
	`, requestBody.OldDatabaseName, requestBody.OldDatabaseName)

	revokePrivilegesRows, err := db.Query(revokePrivilegesQuery)
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-fetch-revoke-privileges")
		return
	}
	defer revokePrivilegesRows.Close()

	var revokeStatement string
	for revokePrivilegesRows.Next() {
		err = revokePrivilegesRows.Scan(&revokeStatement)
		if err != nil {
			ErrorResponse(c, err, startTime, "mysql-scan-revoke-privilege")
			return
		}
		_, err = db.Exec(revokeStatement)
		if err != nil {
			ErrorResponse(c, err, startTime, "mysql-apply-revoke-privilege")
			return
		}
	}
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE `%s`", requestBody.OldDatabaseName))
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-drop-old-database")
		return
	}

	SuccessResponse(c, map[string]string{
		"old_database_name": requestBody.OldDatabaseName,
		"new_database_name": requestBody.NewDatabaseName,
	}, startTime, "mysql-rename-database", "Database Renamed")
}
func mysqlDeleteDatabase(c *gin.Context, mysqlDbHost string, mysqlDbUser string, mysqlDbPassword string, mysqlDbPort string) {
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
		ErrorResponse(c, err, startTime, "mysql-get-user")
		return
	}
	defer rows.Close()

	// Delete existing users
	for rows.Next() {
		var username string
		rows.Scan(&username)
		db.Exec(fmt.Sprintf("DROP USER '%s'@'%%'", username))
	}

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", requestBody.DatabaseName))
	if err != nil {

		ErrorResponse(c, err, startTime, "mysql-drop-database")
		return
	}
	SuccessResponse(c, map[string]string{
		"database_name": requestBody.DatabaseName,
	}, startTime, "mysql-drop-database", "Database has been successfully dropped")

}

func SuccessResponse(c *gin.Context, data map[string]string, startTime int64, action string, message string) {
	c.JSON(200, gin.H{
		"data":            data,
		"error":           false,
		"action":          action,
		"message":         message,
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
