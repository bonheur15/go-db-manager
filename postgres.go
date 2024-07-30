package main

import (
	"database/sql"
	"fmt"

	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func connectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort string) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=disable", postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort)
	return sql.Open("postgres", connStr)
}

func postgresCreateDatabase(c *gin.Context, postgresDbHost, postgresDbUser, postgresDbPassword, postgresDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "postgres-bind-json")
		return
	}

	db, err := connectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort)
	if err != nil {
		ErrorResponse(c, err, startTime, "postgres-connection-open")
		return
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", requestBody.DatabaseName)); err != nil {
		ErrorResponse(c, err, startTime, "postgres-create-database")
		return
	}

	username, _ := randomString(12)
	password, _ := randomString(16)
	_, err = db.Exec(fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", username, password))
	if err != nil {
		ErrorResponse(c, err, startTime, "postgres-create-user")
		return
	}

	if _, err := db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", requestBody.DatabaseName, username)); err != nil {
		ErrorResponse(c, err, startTime, "postgres-grant-privileges-user")
		return
	}

	SuccessResponse(c, map[string]interface{}{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "postgres-create-database", "Database Created")
}

func postgresResetCredentials(c *gin.Context, postgresDbHost, postgresDbUser, postgresDbPassword, postgresDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "postgres-bind-json")
		return
	}

	db, err := connectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort)
	if err != nil {
		ErrorResponse(c, err, startTime, "postgres-connection-open")
		return
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT usename FROM pg_user WHERE usesysid IN (SELECT usesysid FROM pg_database WHERE datname = '%s')", requestBody.DatabaseName))
	if err != nil {
		ErrorResponse(c, err, startTime, "postgres-get-existing-users")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		rows.Scan(&username)
		db.Exec(fmt.Sprintf("DROP USER %s", username))
	}

	username, _ := randomString(12)
	password, _ := randomString(16)
	if _, err := db.Exec(fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", username, password)); err != nil {
		ErrorResponse(c, err, startTime, "postgres-create-user")
		return
	}

	if _, err := db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", requestBody.DatabaseName, username)); err != nil {
		ErrorResponse(c, err, startTime, "postgres-grant-privileges-user")
		return
	}

	SuccessResponse(c, map[string]interface{}{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "postgres-reset-credentials", "Database Credentials Reset")
}

func postgresRenameDatabase(c *gin.Context, postgresDbHost, postgresDbUser, postgresDbPassword, postgresDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		OldDatabaseName string `json:"old_database_name"`
		NewDatabaseName string `json:"new_database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "postgres-bind-json")
		return
	}

	db, err := connectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort)
	if err != nil {
		ErrorResponse(c, err, startTime, "postgres-connection-open")
		return
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("ALTER DATABASE %s RENAME TO %s", requestBody.OldDatabaseName, requestBody.NewDatabaseName)); err != nil {
		ErrorResponse(c, err, startTime, "postgres-rename-database")
		return
	}

	SuccessResponse(c, map[string]interface{}{
		"new_database_name": requestBody.NewDatabaseName,
	}, startTime, "postgres-rename-database", "Database Renamed")
}

// Terminate connections to the specified database
func postgresterminateConnections(db *sql.DB, dbName string) error {
	// Query to find all active connections to the database
	query := `
		SELECT pid
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid();
	`
	rows, err := db.Query(query, dbName)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Terminate each connection
	var pid int
	for rows.Next() {
		if err := rows.Scan(&pid); err != nil {
			return err
		}
		_, err := db.Exec(fmt.Sprintf("SELECT pg_terminate_backend(%d)", pid))
		if err != nil {
			return err
		}
	}

	return nil
}

// Handle deleting a PostgreSQL database and associated users
func postgresDeleteDatabase(c *gin.Context, postgresDbHost, postgresDbUser, postgresDbPassword, postgresDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "postgres-bind-json")
		return
	}

	// Connect to the maintenance database (e.g., postgres)
	adminDb, err := connectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort)
	if err != nil {
		ErrorResponse(c, err, startTime, "postgres-connection-open")
		return
	}
	defer adminDb.Close()

	// Terminate connections to the database associated with the user
	if err := postgresterminateConnections(adminDb, requestBody.DatabaseName); err != nil {
		ErrorResponse(c, err, startTime, "postgres-terminate-connections")
		return
	}
	// Identify users with CONNECT privileges on the specified database
	query := `
		SELECT u.usename
		FROM pg_user u
		WHERE has_database_privilege(u.usename, $1, 'CONNECT');
	`
	rows, err := adminDb.Query(query, requestBody.DatabaseName)
	if err != nil {
		ErrorResponse(c, err, startTime, "postgres-get-existing-users")
		return
	}
	defer rows.Close()

	var userNames []string
	for rows.Next() {
		var userName string
		if err := rows.Scan(&userName); err != nil {
			ErrorResponse(c, err, startTime, "postgres-scan-user-name")
			return
		}
		// if username is postgres skip it
		if userName == "postgres" {
			continue
		}

		userNames = append(userNames, userName)
	}

	// Drop the database
	if _, err := adminDb.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", requestBody.DatabaseName)); err != nil {
		ErrorResponse(c, err, startTime, "postgres-drop-database")
		return
	}

	// Drop each user associated with the database
	for _, userName := range userNames {
		if _, err := adminDb.Exec(fmt.Sprintf("DROP ROLE IF EXISTS %s", userName)); err != nil {
			ErrorResponse(c, err, startTime, "postgres-drop-user")
			return
		}
	}

	SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
	}, startTime, "postgres-delete-database", "Database Deleted")
}

func postgresViewDatabaseStats(c *gin.Context, postgresDbHost, postgresDbUser, postgresDbPassword, postgresDbPort string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "postgres-bind-json")
		return
	}

	db, err := connectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort)
	if err != nil {
		ErrorResponse(c, err, startTime, "postgres-connection-open")
		return
	}
	defer db.Close()

	query := fmt.Sprintf(`
		SELECT table_name AS "Table",
			round(pg_total_relation_size(format('%%I.%%I', current_database(), table_name)) / 1024 / 1024, 2) AS "Size (MB)"
		FROM information_schema.tables
		WHERE table_schema = 'public'
		ORDER BY pg_total_relation_size(format('%%I.%%I', current_database(), table_name)) DESC;
	`)
	rows, err := db.Query(query)
	if err != nil {
		ErrorResponse(c, err, startTime, "postgres-query-database-stats")
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
			ErrorResponse(c, err, startTime, "postgres-scan-database-stat")
			return
		}
		stats = append(stats, stat)
	}

	SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
		"stats":         stats,
	}, startTime, "postgres-view-database-stats", "Database Statistics Retrieved")
}

// Helper functions (ErrorResponse, SuccessResponse, randomString) should be implemented here
