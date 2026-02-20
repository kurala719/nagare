package service

import "testing"

func TestSanitizeSensitiveText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "normal text",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "ip address",
			input: "Host 192.168.1.1 is down",
			want:  "Host x.x.x.x is down",
		},
		{
			name:  "password",
			input: "password=secret123",
			want:  "password=***",
		},
		{
			name:  "api key",
			input: "api-key: abcdef123456",
			want:  "api-key=***",
		},
		{
			name:  "bearer token",
			input: "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			want:  "Authorization: Bearer ***",
		},
		{
			name:  "basic auth url",
			input: "https://user:pass@example.com",
			want:  "https://user:***@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeSensitiveText(tt.input); got != tt.want {
				t.Errorf("sanitizeSensitiveText() = %v, want %v", got, tt.want)
			}
		})
	}
}
