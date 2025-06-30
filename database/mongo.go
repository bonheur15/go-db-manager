package database

import (
	"context"
	"fmt"
	"time"

	"github.com/bonheur15/go-db-manager/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	validate = validator.New()
}

func ConnectToMongoDB(mongoURI string) (*mongo.Client, context.Context, error) {
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	return client, context.Background(), err
}

func MongoCreateDatabase(c *gin.Context, mongoURI string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-validation")
		return
	}

	client, ctx, err := ConnectToMongoDB(mongoURI)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-connection-open")
		return
	}
	defer client.Disconnect(ctx)

	
	existingDatabases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-list-databases")
		return
	}

	for _, dbName := range existingDatabases {
		if dbName == requestBody.DatabaseName {
			utils.ErrorResponse(c, fmt.Errorf("database %s already exists", requestBody.DatabaseName), startTime, "mongo-database-exists")
			return
		}
	}

	
	collection := client.Database(requestBody.DatabaseName).Collection("test")
	if _, err := collection.InsertOne(ctx, bson.M{"test": "data"}); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-create-database")
		return
	}

	username, err := utils.RandomString(12)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-create-user-random-string")
		return
	}
	password, err := utils.RandomString(16)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-create-user-random-string")
		return
	}

	db := client.Database(requestBody.DatabaseName)
	createUserCmd := bson.D{
		{"createUser", username},
		{"pwd", password},
		{"roles", bson.A{bson.D{{"role", "readWrite"}, {"db", requestBody.DatabaseName}}}},
	}
	if err := db.RunCommand(ctx, createUserCmd).Err(); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-create-user")
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "mongo-create-database", "Database Created")
}

func MongoCopyDatabase(oldName, newName string, client *mongo.Client) error {
	oldDB := client.Database(oldName)
	newDB := client.Database(newName)

	collections, err := oldDB.ListCollectionNames(context.Background(), bson.D{})
	if err != nil {
		return err
	}

	for _, collection := range collections {
		oldColl := oldDB.Collection(collection)
		newColl := newDB.Collection(collection)

		cursor, err := oldColl.Find(context.Background(), bson.D{})
		if err != nil {
			return err
		}

		var documents []interface{}
		if err = cursor.All(context.Background(), &documents); err != nil {
			return err
		}

		if len(documents) > 0 {
			_, err = newColl.InsertMany(context.Background(), documents)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func MongoRenameDatabase(c *gin.Context, mongoURI string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		OldDatabaseName string `json:"old_database_name" validate:"required,alphanum"`
		NewDatabaseName string `json:"new_database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-validation")
		return
	}

	client, ctx, err := ConnectToMongoDB(mongoURI)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-connection-open")
		return
	}
	defer client.Disconnect(ctx)
	existingDatabases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-list-databases")
		return
	}
	dbexist := false
	for _, dbName := range existingDatabases {
		if dbName == requestBody.OldDatabaseName {
			dbexist = true
			break
		}
	}
	if !dbexist {
		utils.ErrorResponse(c, fmt.Errorf("database %s does not exist", requestBody.OldDatabaseName), startTime, "mongo-database-not-exist")
		return
	}

	for _, dbName := range existingDatabases {
		if dbName == requestBody.NewDatabaseName {
			utils.ErrorResponse(c, fmt.Errorf("database %s already exists", requestBody.NewDatabaseName), startTime, "mongo-database-exists")
			return
		}
	}

	err = MongoCopyDatabase(requestBody.OldDatabaseName, requestBody.NewDatabaseName, client)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-copy-database")
		return
	}
	client.Database(requestBody.OldDatabaseName).Drop(context.Background())
	utils.SuccessResponse(c, map[string]interface{}{
		"old_database_name": requestBody.OldDatabaseName,
		"new_database_name": requestBody.NewDatabaseName,
	}, startTime, "mongo-rename-database", "Database Renamed")
}

func MongoDeleteDatabase(c *gin.Context, mongoURI string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-validation")
		return
	}

	client, ctx, err := ConnectToMongoDB(mongoURI)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-connection-open")
		return
	}
	defer client.Disconnect(ctx)

	existingDatabases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-list-databases")
		return
	}
	dbexist := false
	for _, dbName := range existingDatabases {
		if dbName == requestBody.DatabaseName {
			dbexist = true
			break
		}
	}
	if !dbexist {
		utils.ErrorResponse(c, fmt.Errorf("database %s does not exist", requestBody.DatabaseName), startTime, "mongo-database-not-exist")
		return
	}

	if err := client.Database(requestBody.DatabaseName).Drop(ctx); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-drop-database")
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
	}, startTime, "mongo-delete-database", "Database Deleted")
}

func MongoViewDatabaseStats(c *gin.Context, mongoURI string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-validation")
		return
	}

	client, ctx, err := ConnectToMongoDB(mongoURI)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-connection-open")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database(requestBody.DatabaseName)
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-list-collections")
		return
	}

	var stats []bson.M
	for _, collectionName := range collections {
		collectionStats := bson.M{}
		if err := db.RunCommand(ctx, bson.D{{"collStats", collectionName}}).Decode(&collectionStats); err != nil {
			utils.ErrorResponse(c, err, startTime, "mongo-collection-stats")
			return
		}
		stats = append(stats, collectionStats)
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
		"stats":         stats,
	}, startTime, "mongo-view-database-stats", "Database Statistics Retrieved")
}

func MongoResetCredentials(c *gin.Context, mongoURI string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name" validate:"required,alphanum"`
		Username     string `json:"username" validate:"required,alphanum"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-bind-json")
		return
	}

	if err := validate.Struct(requestBody); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-validation")
		return
	}

	client, ctx, err := ConnectToMongoDB(mongoURI)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-connection-open")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database(requestBody.DatabaseName)

	
	newPassword, err := utils.RandomString(16)
	if err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-create-user-random-string")
		return
	}

	
	updateCmd := bson.D{
		{"updateUser", requestBody.Username},
		{"pwd", newPassword},
	}
	if err := db.RunCommand(ctx, updateCmd).Err(); err != nil {
		utils.ErrorResponse(c, err, startTime, "mongo-reset-credentials")
		return
	}

	utils.SuccessResponse(c, map[string]interface{}{
		"username": requestBody.Username,
		"password": newPassword,
	}, startTime, "mongo-reset-credentials", "Credentials Reset")
}
