package service

import (
	"testing"
)

func TestMergeAlertComment(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		analysis string
		want     string
	}{
		{
			name:     "empty existing",
			existing: "",
			analysis: "analysis result",
			want:     "analysis result",
		},
		{
			name:     "whitespace existing",
			existing: "   ",
			analysis: "analysis result",
			want:     "analysis result",
		},
		{
			name:     "existing content",
			existing: "original comment",
			analysis: "analysis result",
			want:     "original comment\n\nAI Analysis:\nanalysis result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeAlertComment(tt.existing, tt.analysis); got != tt.want {
				t.Errorf("mergeAlertComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlertAnalysisPrompt(t *testing.T) {
	prompt := alertAnalysisPrompt()
	if prompt == "" {
		t.Error("alertAnalysisPrompt() returned empty string")
	}
}
