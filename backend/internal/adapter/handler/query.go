package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func parseOptionalInt(c *gin.Context, key string) (*int, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalUint(c *gin.Context, key string) (*uint, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil, err
	}
	uParsed := uint(parsed)
	return &uParsed, nil
}

func parseOptionalString(c *gin.Context, key string) *string {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil
	}
	return &value
}

func parseOptionalBool(c *gin.Context, key string) (*bool, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalUnixTime(c *gin.Context, key string) (*time.Time, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	parsedTime := time.Unix(parsed, 0).UTC()
	return &parsedTime, nil
}
