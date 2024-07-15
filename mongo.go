package main

import (
	"context"
	"fmt"

	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func connectToMongoDB(mongoURI string) (*mongo.Client, context.Context, error) {
	clientOptions := options.Client().ApplyURI("mongodb://myUserAdmin:password@localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	return client, context.TODO(), err
}

func mongoCreateDatabase(c *gin.Context, mongoURI string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "mongo-bind-json")
		return
	}

	client, ctx, err := connectToMongoDB(mongoURI)
	if err != nil {
		ErrorResponse(c, err, startTime, "mongo-connection-open")
		return
	}
	defer client.Disconnect(ctx)

	// Check if the database already exists
	existingDatabases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		ErrorResponse(c, err, startTime, "mongo-list-databases")
		return
	}

	for _, dbName := range existingDatabases {
		if dbName == requestBody.DatabaseName {
			ErrorResponse(c, fmt.Errorf("database %s already exists", requestBody.DatabaseName), startTime, "mongo-database-exists")
			return
		}
	}

	collection := client.Database(requestBody.DatabaseName).Collection("test")
	if _, err := collection.InsertOne(ctx, bson.M{"test": "data"}); err != nil {
		ErrorResponse(c, err, startTime, "mongo-create-database")
		return
	}

	username, _ := randomString(12)
	password, _ := randomString(16)

	db := client.Database("admin")
	if _, err := db.RunCommand(ctx, bson.D{
		{"createUser", username},
		{"pwd", password},
		{"roles", bson.A{"readWrite"}},
	}).DecodeBytes(); err != nil {
		ErrorResponse(c, err, startTime, "mongo-create-user")
		return
	}

	SuccessResponse(c, map[string]interface{}{
		"username":      username,
		"password":      password,
		"database_name": requestBody.DatabaseName,
	}, startTime, "mongo-create-database", "Database Created")
}

func mongocopyDatabase(oldName, newName string, client *mongo.Client) error {
	oldDB := client.Database(oldName)
	newDB := client.Database(newName)

	collections, err := oldDB.ListCollectionNames(context.TODO(), bson.D{})
	if err != nil {
		return err
	}

	for _, collection := range collections {
		oldColl := oldDB.Collection(collection)
		newColl := newDB.Collection(collection)

		cursor, err := oldColl.Find(context.TODO(), bson.D{})
		if err != nil {
			return err
		}

		var documents []interface{}
		if err = cursor.All(context.TODO(), &documents); err != nil {
			return err
		}

		if len(documents) > 0 {
			_, err = newColl.InsertMany(context.TODO(), documents)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func mongoRenameDatabase(c *gin.Context, mongoURI string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		OldDatabaseName string `json:"old_database_name"`
		NewDatabaseName string `json:"new_database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "mongo-bind-json")
		return
	}

	client, ctx, err := connectToMongoDB(mongoURI)
	if err != nil {
		ErrorResponse(c, err, startTime, "mongo-connection-open")
		return
	}
	defer client.Disconnect(ctx)
	existingDatabases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		ErrorResponse(c, err, startTime, "mongo-list-databases")
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
		ErrorResponse(c, fmt.Errorf("database %s does not exist", requestBody.OldDatabaseName), startTime, "mongo-database-not-exist")
		return
	}

	for _, dbName := range existingDatabases {
		if dbName == requestBody.NewDatabaseName {
			ErrorResponse(c, fmt.Errorf("database %s already exists", requestBody.NewDatabaseName), startTime, "mongo-database-exists")
			return
		}
	}

	err = mongocopyDatabase(requestBody.OldDatabaseName, requestBody.NewDatabaseName, client)
	if err != nil {
		ErrorResponse(c, err, startTime, "mongo-copy-database")
		return
	}
	client.Database(requestBody.OldDatabaseName).Drop(context.TODO())
	SuccessResponse(c, map[string]interface{}{
		"old_database_name": requestBody.OldDatabaseName,
		"new_database_name": requestBody.NewDatabaseName,
	}, startTime, "mongo-rename-database", "Database Renamed")
}

func mongoDeleteDatabase(c *gin.Context, mongoURI string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "mongo-bind-json")
		return
	}

	client, ctx, err := connectToMongoDB(mongoURI)
	if err != nil {
		ErrorResponse(c, err, startTime, "mongo-connection-open")
		return
	}
	defer client.Disconnect(ctx)

	existingDatabases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		ErrorResponse(c, err, startTime, "mongo-list-databases")
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
		ErrorResponse(c, fmt.Errorf("database %s does not exist", requestBody.DatabaseName), startTime, "mongo-database-not-exist")
		return
	}

	if err := client.Database(requestBody.DatabaseName).Drop(ctx); err != nil {
		ErrorResponse(c, err, startTime, "mongo-drop-database")
		return
	}

	SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
	}, startTime, "mongo-delete-database", "Database Deleted")
}

func mongoViewDatabaseStats(c *gin.Context, mongoURI string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "mongo-bind-json")
		return
	}

	client, ctx, err := connectToMongoDB(mongoURI)
	if err != nil {
		ErrorResponse(c, err, startTime, "mongo-connection-open")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database(requestBody.DatabaseName)
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		ErrorResponse(c, err, startTime, "mongo-list-collections")
		return
	}

	var stats []bson.M
	for _, collectionName := range collections {
		collectionStats := bson.M{}
		if err := db.RunCommand(ctx, bson.D{{"collStats", collectionName}}).Decode(&collectionStats); err != nil {
			ErrorResponse(c, err, startTime, "mongo-collection-stats")
			return
		}
		stats = append(stats, collectionStats)
	}

	SuccessResponse(c, map[string]interface{}{
		"database_name": requestBody.DatabaseName,
		"stats":         stats,
	}, startTime, "mongo-view-database-stats", "Database Statistics Retrieved")
}

func mongoResetCredentials(c *gin.Context, mongoURI string) {
	startTime := time.Now().UnixMilli()
	var requestBody struct {
		DatabaseName string `json:"database_name"`
		Username     string `json:"username"`
	}
	if err := c.BindJSON(&requestBody); err != nil {
		ErrorResponse(c, err, startTime, "mongo-bind-json")
		return
	}

	client, ctx, err := connectToMongoDB(mongoURI)
	if err != nil {
		ErrorResponse(c, err, startTime, "mongo-connection-open")
		return
	}
	defer client.Disconnect(ctx)

	db := client.Database("admin")

	// Check if the user exists
	var result bson.M
	err = db.RunCommand(ctx, bson.D{
		{"usersInfo", bson.D{
			{"user", requestBody.Username},
			{"db", requestBody.DatabaseName},
		}},
	}).Decode(&result)
	if err != nil {
		ErrorResponse(c, fmt.Errorf("(UserNotFound) User '%s@%s' not found", requestBody.Username, requestBody.DatabaseName), startTime, "mongo-user-not-found")
		return
	}

	// Generate new password
	newPassword, _ := randomString(16)

	// Update the user's password
	if _, err := db.RunCommand(ctx, bson.D{
		{"updateUser", requestBody.Username},
		{"pwd", newPassword},
	}).DecodeBytes(); err != nil {
		ErrorResponse(c, err, startTime, "mongo-reset-credentials")
		return
	}

	SuccessResponse(c, map[string]interface{}{
		"username": requestBody.Username,
		"password": newPassword,
	}, startTime, "mongo-reset-credentials", "Credentials Reset")
}
