package service

import (
	"sync"

	"github.com/spf13/viper"
)

func configuredLimit(key string, fallback int) int {
	value := viper.GetInt(key)
	if value <= 0 {
		return fallback
	}
	return value
}

func runWithLimit(total, limit int, fn func(index int)) {
	if total <= 0 {
		return
	}
	if limit <= 0 {
		limit = 1
	}
	if limit > total {
		limit = total
	}

	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	for i := 0; i < total; i++ {
		sem <- struct{}{}
		wg.Add(1)
		idx := i
		go func() {
			defer wg.Done()
			defer func() {
				<-sem
			}()
			fn(idx)
		}()
	}
	wg.Wait()
}
