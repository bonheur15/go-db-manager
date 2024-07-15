package main

import (
	"database/sql"
	"fmt"

	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func connectToDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort string) (*sql.DB, error) {
	return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort))
}

func mysqlCreateDatabase(c *gin.Context, mysqlDbHost, mysqlDbUser, mysqlDbPassword, mysqlDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	db, err := connectToDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort)
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-connection-open")
		return
	}
	defer db.Close()

	if _, err := db.Exec("CREATE DATABASE " + requestBody.DatabaseName); err != nil {
		ErrorResponse(c, err, startTime, "mysql-create-database")
		return
	}

	username, _ := randomString(12)
	password, _ := randomString(16)
	if _, err := db.Exec(fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s'", username, password)); err != nil {
		ErrorResponse(c, err, startTime, "mysql-create-user")
		return
	}

	if _, err := db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%'", requestBody.DatabaseName, username)); err != nil {
		ErrorResponse(c, err, startTime, "mysql-grant-privileges-user")
		return
	}

	if _, err := db.Exec("FLUSH PRIVILEGES"); err != nil {
		ErrorResponse(c, err, startTime, "mysql-flush-privileges-user")
		return
	}

	SuccessResponse(c, map[string]interface{}{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "mysql-create-database", "Database Created")
}

func mysqlResetCredentials(c *gin.Context, mysqlDbHost, mysqlDbUser, mysqlDbPassword, mysqlDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	db, err := connectToDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort)
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
	if _, err := db.Exec(fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s'", username, password)); err != nil {
		ErrorResponse(c, err, startTime, "mysql-create-user")
		return
	}

	if _, err := db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%'", requestBody.DatabaseName, username)); err != nil {
		ErrorResponse(c, err, startTime, "mysql-grant-privileges-user")
		return
	}

	SuccessResponse(c, map[string]interface{}{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "mysql-reset-credentials", "Database Credentials Reset")
}

func mysqlRenameDatabase(c *gin.Context, mysqlDbHost, mysqlDbUser, mysqlDbPassword, mysqlDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		OldDatabaseName string `json:"old_database_name"`
		NewDatabaseName string `json:"new_database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	db, err := connectToDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort)
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-connection-open")
		return
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", requestBody.NewDatabaseName)); err != nil {
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
		err := rows.Scan(&tableName)
		if err != nil {
			ErrorResponse(c, err, startTime, "mysql-scan-table-name")
			return
		}
		if _, err := db.Exec(fmt.Sprintf("CREATE TABLE `%s`.`%s` LIKE `%s`.`%s`",
			requestBody.NewDatabaseName, tableName, requestBody.OldDatabaseName, tableName)); err != nil {
			ErrorResponse(c, err, startTime, "mysql-create-table")
			return
		}
		if _, err := db.Exec(fmt.Sprintf("INSERT INTO `%s`.`%s` SELECT * FROM `%s`.`%s`",
			requestBody.NewDatabaseName, tableName, requestBody.OldDatabaseName, tableName)); err != nil {
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
		err := privilegesRows.Scan(&grantStatement)
		if err != nil {
			ErrorResponse(c, err, startTime, "mysql-scan-privilege")
			return
		}
		if _, err := db.Exec(grantStatement); err != nil {
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
		err := revokePrivilegesRows.Scan(&revokeStatement)
		if err != nil {
			ErrorResponse(c, err, startTime, "mysql-scan-revoke-privilege")
			return
		}
		if _, err := db.Exec(revokeStatement); err != nil {
			ErrorResponse(c, err, startTime, "mysql-revoke-privilege")
			return
		}
	}

	if _, err := db.Exec(fmt.Sprintf("DROP DATABASE `%s`", requestBody.OldDatabaseName)); err != nil {
		ErrorResponse(c, err, startTime, "mysql-drop-old-database")
		return
	}

	SuccessResponse(c, map[string]interface{}{
		"new_database_name": requestBody.NewDatabaseName,
	}, startTime, "mysql-rename-database", "Database Renamed")
}

func mysqlDeleteDatabase(c *gin.Context, mysqlDbHost, mysqlDbUser, mysqlDbPassword, mysqlDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	db, err := connectToDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort)
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-connection-open")
		return
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("DROP DATABASE `%s`", requestBody.DatabaseName)); err != nil {
		ErrorResponse(c, err, startTime, "mysql-drop-database")
		return
	}

	if _, err := db.Exec(fmt.Sprintf("DELETE FROM mysql.db WHERE Db='%s'", requestBody.DatabaseName)); err != nil {
		ErrorResponse(c, err, startTime, "mysql-delete-database-entry")
		return
	}

	SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
	}, startTime, "mysql-delete-database", "Database Deleted")
}

func mysqlViewDatabaseStats(c *gin.Context, mysqlDbHost, mysqlDbUser, mysqlDbPassword, mysqlDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "mysql-bind-json")
		return
	}

	db, err := connectToDB(mysqlDbUser, mysqlDbPassword, mysqlDbHost, mysqlDbPort)
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-connection-open")
		return
	}
	defer db.Close()

	query := fmt.Sprintf(`
		SELECT table_name AS "Table",
			round(((data_length + index_length) / 1024 / 1024), 2) AS "Size (MB)"
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = "%s"
		ORDER BY (data_length + index_length) DESC;
	`, requestBody.DatabaseName)
	rows, err := db.Query(query)
	if err != nil {
		ErrorResponse(c, err, startTime, "mysql-query-database-stats")
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
			ErrorResponse(c, err, startTime, "mysql-scan-database-stat")
			return
		}
		stats = append(stats, stat)
	}

	SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
		"stats":         stats,
	}, startTime, "mysql-view-database-stats", "Database Statistics Retrieved")
}
