package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bonheur15/go-db-manager/handlers"
	"github.com/bonheur15/go-db-manager/utils"
	"github.com/gin-gonic/gin"
	"github.com/gofor-little/env"
	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)



type Config struct {
	MongoURI         string
	MySQLDbHost      string
	MySQLDbUser      string
	MySQLDbPassword  string
	MySQLDbPort      string
	PostgresDbHost   string
	PostgresDbUser   string
	PostgresDbPassword string
	PostgresDbPort   string
	APIKey           string
	Sslmode          string
}



func LoadConfig() (*Config, error) {
	if err := env.Load(".env"); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	config := &Config{
		MongoURI:         os.Getenv("MONGO_URI"),
		MySQLDbHost:      os.Getenv("MYSQL_DB_HOST"),
		MySQLDbUser:      os.Getenv("MYSQL_DB_USER"),
		MySQLDbPassword:  os.Getenv("MYSQL_DB_PASSWORD"),
		MySQLDbPort:      os.Getenv("MYSQL_DB_PORT"),
		PostgresDbHost:   os.Getenv("POSTGRES_DB_HOST"),
		PostgresDbUser:   os.Getenv("POSTGRES_DB_USER"),
		PostgresDbPassword: os.Getenv("POSTGRES_DB_PASSWORD"),
		PostgresDbPort:   os.Getenv("POSTGRES_DB_PORT"),
		APIKey:           os.Getenv("API_KEY"),
		Sslmode:          os.Getenv("SSL_MODE"),
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API_KEY environment variable not set")
	}

	return config, nil
}



func AuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-API-KEY") != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

func main() {
	utils.InitLogger()
	log.Info().Msg("Started Program")

	config, err := LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	rateLimiter := handlers.NewIPRateLimiter(rate.Limit(10), 20)

	routes := gin.Default()
	routes.Use(rateLimiter.RateLimit())
	routes.Use(AuthMiddleware(config.APIKey))

	routes.GET("/server-info", handlers.GetServerInfoHandler)

	
	mysqlRoutes := routes.Group("/mysql")
	{
		mysqlRoutes.POST("/databases", handlers.CreateMySQLHandler(config.MySQLDbHost, config.MySQLDbUser, config.MySQLDbPassword, config.MySQLDbPort))
		mysqlRoutes.PATCH("/databases/:dbName/credentials", handlers.MySQLResetCredentialsHandler(config.MySQLDbHost, config.MySQLDbUser, config.MySQLDbPassword, config.MySQLDbPort))
		mysqlRoutes.PATCH("/databases/:dbName", handlers.MySQLRenameDatabaseHandler(config.MySQLDbHost, config.MySQLDbUser, config.MySQLDbPassword, config.MySQLDbPort))
		mysqlRoutes.DELETE("/databases/:dbName", handlers.MySQLDeleteDatabaseHandler(config.MySQLDbHost, config.MySQLDbUser, config.MySQLDbPassword, config.MySQLDbPort))
		mysqlRoutes.GET("/databases/:dbName/stats", handlers.MySQLViewDatabaseStatsHandler(config.MySQLDbHost, config.MySQLDbUser, config.MySQLDbPassword, config.MySQLDbPort))
	}

	
	mongoRoutes := routes.Group("/mongo")
	{
		mongoRoutes.POST("/databases", handlers.CreateMongoHandler(config.MongoURI))
		mongoRoutes.PATCH("/databases/:dbName/credentials", handlers.MongoResetCredentialsHandler(config.MongoURI))
		mongoRoutes.PATCH("/databases/:dbName", handlers.MongoRenameDatabaseHandler(config.MongoURI))
		mongoRoutes.DELETE("/databases/:dbName", handlers.MongoDeleteDatabaseHandler(config.MongoURI))
		mongoRoutes.GET("/databases/:dbName/stats", handlers.MongoViewDatabaseStatsHandler(config.MongoURI))
	}

	
	postgresRoutes := routes.Group("/postgres")
	{
		postgresRoutes.POST("/databases", handlers.CreatePostgresHandler(config.PostgresDbHost, config.PostgresDbUser, config.PostgresDbPassword, config.PostgresDbPort,config.Sslmode))
		postgresRoutes.PATCH("/databases/:dbName/credentials", handlers.PostgresResetCredentialsHandler(config.PostgresDbHost, config.PostgresDbUser, config.PostgresDbPassword, config.PostgresDbPort,config.Sslmode))
		postgresRoutes.PATCH("/databases/:dbName", handlers.PostgresRenameDatabaseHandler(config.PostgresDbHost, config.PostgresDbUser, config.PostgresDbPassword, config.PostgresDbPort,config.Sslmode))
		postgresRoutes.DELETE("/databases/:dbName", handlers.PostgresDeleteDatabaseHandler(config.PostgresDbHost, config.PostgresDbUser, config.PostgresDbPassword, config.PostgresDbPort,config.Sslmode))
		postgresRoutes.GET("/databases/:dbName/stats", handlers.PostgresViewDatabaseStatsHandler(config.PostgresDbHost, config.PostgresDbUser, config.PostgresDbPassword, config.PostgresDbPort,config.Sslmode))
		postgresRoutes.GET("/databases/queries", handlers.PostgresGetTotalQueriesHandler(config.PostgresDbHost, config.PostgresDbUser, config.PostgresDbPassword, config.PostgresDbPort, config.Sslmode))
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: routes,
	}

	go func() {
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("listen: %s\n")
		}
	}()

	
	quit := make(chan os.Signal, 1)
	
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown:")
	}

	log.Info().Msg("Server exiting")
}


