package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bonheur15/go-db-manager/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func init() {
	validate = validator.New()
}

func ConnectToMySQLDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort string) (*sql.DB, error) {
	return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort))
}

func MysqlCreateDatabase(c *gin.Context, mysqlDbHost, mysqlDbUser, mysqlDbPassword, mysqlDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-validation")
		return
	}

	db, err := ConnectToMySQLDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-connection-open")
		return
	}
	defer db.Close()

	if _, err := db.Exec("CREATE DATABASE `" + requestBody.DatabaseName + "`"); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-create-database")
		return
	}

	username, err := utils.RandomString(12)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-create-user-random-string")
		return
	}
	password, err := utils.RandomString(16)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-create-user-random-string")
		return
	}
	if _, err := db.Exec("CREATE USER ?@'%' IDENTIFIED BY ?", username, password); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-create-user")
		return
	}

	if _, err := db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO ?@'%%'", requestBody.DatabaseName), username); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-grant-privileges-user")
		return
	}

	if _, err := db.Exec("FLUSH PRIVILEGES"); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-flush-privileges-user")
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "mysql-create-database", "Database Created")
}

func MysqlResetCredentials(c *gin.Context, mysqlDbHost, mysqlDbUser, mysqlDbPassword, mysqlDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-validation")
		return
	}

	db, err := ConnectToMySQLDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-connection-open")
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT User FROM mysql.db WHERE Db = ?", requestBody.DatabaseName)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-get-existing-users")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			utils.ErrorResponse(c, err, startTime, "mysql-scan-user")
			return
		}
		if _, err := db.Exec("DROP USER ?@'%'", username); err != nil {
			utils.ErrorResponse(c, err, startTime, "mysql-drop-user")
		}
	}

	username, err := utils.RandomString(12)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-create-user-random-string")
		return
	}
	password, err := utils.RandomString(16)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-create-user-random-string")
		return
	}
	if _, err := db.Exec("CREATE USER ?@'%' IDENTIFIED BY ?", username, password); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-create-user")
		return
	}

	if _, err := db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON `%s`.* TO ?@'%%'", requestBody.DatabaseName), username); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-grant-privileges-user")
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "mysql-reset-credentials", "Database Credentials Reset")
}

func MysqlRenameDatabase(c *gin.Context, mysqlDbHost, mysqlDbUser, mysqlDbPassword, mysqlDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		OldDatabaseName string `json:"old_database_name" validate:"required,alphanum"`
		NewDatabaseName string `json:"new_database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-validation")
		return
	}

	db, err := ConnectToMySQLDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-connection-open")
		return
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", requestBody.NewDatabaseName)); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-create-new-database")
		return
	}

	rows, err := db.Query(fmt.Sprintf("SHOW TABLES FROM `%s`", requestBody.OldDatabaseName))
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-show-tables")
		return
	}
	defer rows.Close()

	var tableName string
	for rows.Next() {
		err := rows.Scan(&tableName)
		if err != nil {
			utils.ErrorResponse(c, err, startTime, "mysql-scan-table-name")
			return
		}
		if _, err := db.Exec(fmt.Sprintf("RENAME TABLE `%s`.`%s` TO `%s`.`%s`",
			requestBody.OldDatabaseName, tableName, requestBody.NewDatabaseName, tableName)); err != nil {
			utils.ErrorResponse(c, err, startTime, "mysql-rename-table")
			return
		}
	}

	if _, err := db.Exec(fmt.Sprintf("DROP DATABASE `%s`", requestBody.OldDatabaseName)); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-drop-old-database")
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"new_database_name": requestBody.NewDatabaseName,
	}, startTime, "mysql-rename-database", "Database Renamed")
}

func MysqlDeleteDatabase(c *gin.Context, mysqlDbHost, mysqlDbUser, mysqlDbPassword, mysqlDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-validation")
		return
	}

	db, err := ConnectToMySQLDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-connection-open")
		return
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("DROP DATABASE `%s`", requestBody.DatabaseName)); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-drop-database")
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
	}, startTime, "mysql-delete-database", "Database Deleted")
}

func MysqlViewDatabaseStats(c *gin.Context, mysqlDbHost, mysqlDbUser, mysqlDbPassword, mysqlDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-validation")
		return
	}

	db, err := ConnectToMySQLDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-connection-open")
		return
	}
	defer db.Close()

	query := `
		SELECT table_name AS "Table",
			round(((data_length + index_length) / 1024 / 1024), 2) AS "Size (MB)"
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		ORDER BY (data_length + index_length) DESC;
	`
	rows, err := db.Query(query, requestBody.DatabaseName)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mysql-query-database-stats")
		return
	}
	defer rows.Close()

	type TableStat struct {
		Table  string  `json:"table"`
		SizeMB float64 `json:"size_mb"`
	}
	var stats []TableStat
	for rows.Next() {
		var stat TableStat
		if err := rows.Scan(&stat.Table, &stat.SizeMB); err != nil {
			utils.ErrorResponse(c, err, startTime, "mysql-scan-database-stat")
			return
		}
		stats = append(stats, stat)
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
		"stats":         stats,
	}, startTime, "mysql-view-database-stats", "Database Statistics Retrieved")
}
