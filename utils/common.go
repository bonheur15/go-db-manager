package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func SuccessResponse(c *gin.Context, data map[string]interface{}, startTime int64, action, message string) {
	log.Info().Str("action", action).Msg(message)
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
	log.Error().Err(err).Str("action", action).Msg(err.Error())
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
	lowerCharset = "abcdefghijklmnopqrstuvwxyz"
	mixedCharset = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func RandomString(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than 0")
	}

	result := make([]byte, length)
	num, err := rand.Int(rand.Reader, big.NewInt(int64(len(lowerCharset))))
	if err != nil {
		return "", err
	}
	result[0] = lowerCharset[num.Int64()]
	for i := 1; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(mixedCharset))))
		if err != nil {
			return "", err
		}
		result[i] = mixedCharset[num.Int64()]
	}

	return string(result), nil
}
