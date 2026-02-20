package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/viper"
	"nagare/internal/model"
)

type rateLimitKey struct {
	key      string
	interval time.Duration
}

type mediaRateLimiter struct {
	mu   sync.Mutex
	last map[string]time.Time
}

var sendLimiter = &mediaRateLimiter{last: map[string]time.Time{}}

func allowMediaSend(media model.Media) (bool, time.Duration) {
	keys := mediaRateLimitKeys(media)
	if len(keys) == 0 {
		return true, 0
	}
	now := time.Now()
	sendLimiter.mu.Lock()
	defer sendLimiter.mu.Unlock()
	for _, item := range keys {
		if item.interval <= 0 {
			continue
		}
		if last, ok := sendLimiter.last[item.key]; ok {
			elapsed := now.Sub(last)
			if elapsed < item.interval {
				return false, item.interval - elapsed
			}
		}
	}
	for _, item := range keys {
		if item.interval > 0 {
			sendLimiter.last[item.key] = now
		}
	}
	return true, 0
}

func mediaRateLimitKeys(media model.Media) []rateLimitKey {
	globalInterval := time.Duration(viper.GetInt("media_rate_limit.global_interval_seconds")) * time.Second
	mediaTypeInterval := time.Duration(viper.GetInt("media_rate_limit.media_type_interval_seconds")) * time.Second
	mediaInterval := time.Duration(viper.GetInt("media_rate_limit.media_interval_seconds")) * time.Second

	keys := make([]rateLimitKey, 0, 3)
	if globalInterval > 0 {
		keys = append(keys, rateLimitKey{key: "global", interval: globalInterval})
	}
	if mediaTypeInterval > 0 && media.Type != "" {
		keys = append(keys, rateLimitKey{key: fmt.Sprintf("type:%s", media.Type), interval: mediaTypeInterval})
	}
	if mediaInterval > 0 && media.ID > 0 {
		keys = append(keys, rateLimitKey{key: fmt.Sprintf("media:%d", media.ID), interval: mediaInterval})
	}
	return keys
}
