package service

import (
	"context"
	"time"

	"github.com/spf13/viper"
)

const defaultAIAnalysisTimeoutSeconds = 30
const defaultAIAnalysisMinSeverity = 2

func aiAnalysisEnabled() bool {
	return viper.GetBool("ai.analysis_enabled")
}

func aiNotificationGuardEnabled() bool {
	return viper.GetBool("ai.notification_guard_enabled")
}

func aiAnalysisTimeout() time.Duration {
	seconds := viper.GetInt("ai.analysis_timeout_seconds")
	if seconds <= 0 {
		seconds = defaultAIAnalysisTimeoutSeconds
	}
	return time.Duration(seconds) * time.Second
}

func aiAnalysisContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), aiAnalysisTimeout())
}

func aiProviderConfig() (uint, string) {
	providerID := viper.GetUint("ai.provider_id")
	if providerID == 0 {
		providerID = 1
	}
	return providerID, viper.GetString("ai.model")
}

func aiAnalysisMinSeverity() int {
	minSeverity := viper.GetInt("ai.analysis_min_severity")
	if minSeverity <= 0 {
		minSeverity = defaultAIAnalysisMinSeverity
	}
	return minSeverity
}
