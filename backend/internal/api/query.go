package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func optionalQueryValue(c *gin.Context, key string) (string, bool) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return "", false
	}
	return value, true
}

func parseOptionalInt(c *gin.Context, key string) (*int, error) {
	value, ok := optionalQueryValue(c, key)
	if !ok {
		return nil, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalUint(c *gin.Context, key string) (*uint, error) {
	value, ok := optionalQueryValue(c, key)
	if !ok {
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
	value, ok := optionalQueryValue(c, key)
	if !ok {
		return nil
	}
	return &value
}

func parseOptionalBool(c *gin.Context, key string) (*bool, error) {
	value, ok := optionalQueryValue(c, key)
	if !ok {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseOptionalUnixTime(c *gin.Context, key string) (*time.Time, error) {
	value, ok := optionalQueryValue(c, key)
	if !ok {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	parsedTime := time.Unix(parsed, 0).UTC()
	return &parsedTime, nil
}
