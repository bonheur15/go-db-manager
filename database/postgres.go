package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bonheur15/go-db-manager/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lib/pq"
)


func init() {
	validate = validator.New()
}

func ConnectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort, sslMode string) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=%s", postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort, sslMode)
	return sql.Open("postgres", connStr)
}

func PostgresCreateDatabase(c *gin.Context, postgresDbHost, postgresDbUser, postgresDbPassword, postgresDbPort, sslMode string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-validation")
		return
	}
	
	db, err := ConnectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort,sslMode)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-connection-open")
		return
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", pq.QuoteIdentifier(requestBody.DatabaseName))); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-create-database")
		return
	}

	username, err := utils.RandomString(12)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-create-user-random-string")
		return
	}
	password, err := utils.RandomString(16)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-create-user-random-string")
		return
	}
	createUserQuery := fmt.Sprintf("CREATE USER %s WITH PASSWORD %s", pq.QuoteIdentifier(username), pq.QuoteLiteral(password))
	_, err = db.Exec(createUserQuery)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-create-user")
		return
	}

	grantQuery := fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", pq.QuoteIdentifier(requestBody.DatabaseName), pq.QuoteIdentifier(username))
	if _, err := db.Exec(grantQuery); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-grant-privileges-user")
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "postgres-create-database", "Database Created")
}

func PostgresResetCredentials(c *gin.Context, postgresDbHost, postgresDbUser, postgresDbPassword, postgresDbPort,sslMode string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-validation")
		return
	}

	db, err := ConnectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort,sslMode)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-connection-open")
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT usename FROM pg_user WHERE usesysid IN (SELECT usesysid FROM pg_database WHERE datname = $1)", requestBody.DatabaseName)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-get-existing-users")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			utils.ErrorResponse(c, err, startTime, "postgres-scan-user")
			return
		}
		if _, err := db.Exec(fmt.Sprintf("DROP USER %s", pq.QuoteIdentifier(username))); err != nil {
			utils.ErrorResponse(c, err, startTime, "postgres-drop-user")
			// Continue to try and drop other users
		}
	}

	username, err := utils.RandomString(12)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-create-user-random-string")
		return
	}
	password, err := utils.RandomString(16)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-create-user-random-string")
		return
	}
	if _, err := db.Exec(fmt.Sprintf("CREATE USER %s WITH PASSWORD %s", pq.QuoteIdentifier(username), pq.QuoteLiteral(password))); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-create-user")
		return
	}

	if _, err := db.Exec(fmt.Sprintf("GRANT ALL PRIVILEGES ON DATABASE %s TO %s", pq.QuoteIdentifier(requestBody.DatabaseName), pq.QuoteIdentifier(username))); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-grant-privileges-user")
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "postgres-reset-credentials", "Database Credentials Reset")
}

func PostgresRenameDatabase(c *gin.Context, postgresDbHost, postgresDbUser, postgresDbPassword, postgresDbPort,sslMode string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		OldDatabaseName string `json:"old_database_name" validate:"required,alphanum"`
		NewDatabaseName string `json:"new_database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-validation")
		return
	}

	db, err := ConnectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort,sslMode)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-connection-open")
		return
	}
	defer db.Close()

	query := fmt.Sprintf("ALTER DATABASE %s RENAME TO %s", pq.QuoteIdentifier(requestBody.OldDatabaseName), pq.QuoteIdentifier(requestBody.NewDatabaseName))
	if _, err := db.Exec(query); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-rename-database")
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"new_database_name": requestBody.NewDatabaseName,
	}, startTime, "postgres-rename-database", "Database Renamed")
}

// Terminate connections to the specified database
func PostgresTerminateConnections(db *sql.DB, dbName string) error {
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
		_, err := db.Exec("SELECT pg_terminate_backend($1)", pid)
		if err != nil {
			// Log error but continue trying to terminate other connections
		}
	}

	return nil
}

// Handle deleting a PostgreSQL database and associated users
func PostgresDeleteDatabase(c *gin.Context, postgresDbHost, postgresDbUser, postgresDbPassword, postgresDbPort,sslMode string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-validation")
		return
	}

	// Connect to the maintenance database (e.g., postgres)
	adminDb, err := ConnectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort,sslMode)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-connection-open")
		return
	}
	defer adminDb.Close()

	// Terminate connections to the database associated with the user
	if err := PostgresTerminateConnections(adminDb, requestBody.DatabaseName); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-terminate-connections")
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
		utils.ErrorResponse(c, err, startTime, "postgres-get-existing-users")
		return
	}
	defer rows.Close()

	var userNames []string
	for rows.Next() {
		var userName string
		if err := rows.Scan(&userName); err != nil {
			utils.ErrorResponse(c, err, startTime, "postgres-scan-user-name")
			return
		}
		// if username is postgres skip it
		if userName == "postgres" {
			continue
		}

		userNames = append(userNames, userName)
	}

	// Drop the database
	if _, err := adminDb.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", pq.QuoteIdentifier(requestBody.DatabaseName))); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-drop-database")
		return
	}

	// Drop each user associated with the database
	for _, userName := range userNames {
		if _, err := adminDb.Exec(fmt.Sprintf("DROP ROLE IF EXISTS %s", pq.QuoteIdentifier(userName))); err != nil {
			utils.ErrorResponse(c, err, startTime, "postgres-drop-user")
			// Continue to try and drop other users
		}
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
	}, startTime, "postgres-delete-database", "Database Deleted")
}

func PostgresViewDatabaseStats(c *gin.Context, postgresDbHost, postgresDbUser, postgresDbPassword, postgresDbPort,sslMode string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-validation")
		return
	}

	db, err := ConnectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort,sslMode)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-connection-open")
		return
	}
	defer db.Close()

	query := `
		SELECT table_name AS "Table",
			round(pg_total_relation_size(format('%%I.%%I', current_database(), table_name)) / 1024 / 1024, 2) AS "Size (MB)"
		FROM information_schema.tables
		WHERE table_schema = 'public'
		ORDER BY pg_total_relation_size(format('%%I.%%I', current_database(), table_name)) DESC;
	`
	rows, err := db.Query(query)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-query-database-stats")
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
			utils.ErrorResponse(c, err, startTime, "postgres-scan-database-stat")
			return
		}
		stats = append(stats, stat)
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
		"stats":         stats,
	}, startTime, "postgres-view-database-stats", "Database Statistics Retrieved")
}

func PostgresGetTotalQueries(c *gin.Context, postgresDbHost, postgresDbUser, postgresDbPassword, postgresDbPort, sslMode string) {
	// Connect to the default database (usually "postgres")
	startTime := time.Now().UnixMilli()
	db, err := ConnectToPostgresDB(postgresDbUser, postgresDbPassword, postgresDbHost, postgresDbPort,sslMode)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "postgres-connection-open")
		return
	}
	defer db.Close()

	// Query to get user activity from pg_stat_statements
	query := `
	SELECT
		 pg_roles.rolname AS username,
		 pg_database.datname,
		 sum(pg_stat_statements.calls) AS total_queries
	FROM
		pg_stat_statements
	JOIN
		pg_database ON pg_stat_statements.dbid = pg_database.oid
	JOIN
		pg_roles ON pg_stat_statements.userid = pg_roles.oid
	WHERE
		 pg_roles.rolname != 'postgres'
	GROUP BY
		 pg_roles.rolname,
		 pg_database.datname
	`
	rows, err := db.Query(query)
	if err != nil {

		utils.ErrorResponse(c, err, startTime, "postgres-get-user-activity")
		return

	}

	defer rows.Close()
	// response that contain all rows
	var userActivities []struct {
		Username     string `json:"username"`
		DatabaseName string `json:"database_name"`
		TotalQueries int    `json:"total_queries"`
	}
	for rows.Next() {
		var userActivity struct {
			Username     string `json:"username"`
			DatabaseName string `json:"database_name"`
			TotalQueries int    `json:"total_queries"`
		}
		if err := rows.Scan(&userActivity.Username, &userActivity.DatabaseName, &userActivity.TotalQueries); err != nil {
			utils.ErrorResponse(c, err, startTime, "postgres-scan-user-activity")
			return
		}
		userActivities = append(userActivities, userActivity)
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"user_activities": userActivities,
	}, startTime, "postgres-get-total-queries", "Database Query retrieved successfully")

}
