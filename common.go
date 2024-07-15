package main

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, data map[string]interface{}, startTime int64, action, message string) {
	c.JSON(200, gin.H{
		"data":            data,
		"error":           false,
		"action":          action,
		"message":         message,
		"timestamp":       time.Now(),
		"action_duration": time.Now().UnixMilli() - startTime,
	})
}

func ErrorResponse(c *gin.Context, err error, startTime int64, action string) {
	c.JSON(400, gin.H{
		"error":           true,
		"message":         err.Error(),
		"action":          action,
		"timestamp":       time.Now(),
		"action_duration": time.Now().UnixMilli() - startTime,
		"data":            nil,
	})
}

const (
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func randomString(length int) (string, error) {
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
